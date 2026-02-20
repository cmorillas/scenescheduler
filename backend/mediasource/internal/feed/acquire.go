// backend/mediasource/internal/feed/acquire.go
//
// This file contains resource acquisition logic with robust error handling.
// It discovers devices, builds constraints, opens media streams with retries,
// and initializes the Feed's internal state.
//
// Order of sections:
// 1) Public Acquisition Method
// 2) Constraint Building
// 3) Device Discovery Helpers
// 4) Stream Close Helper

package feed

import (
	"context"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/prop"
)

// Acquisition tuning parameters
const (
	maxAcquisitionRetries = 3
	acquisitionRetryDelay = 500 * time.Millisecond
)

// ============================================================================
// 1) PUBLIC ACQUISITION METHOD
// ============================================================================

// Acquire executes the complete acquisition process for this Feed.
// It discovers devices, builds constraints (including codec policy), opens the
// media stream with retries, extracts tracks, and initializes internal metadata.
//
// The Feed creates a child context from parentCtx for its own lifecycle.
// Cancelling parentCtx will abort the acquisition. After successful acquisition,
// the Feed manages its own context until Release() is called.
//
// Preconditions: cfg.VideoDeviceIdentifier and/or cfg.AudioDeviceIdentifier set.
// Side effects: populates mediaStream, videoTrack, audioTrack, device metadata;
//               initializes ctx/cancel; sets acquiredAt.
// Returns: error if devices not found or acquisition fails after retries.
func (f *Feed) Acquire(parentCtx context.Context) (err error) {
	// Add panic recovery to prevent silent crashes
	defer func() {
		if r := recover(); r != nil {
			f.logger.Error("Panic in Acquire", "panic", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic in Acquire: %v", r)
		}
	}()

	// Create feed-scoped context for lifecycle management
	ctx, cancel := context.WithCancel(parentCtx)
	f.ctx, f.cancel = ctx, cancel

	f.logger.Debug("Starting device acquisition")

	// 1) Enumerate available devices
	devices := mediadevices.EnumerateDevices()
	f.logger.Debug("Enumerated devices", "count", len(devices))

	// 2) Find configured devices
	var videoDev, audioDev *mediadevices.MediaDeviceInfo
	var errs []string

	if f.cfg.VideoDeviceIdentifier != "" {
		videoDev = findDevice(f.cfg.VideoDeviceIdentifier, mediadevices.VideoInput, devices)
		if videoDev == nil {
			errs = append(errs, fmt.Sprintf("video device %q not found", f.cfg.VideoDeviceIdentifier))
		} else {
			f.logger.Debug("Found video device", "identifier", f.cfg.VideoDeviceIdentifier, "label", decodeDeviceLabel(videoDev.Label))
		}
	}
	if f.cfg.AudioDeviceIdentifier != "" {
		audioDev = findDevice(f.cfg.AudioDeviceIdentifier, mediadevices.AudioInput, devices)
		if audioDev == nil {
			errs = append(errs, fmt.Sprintf("audio device %q not found", f.cfg.AudioDeviceIdentifier))
		} else {
			f.logger.Debug("Found audio device", "identifier", f.cfg.AudioDeviceIdentifier, "label", decodeDeviceLabel(audioDev.Label))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("device discovery failed: %s", strings.Join(errs, "; "))
	}
	if videoDev == nil && audioDev == nil {
		return fmt.Errorf("no devices configured")
	}

	// 3) Build media constraints
	constraints := f.buildConstraints(videoDev, audioDev)

	// 4) Acquire stream with GetUserMedia (with retries)
	// Retries are useful because VirtualCam events may arrive before device is ready.
	var stream mediadevices.MediaStream
	var lastErr error // Track the most recent error from GetUserMedia attempts
	
	for attempt := 1; attempt <= maxAcquisitionRetries; attempt++ {
		f.logger.Debug("Attempting GetUserMedia", "attempt", attempt, "maxRetries", maxAcquisitionRetries)
		
		// Wrap GetUserMedia in panic recovery
		var getUserMediaErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					f.logger.Error("Panic in GetUserMedia", "attempt", attempt, "panic", r)
					getUserMediaErr = fmt.Errorf("panic in GetUserMedia: %v", r)
				}
			}()
			stream, getUserMediaErr = mediadevices.GetUserMedia(constraints)
		}()
		
		if getUserMediaErr == nil {
			f.logger.Debug("GetUserMedia succeeded", "attempt", attempt)
			lastErr = nil // Clear any previous errors on success
			break
		}
		
		lastErr = getUserMediaErr
		f.logger.Warn("GetUserMedia failed", "attempt", attempt, "error", lastErr)
		
		// If we have more retries, wait before next attempt
		if attempt < maxAcquisitionRetries {
			select {
			case <-ctx.Done():
				return fmt.Errorf("acquisition aborted after %d attempts: %w", attempt, ctx.Err())
			case <-time.After(acquisitionRetryDelay):
				// Continue to next attempt
			}
		}
	}
	
	if lastErr != nil {
		return fmt.Errorf("GetUserMedia failed after %d attempts: %w", maxAcquisitionRetries, lastErr)
	}
	
	if stream == nil {
		return fmt.Errorf("GetUserMedia returned nil stream without error")
	}

	f.logger.Debug("Stream is valid, extracting tracks")

	// 5) Extract tracks from stream
	var vTrack *mediadevices.VideoTrack
	var aTrack *mediadevices.AudioTrack
	
	videoTracks := stream.GetVideoTracks()
	audioTracks := stream.GetAudioTracks()
	
	f.logger.Debug("Extracted tracks from stream", 
		"videoTracksCount", len(videoTracks), 
		"audioTracksCount", len(audioTracks))
	
	if len(videoTracks) > 0 {
		vTrack, _ = videoTracks[0].(*mediadevices.VideoTrack)
		f.logger.Debug("Video track cast result", "success", vTrack != nil)
	}
	if len(audioTracks) > 0 {
		aTrack, _ = audioTracks[0].(*mediadevices.AudioTrack)
		f.logger.Debug("Audio track cast result", "success", aTrack != nil)
	}

	// 6) Validate that requested tracks are present
	if videoDev != nil && vTrack == nil {
		f.logger.Error("Video track validation failed", "videoDevConfigured", videoDev != nil, "vTrackNil", vTrack == nil)
		closeStreamSafely(stream, 2*time.Second)
		return fmt.Errorf("video track not acquired from stream")
	}
	if audioDev != nil && aTrack == nil {
		f.logger.Error("Audio track validation failed", "audioDevConfigured", audioDev != nil, "aTrackNil", aTrack == nil)
		closeStreamSafely(stream, 2*time.Second)
		return fmt.Errorf("audio track not acquired from stream")
	}

	f.logger.Debug("Track validation passed, storing resources")

	// 7) Store acquired resources in feed (under lock)
	f.mu.Lock()
	f.mediaStream = stream
	f.videoTrack = vTrack
	f.audioTrack = aTrack
	f.videoDevice = videoDev
	f.audioDevice = audioDev
	f.acquiredAt = time.Now()
	f.videoFailed = false
	f.audioFailed = false
	f.mu.Unlock()

	return nil
}

