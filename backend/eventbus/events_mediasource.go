// scenescheduler/backend/eventbus/events_mediasource.go
package eventbus

import (
    "time"

    "github.com/pion/mediadevices"
    "github.com/pion/webrtc/v4"
)

// This file contains event definitions related to the lifecycle and state
// of the MediaSource component.

// =============================================================================
// MediaSource Lifecycle Events
// =============================================================================

// MediaSourceReady is published when the media source is successfully acquired or re-acquired.
// It carries the essential resources (tracks, codecs) needed by other components
// like the WHEP handler. This event is the primary signal that the media is available.
type MediaSourceReady struct {
    Timestamp     time.Time
    VideoTrack    webrtc.TrackLocal           // The active video track. Can be nil if only audio is used.
    AudioTrack    webrtc.TrackLocal           // The active audio track. Can be nil if only video is used.
    CodecSelector *mediadevices.CodecSelector // The codecs configured for the tracks.
    VideoDeviceName string
    AudioDeviceName string
}

func (e MediaSourceReady) GetTopic() string { return "mediasource.lifecycle.ready" }

// MediaSourceLost is published by the track monitoring mechanism when a track
// unexpectedly closes. This is the trigger for recovery attempts and for other
// components to immediately release the source and close connections.
type MediaSourceLost struct {
    Timestamp time.Time
    Reason    string
}

func (e MediaSourceLost) GetTopic() string { return "mediasource.lifecycle.lost" }

// MediaAcquireFailed is published when an attempt to acquire the media source fails.
// This event is for informational purposes, allowing the UI to display a transient
// error. The manager may still attempt to recover later upon receiving new triggers.
type MediaAcquireFailed struct {
    Timestamp time.Time
    Reason    string // The error message encountered during acquisition.
}

func (e MediaAcquireFailed) GetTopic() string { return "mediasource.lifecycle.acquire_failed" }

// MediaSourceStopped is published when the media source is cleanly and
// intentionally released, for example, when the user stops the OBS virtual camera.
type MediaSourceStopped struct {
    Timestamp time.Time
    Reason    string
}

func (e MediaSourceStopped) GetTopic() string { return "mediasource.lifecycle.stopped" }

// MediaSourceUnrecoverable is published when the manager enters a permanent failure
// state, typically due to a driver or hardware issue (e.g., a timeout while closing a track).
// Once this event is fired, the manager will cease all attempts to acquire media.
// The UI should display a critical error and advise the user to restart the application.
type MediaSourceUnrecoverable struct {
    Timestamp time.Time
    Reason    string
}

func (e MediaSourceUnrecoverable) GetTopic() string { return "mediasource.lifecycle.unrecoverable" }
