// backend/mediasource/eventpub.go
//
// This file contains all methods responsible for publishing events from the
// mediasource Manager to the application's EventBus.
//
// Contents:
// - Event Publishing Methods

package mediasource

import (
	"time"

	"github.com/pion/webrtc/v4"
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// Event Publishing Methods
// ============================================================================

// publishMediaReady notifies that media has been successfully acquired.
func (m *Manager) publishMediaReady() {
	feed := m.getFeed()
	if feed == nil {
		m.logger.Warn("Cannot publish MediaReady: no active feed")
		return
	}

	var videoTrack, audioTrack webrtc.TrackLocal
	if vt := feed.GetVideoTrack(); vt != nil {
		videoTrack = &TrackProxy{TrackLocal: vt}
	}
	if at := feed.GetAudioTrack(); at != nil {
		audioTrack = &TrackProxy{TrackLocal: at}
	}

	videoName := feed.GetVideoDeviceName()
	audioName := feed.GetAudioDeviceName()

	eventbus.Publish(m.eventBus, eventbus.MediaSourceReady{
		Timestamp:       time.Now(),
		VideoTrack:      videoTrack,
		AudioTrack:      audioTrack,
		CodecSelector:   m.codecSelector,
		VideoDeviceName: videoName,
		AudioDeviceName: audioName,
	})

	m.logger.Debug("Published MediaSourceReady event",
		"hasVideo", videoTrack != nil, "hasAudio", audioTrack != nil)
}

// publishMediaLost notifies that the media feed was unexpectedly lost.
func (m *Manager) publishMediaLost(reason string) {
	eventbus.Publish(m.eventBus, eventbus.MediaSourceLost{
		Timestamp: time.Now(),
		Reason:    reason,
	})
	m.logger.Warn("Published MediaSourceLost event", "reason", reason)
}

// publishMediaStopped notifies that the media feed was intentionally stopped.
func (m *Manager) publishMediaStopped(reason string) {
	eventbus.Publish(m.eventBus, eventbus.MediaSourceStopped{
		Timestamp: time.Now(),
		Reason:    reason,
	})
	m.logger.Debug("Published MediaSourceStopped event", "reason", reason)
}

// publishMediaAcquireFailed notifies that an acquisition attempt failed.
func (m *Manager) publishMediaAcquireFailed(reason string) {
	eventbus.Publish(m.eventBus, eventbus.MediaAcquireFailed{
		Timestamp: time.Now(),
		Reason:    reason,
	})
	m.logger.Error("Published MediaAcquireFailed event", "reason", reason)
}
