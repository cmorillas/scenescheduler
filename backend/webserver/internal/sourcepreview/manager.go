// backend/webserver/internal/sourcepreview/manager.go
//
// Manager implementation for the source preview module.
//
// Contents:
// - Constructor (New)
// - StartPreview - Main API to start a preview
// - StopPreview - Stop an active preview
// - Shutdown - Graceful cleanup

package sourcepreview

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"scenescheduler/backend/logger"
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new preview manager instance.
// It discovers the hls-generator binary and performs startup cleanup.
//
// Parameters:
//   - logger: Logger instance (will be scoped to "sourcepreview" module)
//   - hlsBasePath: Base directory for HLS files (e.g., "./hls")
//
// Returns:
//   - *Manager: Configured manager ready to handle requests
//   - error: Only if cleanup fails (fatal)
//
// Non-fatal issues (binary not found) are logged but don't prevent creation.
// Preview requests will fail gracefully with clear error messages.
func New(logger *logger.Logger, hlsBasePath string) (*Manager, error) {
	logger = logger.WithModule("sourcepreview")

	// Discover hls-generator binary
	binaryPath, err := findHLSGeneratorBinary()
	if err != nil {
		logger.Error("hls-generator binary not found - preview requests will fail",
			"error", err,
			"suggestion", "Place hls-generator in same directory as scenescheduler executable")
		// Don't fail constructor - allow WebServer to start
		binaryPath = "" // Will cause preview requests to fail gracefully
	} else {
		logger.Info("Found hls-generator binary", "path", binaryPath)
	}

	// Startup cleanup: remove all old preview directories
	// This handles leftover directories from previous runs
	if err := os.RemoveAll(hlsBasePath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to clean HLS base directory on startup", "error", err)
	}

	// Recreate empty HLS base directory
	if err := os.MkdirAll(hlsBasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create HLS base directory: %w", err)
	}

	logger.Debug("HLS base directory ready", "path", hlsBasePath)

	return &Manager{
		logger:           logger,
		hlsBasePath:      hlsBasePath,
		hlsGeneratorPath: binaryPath,
		activePreviews:   make(map[uint64]*Session),
		connIDToPreview:  make(map[string]uint64),
		// nextPreviewID starts at 0, first Add(1) will return 1
	}, nil
}

// =============================================================================
// Public API
// =============================================================================

// StartPreview initiates HLS preview generation for a source.
// This is a non-blocking operation - result is delivered via callbacks.
//
// Behavior:
//   - Validates request
//   - Auto-cancels previous preview if client already has one active
//   - Spawns hls-generator process asynchronously
//   - Polls for playlist.m3u8 creation (30s timeout)
//   - Calls OnReady(hlsURL) on success or OnError(msg) on failure
//
// Returns:
//   - error: Immediate validation errors only (binary not found)
//   - nil: Request accepted, result will be delivered via callbacks
func (m *Manager) StartPreview(req StartPreviewRequest) error {
	// Validation: check binary availability
	if m.hlsGeneratorPath == "" {
		errMsg := "hls-generator binary not found. Please ensure hls-generator is installed."
		m.logger.Error("Cannot start preview - binary not available", "remoteAddr", req.RemoteAddr)
		if req.OnError != nil {
			req.OnError(errMsg)
		}
		return ErrBinaryNotFound
	}

	// SAFETY NET: Auto-cancel previous preview if client already has one
	m.mu.Lock()
	if oldPreviewID, exists := m.connIDToPreview[req.ConnectionID]; exists {
		if oldSession, ok := m.activePreviews[oldPreviewID]; ok {
			m.logger.Warn("Client requested new preview without stopping previous - auto-cancelling",
				"connectionID", req.ConnectionID,
				"remoteAddr", req.RemoteAddr,
				"oldPreviewID", oldPreviewID,
				"previousSourceURI", oldSession.SourceURI,
				"newSourceURI", req.SourceURI)
		}

		// Stop old preview WITHOUT holding lock (avoid deadlock)
		m.mu.Unlock()
		m.stopPreviewInternal(oldPreviewID)
		m.mu.Lock()
	}
	m.mu.Unlock()

	// Generate incremental preview ID (atomic, thread-safe)
	previewID := m.nextPreviewID.Add(1)

	// Create preview directory: {hlsBase}/preview-{id}/
	tempDir := filepath.Join(m.hlsBasePath, fmt.Sprintf("preview-%d", previewID))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create preview directory: %v", err)
		m.logger.Error("Failed to create preview directory",
			"error", err,
			"previewID", previewID,
			"remoteAddr", req.RemoteAddr)
		if req.OnError != nil {
			req.OnError(errMsg)
		}
		return err
	}

	m.logger.Debug("Preview directory created", "path", tempDir, "previewID", previewID)

	// Create session
	session := &Session{
		PreviewID:    previewID,
		ConnectionID: req.ConnectionID,
		RemoteAddr:   req.RemoteAddr,
		SourceURI:    req.SourceURI,
		InputKind:    req.InputKind,
		TempDir:      tempDir,
		CreatedAt:    time.Now(),
		onReady:      req.OnReady,
		onError:      req.OnError,
		onStopped:    req.OnStopped,
	}

	// Store session
	m.addSession(session)

	m.logger.Info("Starting preview",
		"previewID", previewID,
		"connectionID", req.ConnectionID,
		"remoteAddr", req.RemoteAddr,
		"sourceURI", req.SourceURI)

	// Spawn process asynchronously
	go m.processPreview(session)

	return nil
}

