// backend/webserver/internal/whep/handler.go
//
// This file defines the WHEP handler's core struct and its lifecycle methods.
// The handler is an encapsulated internal component of the WebServer, responsible
// for managing the state of available media tracks and WebRTC session lifecycle.
//
// It strictly follows the "ambassador" pattern: it has no knowledge of the
// EventBus and is controlled entirely by its parent, the WebServer module.
//
// Contents:
// - Struct Definition
// - Constructor (New)
// - Public API (for parent module)
// - Lifecycle Methods
// - Helpers

package whep

import (
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
	"scenescheduler/backend/logger"
)

// =============================================================================
// Struct Definition
// =============================================================================

// Handler manages WHEP sessions and the lifecycle of media tracks for streaming.
type Handler struct {
	// --- Dependencies ---
	logger               *logger.Logger
	api                  *webrtc.API
	iceServers           []webrtc.ICEServer
	onConnectionsChanged func(count int) // Callback to notify parent of changes

	// --- Internal State ---
	sessions   map[string]*webrtc.PeerConnection
	sessionsMu sync.RWMutex

	videoTrack webrtc.TrackLocal
	audioTrack webrtc.TrackLocal
	tracksMu   sync.RWMutex
}

// =============================================================================
// Constructor
// =============================================================================

// New creates a new WHEP handler.
// It requires a callback function to notify its parent about connection count changes,
// ensuring it remains decoupled from the EventBus.
func New(log *logger.Logger, onConnectionsChanged func(count int)) *Handler {
	handlerLogger := log.WithModule("whep")

	// Setup WebRTC API
	api, err := createWebRTCAPI()
	if err != nil {
		// A failure here is unrecoverable, so panic is appropriate.
		panic(fmt.Errorf("failed to create WebRTC API: %w", err))
	}

	return &Handler{
		logger:               handlerLogger,
		api:                  api,
		iceServers:           []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
		sessions:             make(map[string]*webrtc.PeerConnection),
		onConnectionsChanged: onConnectionsChanged,
	}
}

// =============================================================================
// Public API (for parent module)
// =============================================================================

// SetTracks is called by the parent WebServer to update the available media tracks.
// This is the controlled entry point for external state changes.
func (h *Handler) SetTracks(video, audio webrtc.TrackLocal) {
	h.tracksMu.Lock()
	defer h.tracksMu.Unlock()

	// Only log if there's an actual change.
	if h.videoTrack != video || h.audioTrack != audio {
		h.logger.Debug("Media tracks have been updated by parent server",
			"hasVideo", video != nil,
			"hasAudio", audio != nil,
		)
		h.videoTrack = video
		h.audioTrack = audio

		// If all tracks are removed, proactively close existing sessions.
		if video == nil && audio == nil {
			go h.CloseAllSessions()
		}
	}
}

// =============================================================================
// Lifecycle Methods
// =============================================================================

// Shutdown performs graceful cleanup of all WHEP resources.
// This is called by the parent WebServer during its cleanup phase.
func (h *Handler) Shutdown() {
	h.logger.Debug("WHEP handler shutting down.")
	h.CloseAllSessions()
}

// =============================================================================
// Helpers
// =============================================================================

// createWebRTCAPI initializes the Pion WebRTC API with a custom setting engine.
func createWebRTCAPI() (*webrtc.API, error) {
	// Currently, no custom settings are needed, but this is the place to put
	// them (e.g., network types, ICE timeouts).
	settingEngine := webrtc.SettingEngine{}
	return webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine)), nil
}