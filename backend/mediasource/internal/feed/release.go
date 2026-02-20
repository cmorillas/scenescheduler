// backend/mediasource/internal/feed/release.go
//
// This file contains resource release and cleanup logic for the Feed.
// It ensures graceful shutdown of monitoring goroutines and proper cleanup
// of media resources.
//
// Order of sections:
// 1) Public Release API
// 2) Cleanup Orchestration
// 3) Field Reset

package feed

import (
	"context"
	"time"
)

// ============================================================================
// 1) PUBLIC RELEASE API
// ============================================================================

// Release performs graceful shutdown and cleanup of all feed resources.
// It cancels monitoring goroutines, waits for them to exit, closes tracks
// and streams, and resets internal state.
//
// The timeout parameter controls how long to wait for monitoring goroutines
// to exit before proceeding with cleanup anyway.
//
// Preconditions: Can be called at any time, even if feed was never acquired.
// Side effects: Cancels f.ctx, closes all tracks/streams, resets fields.
// Returns: error if cleanup encounters issues (currently always returns nil).
func (f *Feed) Release(timeout time.Duration) error {
	// Create a context for the release operation with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	f.releaseResources(ctx)
	return nil
}

// ============================================================================
// 2) CLEANUP ORCHESTRATION
// ============================================================================

// releaseResources performs the full teardown sequence.
// Steps:
// 1) Cancel the feed context (stops monitoring goroutines)
// 2) **IMMEDIATELY close tracks to release device locks**
// 3) Wait for all monitors to exit (with timeout)
// 4) Reset internal fields
func (f *Feed) releaseResources(ctx context.Context) {
	// Step 1: Cancel monitoring and any pending work
	if f.cancel != nil {
		f.cancel()
	}

	// Step 2: CRITICAL - Close tracks IMMEDIATELY to release device
	// This must happen BEFORE waiting for monitors, otherwise the device
	// stays locked and prevents re-acquisition
	f.closeTracks()

	// Step 3: Wait for all monitor goroutines to exit
	// They will detect context cancellation and exit
	done := make(chan struct{})
	go func() {
		f.monitorWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All monitors exited gracefully
	case <-ctx.Done():
		// Timeout waiting for monitors - already closed tracks anyway
	}

	// Step 4: Reset fields
	f.resetFields()
}

// ============================================================================
// 3) FIELD RESET
// ============================================================================

// resetFields zeroes internal references to avoid stale usage.
// We keep acquiredAt for post-mortem diagnostics.
func (f *Feed) resetFields() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ctx = nil
	f.cancel = nil
	f.videoDevice = nil
	f.audioDevice = nil
	f.videoFailed = false
	f.audioFailed = false
	f.lastVideoRead = time.Time{}
	f.lastAudioRead = time.Time{}
	
	// Keep acquiredAt for diagnostics
	// Uncomment to reset: f.acquiredAt = time.Time{}
}