// StopPreview terminates an active preview session for a client.
// It kills the process, cleans up filesystem, and removes session tracking.
//
// This operation is idempotent - calling on non-existent session is a no-op.
// This is automatically called when WebSocket disconnects.
//
// Parameters:
//   - connectionID: WebSocket connection ID (unique identifier)
//
// Returns:
//   - error: Only if session cleanup fails (logged, not fatal)
func (m *Manager) StopPreview(connectionID string) error {
	m.mu.Lock()
	previewID, exists := m.connIDToPreview[connectionID]
	if !exists {
		m.mu.Unlock()
		return nil // Idempotent - stopping non-existent session is no-op
	}

	session, exists := m.activePreviews[previewID]
	if !exists {
		m.mu.Unlock()
		return nil
	}

	// Remove from maps immediately (before killing process)
	delete(m.activePreviews, previewID)
	delete(m.connIDToPreview, connectionID)
	m.mu.Unlock()

	m.logger.Debug("Stopping preview", "previewID", previewID, "connectionID", connectionID, "remoteAddr", session.RemoteAddr)

	// Kill process (outside lock)
	if session.Process != nil {
		m.killProcess(session.Process)
	}

	// Cleanup filesystem
	if err := os.RemoveAll(session.TempDir); err != nil {
		m.logger.Warn("Failed to remove temp directory", "path", session.TempDir, "error", err)
		return err
	}

	m.logger.Debug("Preview stopped and cleaned up", "previewID", previewID)
	return nil
}

// Shutdown performs graceful cleanup of all resources.
// Called by WebServer during shutdown sequence.
//
// It kills all active processes in parallel using WaitGroup,
// removes all preview directories, and clears internal state.
// This operation is idempotent.
//
// Returns:
//   - error: Only if filesystem cleanup fails (logged, not fatal)
func (m *Manager) Shutdown() error {
	var shutdownErr error

	m.shutdownOnce.Do(func() {
		// Get all active preview IDs (copy to avoid holding lock during cleanup)
		m.mu.RLock()
		previewIDs := make([]uint64, 0, len(m.activePreviews))
		for id := range m.activePreviews {
			previewIDs = append(previewIDs, id)
		}
		m.mu.RUnlock()

		if len(previewIDs) == 0 {
			m.logger.Info("Preview manager shutdown: no active previews")
		} else {
			m.logger.Info("Shutting down preview manager", "activePreviews", len(previewIDs))

			// Kill all processes in parallel
			var wg sync.WaitGroup
			for _, id := range previewIDs {
				wg.Add(1)
				go func(previewID uint64) {
					defer wg.Done()
					m.stopPreviewInternal(previewID)
				}(id)
			}
			wg.Wait()

			m.logger.Debug("All preview processes terminated", "count", len(previewIDs))
		}

		// Cleanup all preview directories
		if err := os.RemoveAll(m.hlsBasePath); err != nil {
			m.logger.Error("Failed to remove HLS base directory", "error", err)
			shutdownErr = err
		}

		// Clear maps
		m.mu.Lock()
		m.activePreviews = nil
		m.connIDToPreview = nil
		m.mu.Unlock()

		m.logger.Info("Preview manager shutdown complete")
	})

	return shutdownErr
}

// =============================================================================
// Internal Helpers
// =============================================================================

// stopPreviewInternal is internal version that operates on PreviewID directly.
// Used when already holding lock or during auto-cancel scenarios.
func (m *Manager) stopPreviewInternal(previewID uint64) {
	m.mu.RLock()
	session, exists := m.activePreviews[previewID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	// Cancel timeout timer if it exists (prevents double cleanup)
	if session.TimeoutTimer != nil {
		session.TimeoutTimer.Stop()
	}

	m.mu.Lock()
	delete(m.activePreviews, previewID)
	if session.ConnectionID != "" {
		delete(m.connIDToPreview, session.ConnectionID)
	}
	m.mu.Unlock()

	if session.Process != nil {
		m.killProcess(session.Process)
	}

	os.RemoveAll(session.TempDir)
}
