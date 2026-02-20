// backend/mediasource/internal/feed/feed.go
//
// This file defines the Feed component: a self-contained media acquisition
// and monitoring unit. It exposes a minimal API to the mediasource.Manager
// and manages its own lifecycle, health checks, and resource cleanup.
//
// Order of sections:
// 1) Feed Struct Definition
// 2) Constructor
// 3) Public Getters (thread-safe)
// 4) Metadata Getters

package feed

import (
	"context"
	"sync"
	"time"

	"github.com/pion/mediadevices"
	"scenescheduler/backend/config"
	"scenescheduler/backend/logger"
)

// ============================================================================
// 1) FEED STRUCT DEFINITION
// ============================================================================

// Feed encapsulates a single media acquisition session with lifecycle management.
// It owns the media stream, tracks, device metadata, and monitoring goroutines.
// All public methods are thread-safe.
type Feed struct {
	// --- Dependencies (immutable) ---
	cfg           config.MediaSourceConfig
	codecSelector *mediadevices.CodecSelector
	logger        *logger.Logger

	// --- Media Resources (guarded by mu) ---
	mu          sync.RWMutex
	mediaStream mediadevices.MediaStream
	videoTrack  *mediadevices.VideoTrack
	audioTrack  *mediadevices.AudioTrack
	videoDevice *mediadevices.MediaDeviceInfo
	audioDevice *mediadevices.MediaDeviceInfo
	videoFailed bool
	audioFailed bool
	acquiredAt  time.Time
	lastVideoRead time.Time
	lastAudioRead time.Time

	// --- Lifecycle Management ---
	ctx       context.Context
	cancel    context.CancelFunc
	monitorWg sync.WaitGroup
}

// ============================================================================
// 2) CONSTRUCTOR
// ============================================================================

// New creates a new Feed instance ready for acquisition.
// The Feed is not active until Acquire() is called successfully.
func New(cfg config.MediaSourceConfig, codecSelector *mediadevices.CodecSelector, log *logger.Logger) *Feed {
	return &Feed{
		cfg:           cfg,
		codecSelector: codecSelector,
		logger:        log.WithModule("feed"),
	}
}

// ============================================================================
// 3) PUBLIC GETTERS (THREAD-SAFE)
// ============================================================================

// GetVideoTrack returns the active video track, or nil if unavailable or failed.
// With progressive degradation policy, this returns nil if video failed but
// the feed continues with audio only.
// The caller MUST NOT close or modify the returned track.
func (f *Feed) GetVideoTrack() *mediadevices.VideoTrack {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.videoTrack == nil || f.videoFailed {
		return nil
	}
	return f.videoTrack
}

// GetAudioTrack returns the active audio track, or nil if unavailable or failed.
// With progressive degradation policy, this returns nil if audio failed but
// the feed continues with video only.
// The caller MUST NOT close or modify the returned track.
func (f *Feed) GetAudioTrack() *mediadevices.AudioTrack {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.audioTrack == nil || f.audioFailed {
		return nil
	}
	return f.audioTrack
}

// ============================================================================
// 4) METADATA GETTERS
// ============================================================================

// GetVideoDeviceName returns the friendly name of the video device.
// Returns empty string if no video device was acquired.
func (f *Feed) GetVideoDeviceName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.videoDevice == nil {
		return ""
	}
	return decodeDeviceLabel(f.videoDevice.Label)
}

// GetAudioDeviceName returns the friendly name of the audio device.
// Returns empty string if no audio device was acquired.
func (f *Feed) GetAudioDeviceName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.audioDevice == nil {
		return ""
	}
	return decodeDeviceLabel(f.audioDevice.Label)
}

// GetAcquiredAt returns the timestamp when the feed was successfully acquired.
// Returns zero time if the feed has never been acquired.
func (f *Feed) GetAcquiredAt() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.acquiredAt
}

// ============================================================================
// 5) INTERNAL HELPERS
// ============================================================================

// closeTracks closes all active tracks safely.
// Called during release to clean up media resources.
func (f *Feed) closeTracks() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// Close video track with panic recovery
	if f.videoTrack != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					f.logger.Warn("Panic while closing video track (recovered)", "panic", r)
				}
			}()
			f.videoTrack.Close()
		}()
		f.videoTrack = nil
	}
	
	// Close audio track with panic recovery
	if f.audioTrack != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					f.logger.Warn("Panic while closing audio track (recovered)", "panic", r)
				}
			}()
			f.audioTrack.Close()
		}()
		f.audioTrack = nil
	}
	
	// mediadevices.MediaStream doesn't have a Close() method
	// Tracks are closed individually above
	f.mediaStream = nil
}