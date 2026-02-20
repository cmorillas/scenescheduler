// backend/webserver/internal/whep/sessions.go
//
// This file handles the WHEP protocol's session management, including creation,
// deletion, and lifecycle of individual WebRTC peer connections.
// It is a part of the encapsulated `whep` internal package and has no knowledge
// of the main application's EventBus.
//
// Contents:
// - HTTP Handler
// - Session Creation
// - Session Deletion
// - Session Management (internal)
// - Helpers

package whep

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/pion/webrtc/v4"
)

// =============================================================================
// HTTP Handler
// =============================================================================

// HandleWhepRequest is the main HTTP entry point for all WHEP requests.
func (h *Handler) HandleWhepRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for browser compatibility
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Location")

	switch r.Method {
	case http.MethodPost:
		h.handleNewSession(w, r)
	case http.MethodDelete:
		h.handleDeleteSession(w, r)
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// =============================================================================
// Session Creation
// =============================================================================

// handleNewSession creates a new WebRTC session based on an incoming SDP offer.
func (h *Handler) handleNewSession(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("WHEP POST request received, creating new session.")

	// Atomically get the current tracks.
	h.tracksMu.RLock()
	videoTrack := h.videoTrack
	audioTrack := h.audioTrack
	h.tracksMu.RUnlock()

	// Check if media tracks from the parent are available.
	if videoTrack == nil && audioTrack == nil {
		h.logger.Warn("WHEP request rejected: media source tracks are not available.")
		http.Error(w, "Media source is not ready yet, please try again shortly", http.StatusServiceUnavailable)
		return
	}

	// Read SDP offer from the client.
	offerSDP, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read SDP offer", http.StatusBadRequest)
		return
	}
	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: string(offerSDP)}

	// Create a new PeerConnection.
	peerConnection, err := h.api.NewPeerConnection(webrtc.Configuration{ICEServers: h.iceServers})
	if err != nil {
		h.logger.Error("Failed to create PeerConnection", "error", err)
		http.Error(w, "Failed to create PeerConnection", http.StatusInternalServerError)
		return
	}

	sessionID := generateID()
	h.addSession(sessionID, peerConnection)

	// Cleanup helper for error cases during setup.
	cleanupOnError := func(logMsg string, userMsg string, statusCode int) {
		h.logger.Error(logMsg, "error", err, "sessionID", sessionID)
		go h.removeSession(sessionID) // Use goroutine to avoid blocking.
		http.Error(w, userMsg, statusCode)
	}

	// Monitor the connection state to clean up failed or closed connections.
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		h.logger.Debug("Peer Connection State changed", "sessionID", sessionID, "state", s.String())
		if s >= webrtc.PeerConnectionStateFailed {
			go func() {
				h.logger.Warn("Session entered failed/closed state, cleaning up.", "sessionID", sessionID)
				h.removeSession(sessionID)
			}()
		}
	})

	// Add tracks to the new connection.
	if videoTrack != nil {
		if _, err = peerConnection.AddTrack(videoTrack); err != nil {
			cleanupOnError("Failed to add video track", "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	if audioTrack != nil {
		if _, err = peerConnection.AddTrack(audioTrack); err != nil {
			cleanupOnError("Failed to add audio track", "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Set the client's offer.
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		cleanupOnError("Failed to set remote description", "Invalid SDP Offer", http.StatusBadRequest)
		return
	}

	// Create the server's answer.
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		cleanupOnError("Failed to create answer", "Could not create SDP Answer", http.StatusInternalServerError)
		return
	}

	// Wait for ICE gathering to complete before sending the answer.
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		cleanupOnError("Failed to set local description", "Internal server error", http.StatusInternalServerError)
		return
	}

	<-gatherComplete

	// Send the answer back to the client.
	w.Header().Set("Content-Type", "application/sdp")
	w.Header().Set("Location", "/whep/"+sessionID)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, peerConnection.LocalDescription().SDP)
	h.logger.Debug("WebRTC session created successfully", "sessionID", sessionID)
}

// =============================================================================
// Session Deletion
// =============================================================================

// handleDeleteSession terminates an existing WHEP session by its ID.
func (h *Handler) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := extractSessionID(r.URL.Path)
	if h.removeSession(sessionID) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Session not found", http.StatusNotFound)
	}
}

// =============================================================================
// Session Management (internal)
// =============================================================================

// addSession registers a new peer connection and notifies the parent.
func (h *Handler) addSession(id string, pc *webrtc.PeerConnection) {
	h.sessionsMu.Lock()
	h.sessions[id] = pc
	newCount := len(h.sessions)
	h.sessionsMu.Unlock()

	// Notify parent of the connection change via callback.
	if h.onConnectionsChanged != nil {
		h.onConnectionsChanged(newCount)
	}
}

// removeSession closes and removes a peer connection and notifies the parent.
// Returns true if a session was found and removed.
func (h *Handler) removeSession(id string) bool {
	h.sessionsMu.Lock()
	pc, ok := h.sessions[id]
	if ok {
		delete(h.sessions, id)
	}
	newCount := len(h.sessions)
	h.sessionsMu.Unlock()

	if pc != nil {
		// Close the connection if it's not already closed.
		if pc.ConnectionState() != webrtc.PeerConnectionStateClosed {
			h.logger.Debug("Closing peer connection for session", "sessionID", id)
			if err := pc.Close(); err != nil {
				h.logger.Error("Error during peer connection close", "sessionID", id, "error", err)
			}
		}
	}

	// Notify parent of the connection change if a session was actually removed.
	if ok && h.onConnectionsChanged != nil {
		h.onConnectionsChanged(newCount)
	}

	return ok
}

// CloseAllSessions terminates all active WebRTC sessions gracefully.
func (h *Handler) CloseAllSessions() {
	h.sessionsMu.Lock()
	sessionIDs := make([]string, 0, len(h.sessions))
	for id := range h.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	h.sessionsMu.Unlock()

	if len(sessionIDs) == 0 {
		return
	}

	h.logger.Debug("Closing all WebRTC sessions", "count", len(sessionIDs))

	var wg sync.WaitGroup
	for _, id := range sessionIDs {
		wg.Add(1)
		go func(sessionID string) {
			defer wg.Done()
			h.removeSession(sessionID)
		}(id)
	}
	wg.Wait()
}

// =============================================================================
// Helpers
// =============================================================================

// generateID creates a cryptographically secure random session ID.
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// This is a critical failure of the OS's entropy source.
		panic(fmt.Sprintf("failed to generate random ID: %v", err))
	}
	return hex.EncodeToString(bytes)
}

// extractSessionID parses the session ID from a WHEP resource URL path.
func extractSessionID(path string) string {
	// e.g., /whep/sessionId123 -> sessionId123
	trimmedPath := strings.TrimPrefix(path, "/whep/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
