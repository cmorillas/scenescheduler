// backend/webserver/internal/sourcepreview/process.go
//
// Process management utilities for hls-generator.
//
// DESIGN NOTE: Binary Coupling
// =============================
// This file is intentionally coupled to the "hls-generator" binary (external C++ tool).
// No abstraction layer exists because we follow the YAGNI principle - we don't need
// multiple generator implementations yet.
//
// If migrating to FFmpeg in the future, modify these sections:
//
// 1. findHLSGeneratorBinary() - Change binary name from "hls-generator" to "ffmpeg"
// 2. spawnProcess() - Replace CLI invocation:
//    Current:  hls-generator <sourceURI> <outputDir>
//    FFmpeg:   ffmpeg -i <sourceURI> -c:v libx264 -hls_time 4 -hls_list_size 5 \
//              -hls_flags delete_segments -hls_segment_filename <outputDir>/segment_%03d.ts \
//              <outputDir>/playlist.m3u8
// 3. Log messages referencing "hls-generator"
// 4. Verify playlist.m3u8 generation behavior matches (see preview.go:78)
//
// Expected CLI Interface:
//   - Takes source URI (RTMP, file path, device) as input
//   - Generates playlist.m3u8 in output directory
//   - Responds to SIGTERM for graceful shutdown
//   - Writes diagnostic info to stderr
//
// Contents:
// - findHLSGeneratorBinary - Binary discovery
// - spawnProcess - Process spawning
// - killProcess - Graceful process termination

package sourcepreview

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// findHLSGeneratorBinary discovers the hls-generator binary location.
// It searches in the following order:
//  1. Same directory as the scenescheduler executable
//  2. System PATH
//
// Returns the absolute path to the binary, or error if not found.
func findHLSGeneratorBinary() (string, error) {
	// 1. Same directory as executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		candidate := filepath.Join(exeDir, "hls-generator")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// 2. System PATH
	path, err := exec.LookPath("hls-generator")
	if err == nil {
		return path, nil
	}

	return "", ErrBinaryNotFound
}

// spawnProcess starts an hls-generator process with stderr capture.
//
// Parameters:
//   - sourceURI: The source URI to process (RTMP, file, etc.)
//   - outputDir: Directory where HLS files will be written
//
// Returns:
//   - *ProcessHandle: Handle to the spawned process with stderr buffer
//   - error: If process failed to start
func (m *Manager) spawnProcess(sourceURI, outputDir string) (*ProcessHandle, error) {
	// Create circular buffer for stderr
	stderrBuf := &bytes.Buffer{}
	limitedWriter := &limitedWriter{buf: stderrBuf, maxSize: stderrBufferSize}

	cmd := exec.Command(m.hlsGeneratorPath, sourceURI, outputDir)
	cmd.Stderr = limitedWriter

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	return &ProcessHandle{
		Cmd:       cmd,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now(),
		StderrBuf: stderrBuf,
	}, nil
}

// killProcess terminates a process gracefully (SIGTERM) with fallback to SIGKILL.
//
// Process termination sequence:
//  1. Send SIGTERM (graceful shutdown signal)
//  2. Wait up to processKillTimeout (5 seconds)
//  3. If still running, send SIGKILL (forced termination)
func (m *Manager) killProcess(handle *ProcessHandle) {
	if handle == nil || handle.Cmd == nil || handle.Cmd.Process == nil {
		return
	}

	m.logger.Debug("Killing process", "pid", handle.PID)

	// Try graceful SIGTERM first
	if err := handle.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		m.logger.Warn("Failed to send SIGTERM", "pid", handle.PID, "error", err)
	}

	// Wait with timeout
	done := make(chan error, 1)
	go func() {
		done <- handle.Cmd.Wait()
	}()

	select {
	case <-time.After(processKillTimeout):
		// Force kill with SIGKILL
		m.logger.Warn("Process did not terminate gracefully, sending SIGKILL", "pid", handle.PID)
		_ = handle.Cmd.Process.Kill()
		<-done // Wait for final cleanup

	case <-done:
		// Terminated gracefully
		m.logger.Debug("Process terminated gracefully", "pid", handle.PID)
	}
}

// -----------------------------------------------------------------------------
// limitedWriter - Circular Buffer for Stderr Capture
// -----------------------------------------------------------------------------

// limitedWriter implements a circular buffer that keeps only the last N bytes.
// Used to capture stderr output without unbounded memory growth.
type limitedWriter struct {
	buf     *bytes.Buffer
	maxSize int
}

// Write implements io.Writer interface with circular buffer behavior.
// If adding would exceed maxSize, oldest data is discarded.
func (w *limitedWriter) Write(p []byte) (n int, err error) {
	// If adding would exceed max, remove oldest data
	if w.buf.Len()+len(p) > w.maxSize {
		excess := (w.buf.Len() + len(p)) - w.maxSize
		w.buf.Next(excess) // Drop oldest bytes
	}
	return w.buf.Write(p)
}
