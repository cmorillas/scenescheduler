// backend/mediasource/internal/feed/monitor.go
//
// This file implements robust feed and track monitoring with health checks.
// It follows a progressive degradation policy: the feed continues operating
// with partial tracks until all configured tracks fail.
//
// Order of sections:
// 1) Public Monitoring API
// 2) Monitoring Orchestration
// 3) Per-Track Monitoring
// 4) Health Check Implementation
// 5) Failure Detection and Reporting

package feed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/pion/mediadevices"
)

// Health check tuning parameters
const (
	healthCheckInterval = 5 * time.Second
	healthCheckTimeout  = 2 * time.Second
	frozenThreshold     = 30 * time.Second
)

// ============================================================================
// 1) PUBLIC MONITORING API
// ============================================================================

// StartMonitoring launches health monitoring for all acquired tracks.
// It spawns goroutines that perform periodic health checks and detect failures.
//
// The monitoring goroutines will exit when the feed's internal context is
// cancelled or when total failure occurs.
func (f *Feed) StartMonitoring(onFailure func(string)) {
	if f.ctx == nil {
		return // Feed not acquired
	}

	f.logger.Debug("Starting feed monitoring")

	// Use feed's internal context for the lifetime of all monitoring tasks.
	monitorCtx := f.ctx

	// Track monitor lifecycle with WaitGroup.
	failureChan := make(chan string, 2)

	f.mu.RLock()
	hasVideo := f.videoTrack != nil
	hasAudio := f.audioTrack != nil
	f.mu.RUnlock()

	// Determine how many monitors will be started.
	numMonitors := 0
	if hasVideo {
		numMonitors++
	}
	if hasAudio {
		numMonitors++
	}
	f.monitorWg.Add(numMonitors)

	if hasVideo {
		go f.monitorTrack(monitorCtx, "video", failureChan)
	}
	if hasAudio {
		go f.monitorTrack(monitorCtx, "audio", failureChan)
	}

	// Launch failure aggregator if there's anything to monitor.
	if numMonitors > 0 {
		go f.handleFailures(monitorCtx, failureChan, onFailure)
	}
}

// ============================================================================
// 2) MONITORING ORCHESTRATION
// ============================================================================

// handleFailures aggregates track failures and applies the degradation policy.
// It only triggers total failure (onFailure callback) when all configured
// tracks have failed.
func (f *Feed) handleFailures(ctx context.Context, failureChan <-chan string, onFailure func(string)) {
	for {
		select {
		case <-ctx.Done():
			f.logger.Debug("Failure handler exiting due to context cancellation.")
			return
		case reason := <-failureChan:
			f.logger.Warn("Track failure detected", "reason", reason)
			if f.checkTotalFailure() {
				f.logger.Error("Total feed failure: all configured tracks have failed.")
				if onFailure != nil {
					onFailure(reason)
				}
				// Cancel the feed's main context to stop everything.
				if f.cancel != nil {
					f.cancel()
				}
				return
			}
			f.logger.Debug("Continuing with degraded service (partial track failure).")
		}
	}
}

// checkTotalFailure returns true if all configured tracks have failed.
func (f *Feed) checkTotalFailure() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	videoConfigured := f.videoDevice != nil
	audioConfigured := f.audioDevice != nil

	videoHasFailed := f.videoFailed
	audioHasFailed := f.audioFailed

	if videoConfigured && audioConfigured {
		return videoHasFailed && audioHasFailed
	}
	if videoConfigured {
		return videoHasFailed
	}
	if audioConfigured {
		return audioHasFailed
	}

	// If nothing was configured, there's no failure.
	return false
}

// ============================================================================
// 3) PER-TRACK MONITORING
// ============================================================================