// ============================================================================
// 2) CONSTRAINT BUILDING
// ============================================================================

// buildConstraints creates media constraints for GetUserMedia based on
// discovered devices and quality profile settings from the config.
// The Manager has already expanded cfg with derived quality parameters.
func (f *Feed) buildConstraints(videoDevice, audioDevice *mediadevices.MediaDeviceInfo) mediadevices.MediaStreamConstraints {
	
	constraints := mediadevices.MediaStreamConstraints{
		Codec: f.codecSelector,
	}

	if videoDevice != nil {
		constraints.Video = func(c *mediadevices.MediaTrackConstraints) {
			c.DeviceID = prop.String(videoDevice.DeviceID)
			c.Width = prop.Int(f.cfg.VideoWidth)
			c.Height = prop.Int(f.cfg.VideoHeight)
		}
	}

	if audioDevice != nil {
		constraints.Audio = func(c *mediadevices.MediaTrackConstraints) {
			c.DeviceID = prop.String(audioDevice.DeviceID)
			c.ChannelCount = prop.Int(f.cfg.AudioChannels)
			c.SampleRate = prop.Int(f.cfg.AudioSampleRate)
		}
	}

	return constraints
}

// ============================================================================
// 3) DEVICE DISCOVERY HELPERS
// ============================================================================

// findDevice looks for a device in a list by matching its identifier against
// the device's ID or label (including hex-decoded label).
// Returns nil if no matching device is found.
func findDevice(identifier string, deviceKind mediadevices.MediaDeviceType, all []mediadevices.MediaDeviceInfo) *mediadevices.MediaDeviceInfo {
	if identifier == "" {
		return nil
	}
	
	for i := range all {
		dev := all[i]
		if dev.Kind != deviceKind {
			continue
		}
		
		// Match against DeviceID, raw label, or decoded label
		if dev.DeviceID == identifier || 
		   dev.Label == identifier || 
		   decodeDeviceLabel(dev.Label) == identifier {
			return &all[i]
		}
	}
	return nil
}

// decodeDeviceLabel attempts to decode a hex-encoded device label.
// If decoding fails or produces non-printable characters, returns the raw label.
// Returns "Unknown Device" for empty labels.
func decodeDeviceLabel(label string) string {
	if label == "" {
		return "Unknown Device"
	}
	
	decoded, err := hex.DecodeString(label)
	if err != nil || len(decoded) == 0 {
		return label
	}
	
	// Validate that decoded bytes are printable ASCII
	for _, b := range decoded {
		if b < 32 || b > 126 {
			return label
		}
	}
	
	return string(decoded)
}

// ============================================================================
// 4) STREAM CLOSE HELPER
// ============================================================================

// closeStreamSafely closes a media stream with a timeout to prevent hanging.
// Used for cleanup when acquisition partially succeeds but validation fails.
func closeStreamSafely(stream mediadevices.MediaStream, timeout time.Duration) {
	ch := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Silent recovery - closing is best effort
			}
		}()
		
		// Close all tracks individually (mediaStream doesn't have Close())
		for _, track := range stream.GetTracks() {
			track.Close()
		}
		close(ch)
	}()

	select {
	case <-ch:
		// Closed successfully
	case <-time.After(timeout):
		// Timed out, but nothing we can do
	}
}