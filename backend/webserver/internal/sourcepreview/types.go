// backend/webserver/internal/sourcepreview/types.go
//
// Type definitions for the source preview module.
//
// Contents:
// - Constants (timeouts, buffer sizes)
// - Error definitions
// - Manager struct
// - Session struct
// - ProcessHandle struct
// - StartPreviewRequest struct

package sourcepreview

import (
	"bytes"
	"errors"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"scenescheduler/backend/logger"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// playlistWaitTimeout is the maximum time to wait for playlist.m3u8 creation
	playlistWaitTimeout = 30 * time.Second

	// processKillTimeout is the grace period for SIGTERM before SIGKILL
	processKillTimeout = 5 * time.Second

	// previewMaxRuntime is the maximum time a preview can run before auto-stop
	previewMaxRuntime = 30 * time.Second

	// pollInterval is the interval for checking playlist file existence
	pollInterval = 500 * time.Millisecond

	// stderrBufferSize is the maximum size of stderr circular buffer
	stderrBufferSize = 1024 // 1KB
)

// =============================================================================
// Errors
// =============================================================================

var (
	// ErrBinaryNotFound indicates hls-generator binary was not found
	ErrBinaryNotFound = errors.New("hls-generator binary not found")

	// ErrSessionNotFound indicates requested session does not exist
	ErrSessionNotFound = errors.New("preview session not found")
)

// =============================================================================
// Manager
// =============================================================================

// Manager orchestrates HLS preview generation for program sources.
// It manages temporary hls-generator processes and filesystem resources.
// Each WebSocket client can have at most one active preview.
type Manager struct {
	// --- Dependencies (immutable) ---
	logger      *logger.Logger
	hlsBasePath string // Base directory for HLS files (e.g., "./hls")

	// --- Binary Discovery ---
	hlsGeneratorPath string // Cached path to hls-generator binary

	// --- Active Sessions (guarded by mu) ---
	mu              sync.RWMutex
	activePreviews  map[uint64]*Session // Key: previewID (incremental)
	connIDToPreview map[string]uint64   // Key: connectionID → previewID (for lookup)
	nextPreviewID   atomic.Uint64       // Incremental counter (thread-safe)

	// --- Shutdown Coordination ---
	shutdownOnce sync.Once
}

// =============================================================================
// Session
// =============================================================================

// Session represents an active HLS preview generation session.
type Session struct {
	PreviewID    uint64 // Incremental ID (1, 2, 3, ...)
	ConnectionID string // WebSocket connection ID (unique identifier)
	RemoteAddr   string // WebSocket remote address (for logging/debugging)
	SourceURI    string // Source URI to preview
	InputKind    string // OBS input kind (ffmpeg_source, etc)
	TempDir      string // Filesystem path: {hlsBase}/preview-{previewID}/
	Process      *ProcessHandle
	CreatedAt    time.Time
	TimeoutTimer *time.Timer // Auto-stop timer (canceled on manual stop)

	// Async callbacks for result notification
	onReady   func(hlsURL string)
	onError   func(errorMsg string)
	onStopped func(reason string) // Called when preview is auto-stopped
}

// =============================================================================
// ProcessHandle
// =============================================================================

// ProcessHandle holds resources for an active hls-generator process.
type ProcessHandle struct {
	Cmd       *exec.Cmd
	PID       int
	StartedAt time.Time
	StderrBuf *bytes.Buffer // Circular buffer (max 1KB) for error diagnostics
}

// =============================================================================
// StartPreviewRequest
// =============================================================================

// StartPreviewRequest contains all parameters for a preview request.
type StartPreviewRequest struct {
	ConnectionID  string      // WebSocket connection ID (unique identifier)
	RemoteAddr    string      // WebSocket remote address (IP:port, for logging)
	SourceURI     string      // Source URI to preview
	InputKind     string      // OBS input kind
	InputSettings interface{} // OBS input settings (optional)

	// Async callbacks (called from goroutine)
	OnReady   func(hlsURL string)   // Called when HLS stream is ready
	OnError   func(errorMsg string) // Called on any error
	OnStopped func(reason string)   // Called when preview is auto-stopped (timeout, crash, etc)
}