// monitorTrack monitors a single track with periodic health checks.
// It detects both abrupt failures (OnEnded callback) and gradual degradation
// (frozen device detected via health checks).
func (f *Feed) monitorTrack(ctx context.Context, trackType string, failureChan chan<- string) {
	defer f.monitorWg.Done()

	var track mediadevices.Track
	f.mu.RLock()
	if trackType == "video" {
		track = f.videoTrack
	} else {
		track = f.audioTrack
	}
	f.mu.RUnlock()

	if track == nil {
		return
	}

	f.logger.Debug("Starting track monitor", "trackType", trackType, "trackID", track.ID())
	defer f.logger.Debug("Stopping track monitor", "trackType", trackType)

	// Register OnEnded callback for abrupt failures.
	// This callback uses the 'ctx' variable captured from this function's scope,
	// which is safe from the race condition with f.ctx being nilled on cleanup.
	track.OnEnded(func(err error) {
		select {
		case <-ctx.Done():
			// This is an expected shutdown, not a failure.
			f.logger.Debug("Track OnEnded ignored: feed context is already cancelled", "trackType", trackType)
			return
		default:
			// The feed is active, so this is an unexpected failure.
			f.logger.Warn("Track OnEnded fired unexpectedly", "trackType", trackType, "error", err)
			if err != nil && err != io.EOF {
				f.markTrackFailed(trackType)
				// Non-blocking send to the failure channel.
				select {
				case failureChan <- fmt.Sprintf("%s track ended: %v", trackType, err):
				default:
				}
			}
		}
	})

	// Periodic health check loop.
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()
	var checkInProgress atomic.Bool

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !checkInProgress.CompareAndSwap(false, true) {
				f.logger.Debug("Skipping health check, previous still running", "trackType", trackType)
				continue
			}
			checkCtx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
			healthErr := f.checkTrackHealth(checkCtx, track)
			cancel()
			checkInProgress.Store(false)

			now := time.Now()
			if healthErr != nil {
				lastRead := f.getLastRead(trackType)
				if !lastRead.IsZero() {
					since := now.Sub(lastRead)
					if since > frozenThreshold {
						f.markTrackFailed(trackType)
						select {
						case failureChan <- fmt.Sprintf("%s device frozen: no data for %v", trackType, since):
						default:
						}
						return // Stop monitoring this failed track.
					}
				}
			} else {
				f.updateLastRead(trackType, now)
			}
		}
	}
}

// ============================================================================
// 4) HEALTH CHECK IMPLEMENTATION
// ============================================================================

// checkTrackHealth performs a real health check by attempting to read data
// from the track. Returns error if the read fails or times out.
func (f *Feed) checkTrackHealth(ctx context.Context, track mediadevices.Track) error {
	if track == nil {
		return errors.New("track is nil")
	}

	result := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("panic during health check: %v", r)
			}
		}()

		switch t := track.(type) {
		case *mediadevices.VideoTrack:
			reader := t.NewReader(false)
			_, release, err := reader.Read()
			if release != nil {
				release()
			}
			result <- err
		case *mediadevices.AudioTrack:
			reader := t.NewReader(false)
			_, release, err := reader.Read()
			if release != nil {
				release()
			}
			result <- err
		default:
			result <- fmt.Errorf("unknown track type: %T", track)
		}
	}()

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("health check timed out, device may be frozen")
	}
}

// ============================================================================
// 5) FAILURE DETECTION AND REPORTING
// ============================================================================

// markTrackFailed marks a track as failed in the feed state.
func (f *Feed) markTrackFailed(trackType string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if trackType == "video" {
		f.videoFailed = true
	} else if trackType == "audio" {
		f.audioFailed = true
	}
}

// updateLastRead updates the last successful read timestamp for a track.
func (f *Feed) updateLastRead(trackType string, t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if trackType == "video" {
		f.lastVideoRead = t
	} else if trackType == "audio" {
		f.lastAudioRead = t
	}
}

// getLastRead retrieves the last successful read timestamp for a track.
func (f *Feed) getLastRead(trackType string) time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if trackType == "video" {
		return f.lastVideoRead
	}
	return f.lastAudioRead
}
