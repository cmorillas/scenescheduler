// backend/mediasource/helpers.go
//
// This file contains pure, stateless helper functions for the mediasource module.
// These functions have no receiver and do not depend on Manager state.
//
// Contents:
// - Codec and Quality Profile Configuration
// - Device Label Decoding

package mediasource

import (
	"encoding/hex"
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/codec/vpx"
	"scenescheduler/backend/config"
)

// ============================================================================
// Codec and Quality Profile Configuration
// ============================================================================

// qualityProfile defines encoding parameters for a specific quality level.
type qualityProfile struct {
	VideoWidth   int
	VideoHeight  int
	VideoBitrate int // In bps
	AudioChannels int
	AudioSampleRate int
}

// applyQualityProfile expands the config with derived quality parameters.
func applyQualityProfile(cfg *config.MediaSourceConfig) {
	profile := getQualityProfile(cfg.Quality)
	cfg.VideoWidth = profile.VideoWidth
	cfg.VideoHeight = profile.VideoHeight
	cfg.VideoBitrate = profile.VideoBitrate
	cfg.AudioChannels = profile.AudioChannels
	cfg.AudioSampleRate = profile.AudioSampleRate
}

// createCodecSelector builds a codec selector with quality-tuned parameters.
func createCodecSelector(cfg *config.MediaSourceConfig) (*mediadevices.CodecSelector, error) {
	vp8Params, err := vpx.NewVP8Params()
	if err != nil {
		return nil, fmt.Errorf("failed to create VP8 parameters: %w", err)
	}
	vp8Params.BitRate = cfg.VideoBitrate

	opusParams, err := opus.NewParams()
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus parameters: %w", err)
	}

	return mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&vp8Params),
		mediadevices.WithAudioEncoders(&opusParams),
	), nil
}

// getQualityProfile returns the encoding parameters for a given quality level.
func getQualityProfile(quality string) qualityProfile {
	profiles := map[string]qualityProfile{
		"low":    {640, 480, 500000, 1, 16000},
		"medium": {1280, 720, 2000000, 2, 48000},
		"high":   {1920, 1080, 4000000, 2, 48000},
		"ultra":  {3840, 2160, 8000000, 2, 48000},
	}
	if p, ok := profiles[quality]; ok {
		return p
	}
	return profiles["low"] // Default fallback
}

// ============================================================================
// Device Label Decoding
// ============================================================================

// decodeDeviceLabel attempts to decode a hex-encoded device label.
func decodeDeviceLabel(label string) string {
	if label == "" {
		return "Unknown Device"
	}
	decoded, err := hex.DecodeString(label)
	if err != nil {
		return label
	}
	for _, b := range decoded {
		if b < 32 || b > 126 {
			return label // Not printable ASCII
		}
	}
	return string(decoded)
}
