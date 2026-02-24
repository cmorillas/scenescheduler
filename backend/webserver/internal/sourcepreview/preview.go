// backend/webserver/internal/sourcepreview/preview.go
//
// Preview processing goroutine implementation.
//
// Contents:
// - processPreview - Main async goroutine for generating HLS previews

package sourcepreview

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// processPreview is the async goroutine that spawns hls-generator,
// waits for playlist creation, and invokes callbacks.
//
// This function runs in its own goroutine and manages the entire lifecycle
// of generating an HLS preview stream:
//  1. Spawn hls-generator process
//  2. Poll for playlist.m3u8 creation
//  3. Invoke OnReady callback when ready, or OnError on failure
//
// The function includes panic recovery to ensure failures don't crash the server.
func (m *Manager) processPreview(session *Session) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Preview processing panic",
				"previewID", session.PreviewID,
				"remoteAddr", session.RemoteAddr,
				"panic", r)
			if session.onError != nil {
				session.onError(fmt.Sprintf("internal error: %v", r))
			}
			m.removeSession(session.PreviewID)
		}
	}()

	// 1. Spawn hls-generator process
	m.logger.Debug("Spawning hls-generator process",
		"previewID", session.PreviewID,
		"remoteAddr", session.RemoteAddr)

	process, err := m.spawnProcess(session.SourceURI, session.TempDir)
	if err != nil {
		m.logger.Error("Failed to spawn process",
			"error", err,
			"previewID", session.PreviewID,
			"remoteAddr", session.RemoteAddr)

		if session.onError != nil {
			stderr := ""
			if process != nil && process.StderrBuf != nil {
				stderr = process.StderrBuf.String()
			}
			session.onError(fmt.Sprintf("failed to spawn process: %v\nStderr: %s", err, stderr))
		}
		m.removeSession(session.PreviewID)
		return
	}

	// 2. Store process handle in session
	m.mu.Lock()
	if s, ok := m.activePreviews[session.PreviewID]; ok {
		s.Process = process
	}
	m.mu.Unlock()

	m.logger.Debug("Process spawned",
		"previewID", session.PreviewID,
		"pid", process.PID,
		"remoteAddr", session.RemoteAddr)

	// 3. Wait for playlist.m3u8 to be created
	playlistPath := filepath.Join(session.TempDir, "playlist.m3u8")

	ctx, cancel := context.WithTimeout(context.Background(), playlistWaitTimeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if playlist exists
			if _, err := os.Stat(playlistPath); err == nil {
				// Playlist file exists - verify it has at least one segment
				hasSegments, err := playlistHasSegments(playlistPath)
				if err != nil {
					m.logger.Debug("Error reading playlist",
						"previewID", session.PreviewID,
						"error", err)
					continue // Keep polling
				}

				if !hasSegments {
					m.logger.Debug("Playlist exists but has no segments yet, waiting...",
						"previewID", session.PreviewID)
					continue // Keep polling
				}

				// SUCCESS! Playlist has segments
				hlsURL := fmt.Sprintf("/hls/preview-%d/playlist.m3u8", session.PreviewID)

				m.logger.Info("Preview ready",
					"previewID", session.PreviewID,
					"remoteAddr", session.RemoteAddr,
					"hlsURL", hlsURL)

				if session.onReady != nil {
					session.onReady(hlsURL)
				}

				// Schedule automatic timeout to prevent resource accumulation
				timeoutTimer := time.AfterFunc(previewMaxRuntime, func() {
					m.logger.Info("Preview auto-stopped after maximum runtime",
						"previewID", session.PreviewID,
						"connectionID", session.ConnectionID,
						"runtime", previewMaxRuntime)

					// Notify frontend BEFORE cleanup (so it can gracefully stop HLS.js)
					if session.onStopped != nil {
						session.onStopped("Preview automatically stopped after 30 seconds")
					}

					// Small delay to ensure WebSocket message is sent before cleanup
					time.Sleep(100 * time.Millisecond)

					if err := m.StopPreview(session.ConnectionID); err != nil {
						m.logger.Error("Failed to stop preview on timeout",
							"previewID", session.PreviewID,
							"error", err)
					}
				})

				// Store timer in session so it can be canceled if manually stopped
				m.mu.Lock()
				if s, ok := m.activePreviews[session.PreviewID]; ok {
					s.TimeoutTimer = timeoutTimer
				}
				m.mu.Unlock()

				return
			}

		case <-ctx.Done():
			// TIMEOUT - playlist not created in time
			stderr := ""
			if process.StderrBuf != nil {
				stderr = process.StderrBuf.String()
			}

			errMsg := fmt.Sprintf("preview generation timed out after %v. Playlist file was not created.\nStderr: %s",
				playlistWaitTimeout, stderr)

			m.logger.Error("Preview timeout",
				"previewID", session.PreviewID,
				"remoteAddr", session.RemoteAddr,
				"timeout", playlistWaitTimeout)

			if session.onError != nil {
				session.onError(errMsg)
			}

			// Cleanup failed preview
			m.StopPreview(session.ConnectionID)
			return
		}
	}
}

// playlistHasSegments checks if an HLS playlist contains at least one segment entry.
// It reads the playlist file and looks for the #EXTINF: tag which indicates a segment.
//
// Returns:
//   - true if the playlist contains at least one #EXTINF: entry
//   - false if the playlist is empty or contains only header tags
//   - error if the file cannot be read
func playlistHasSegments(playlistPath string) (bool, error) {
	content, err := os.ReadFile(playlistPath)
	if err != nil {
		return false, err
	}

	// Check for #EXTINF tag which indicates a segment entry
	// Example line: #EXTINF:4.000000,
	return bytes.Contains(content, []byte("#EXTINF:")), nil
}
