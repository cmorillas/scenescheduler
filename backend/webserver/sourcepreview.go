// backend/webserver/sourcepreview.go
//
// Bridge file for internal/sourcepreview module.
// Coordinates WebSocket preview requests with the sourcepreview manager.
// Handles request parsing, callback orchestration, and client-specific responses.

package webserver

import (
	"encoding/json"

	"scenescheduler/backend/webserver/internal/sourcepreview"
)

// =============================================================================
// Preview Request Handlers
// =============================================================================

// handleStartPreview processes a WebSocket request to start a source preview.
// It extracts the source info from the payload and delegates to the preview manager.
func (s *WebServer) handleStartPreview(clientID, remoteAddr string, payload json.RawMessage) {
	s.logger.Debug("Received startPreview request",
		"clientID", clientID,
		"remoteAddr", remoteAddr)

	// Check if preview manager is available
	if s.previewManager == nil {
		s.logger.Warn("Preview manager not available")
		s.sendPreviewError(clientID, "Preview system not available")
		return
	}

	// Parse request payload
	var req struct {
		InputKind     string                 `json:"inputKind"`
		URI           string                 `json:"uri"`
		InputSettings map[string]interface{} `json:"inputSettings"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		s.logger.Error("Failed to parse startPreview payload", "error", err)
		s.sendPreviewError(clientID, "Invalid preview request format")
		return
	}

	s.logger.Info("Starting preview",
		"clientID", clientID,
		"remoteAddr", remoteAddr,
		"inputKind", req.InputKind,
		"uri", req.URI)

	// Create preview request for the manager
	previewReq := sourcepreview.StartPreviewRequest{
		ConnectionID:  clientID,
		RemoteAddr:    remoteAddr,
		SourceURI:     req.URI,
		InputKind:     req.InputKind,
		InputSettings: req.InputSettings,

		// Callbacks for async result notification
		OnReady: func(hlsURL string) {
			s.sendPreviewReady(clientID, hlsURL)
		},
		OnError: func(errorMsg string) {
			s.sendPreviewError(clientID, errorMsg)
		},
		OnStopped: func(reason string) {
			s.sendPreviewStopped(clientID, reason)
		},
	}

	// Start preview (async - result via callbacks)
	if err := s.previewManager.StartPreview(previewReq); err != nil {
		s.logger.Error("Failed to start preview", "error", err)
		s.sendPreviewError(clientID, err.Error())
		return
	}

	// Success logged by callbacks
}

// handleStopPreview processes a WebSocket request to stop a source preview.
func (s *WebServer) handleStopPreview(clientID, remoteAddr string) {
	s.logger.Debug("Received stopPreview request",
		"clientID", clientID,
		"remoteAddr", remoteAddr)

	// Check if preview manager is available
	if s.previewManager == nil {
		s.logger.Warn("Preview manager not available")
		return
	}

	// Stop preview for this client
	if err := s.previewManager.StopPreview(clientID); err != nil {
		s.logger.Warn("Failed to stop preview", "error", err, "clientID", clientID, "remoteAddr", remoteAddr)
		return
	}

	s.logger.Debug("Preview stopped successfully", "clientID", clientID, "remoteAddr", remoteAddr)
}

// =============================================================================
// Response Helpers
// =============================================================================

// sendPreviewReady sends a previewReady message to the WebSocket client.
func (s *WebServer) sendPreviewReady(clientID, hlsURL string) {
	s.logger.Info("Preview ready, sending to client",
		"clientID", clientID,
		"hlsURL", hlsURL)

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"hlsUrl": hlsURL,
	})
	if err != nil {
		s.logger.Error("Failed to marshal previewReady payload", "error", err)
		return
	}

	s.wsHandler.SendToClient(clientID, "previewReady", json.RawMessage(payload))
}

// sendPreviewError sends a previewError message to the WebSocket client.
func (s *WebServer) sendPreviewError(clientID, errorMsg string) {
	s.logger.Warn("Preview error, sending to client",
		"clientID", clientID,
		"error", errorMsg)

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"error": errorMsg,
	})
	if err != nil {
		s.logger.Error("Failed to marshal previewError payload", "error", err)
		return
	}

	s.wsHandler.SendToClient(clientID, "previewError", json.RawMessage(payload))
}

// sendPreviewStopped sends a previewStopped message to the WebSocket client.
func (s *WebServer) sendPreviewStopped(clientID, reason string) {
	s.logger.Info("Preview stopped, sending notification to client",
		"clientID", clientID,
		"reason", reason)

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"reason": reason,
	})
	if err != nil {
		s.logger.Error("Failed to marshal previewStopped payload", "error", err)
		return
	}

	s.wsHandler.SendToClient(clientID, "previewStopped", json.RawMessage(payload))
}

// =============================================================================
// Disconnect Handler
// =============================================================================

// handleClientDisconnected is called when a WebSocket client disconnects.
// It ensures any active previews for that client are stopped and cleaned up.
//
// This should be called from the WebSocket disconnect callback in constructor.go.
func (s *WebServer) handleClientDisconnected(clientID, ip, userAgent, reason string) {
	s.logger.Debug("Client disconnected, checking for active previews",
		"clientID", clientID,
		"ip", ip,
		"reason", reason)

	// Stop any active preview for this client
	if s.previewManager != nil {
		if err := s.previewManager.StopPreview(clientID); err != nil {
			// Log at debug level since this is expected if no preview was active
			s.logger.Debug("No active preview to stop for disconnected client", "clientID", clientID, "ip", ip)
		} else {
			s.logger.Info("Stopped active preview for disconnected client", "clientID", clientID, "ip", ip)
		}
	}
}
