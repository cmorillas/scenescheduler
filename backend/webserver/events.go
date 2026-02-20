// backend/webserver/events.go
//
// This file contains all EventBus subscription management and event handlers
// for the WebServer module.
//
// Contents:
// - Subscription Setup
// - Event Handlers (MediaSource)
// - Event Handlers (WebSocket Control)

package webserver

import (
	"encoding/json"

	"scenescheduler/backend/eventbus"
)

// =============================================================================
// Subscription Setup
// =============================================================================

// subscribeToEvents sets up listeners for events the WebServer needs to react to.
// This is called from New() to ensure the module is ready immediately.
func (s *WebServer) subscribeToEvents() {
	s.logger.Debug("WebServer subscribing to application events.")

	// MediaSource events (relayed to WHEP handler)
	unsub1, err1 := eventbus.Subscribe(s.bus, "WebServer", s.handleMediaSourceReady)
	s.addUnsubscriber(unsub1, err1, "MediaSourceReady")

	unsub2, err2 := eventbus.Subscribe(s.bus, "WebServer", s.handleMediaSourceStopped)
	s.addUnsubscriber(unsub2, err2, "MediaSourceStopped")

	// WebSocket control events (relayed to WebSocket handler)
	unsub3, err3 := eventbus.Subscribe(s.bus, "WebServer", s.handleSendMessageToClient)
	s.addUnsubscriber(unsub3, err3, "WebSocketSendMessageToClient")

	unsub4, err4 := eventbus.Subscribe(s.bus, "WebServer", s.handleBroadcastMessage)
	s.addUnsubscriber(unsub4, err4, "WebSocketBroadcastMessage")

	unsub5, err5 := eventbus.Subscribe(s.bus, "WebServer", s.handleOBSProgramChanged)
	s.addUnsubscriber(unsub5, err5, "OBSProgramChanged")

	// OBS connection events (broadcast to all clients)
	unsub6, err6 := eventbus.Subscribe(s.bus, "WebServer", s.handleOBSConnected)
	s.addUnsubscriber(unsub6, err6, "OBSConnected")

	unsub7, err7 := eventbus.Subscribe(s.bus, "WebServer", s.handleOBSDisconnected)
	s.addUnsubscriber(unsub7, err7, "OBSDisconnected")

	// OBS VirtualCam events (broadcast to all clients for preview status)
	unsub8, err8 := eventbus.Subscribe(s.bus, "WebServer", s.handleVirtualCamStarted)
	s.addUnsubscriber(unsub8, err8, "OBSVirtualCamStarted")

	unsub9, err9 := eventbus.Subscribe(s.bus, "WebServer", s.handleVirtualCamStopped)
	s.addUnsubscriber(unsub9, err9, "OBSVirtualCamStopped")

	// Status response (send to specific client that requested it)
	unsub10, err10 := eventbus.Subscribe(s.bus, "WebServer", s.handleStatusResponse)
	s.addUnsubscriber(unsub10, err10, "StatusResponse")
}

// addUnsubscriber is a helper to reduce boilerplate in the subscription process.
func (s *WebServer) addUnsubscriber(unsub eventbus.UnsubscribeFunc, err error, topic string) {
	if err != nil {
		s.logger.Error("Failed to subscribe to event", "topic", topic, "error", err)
	} else {
		s.unsubscribeFuncs = append(s.unsubscribeFuncs, unsub)
	}
}

// unsubscribeAllEvents cleans up all event subscriptions.
// Called during server shutdown to prevent memory leaks.
func (s *WebServer) unsubscribeAllEvents() {
	s.logger.Debug("Unsubscribing from all events.")
	for _, unsub := range s.unsubscribeFuncs {
		unsub()
	}
	s.unsubscribeFuncs = nil
}

// =============================================================================
// Event Handlers (MediaSource)
// =============================================================================

// handleMediaSourceReady receives track info from the MediaSource module
// and relays it to the internal WHEP handler via a direct method call.
//
// Topic: mediasource.lifecycle.ready
func (s *WebServer) handleMediaSourceReady(event eventbus.MediaSourceReady) {
	s.logger.Debug("Received MediaSourceReady, relaying tracks to WHEP handler.")
	if s.whepHandler != nil {
		s.whepHandler.SetTracks(event.VideoTrack, event.AudioTrack)
	}
}

// handleMediaSourceStopped is notified when the media source is no longer available.
// It instructs the internal WHEP handler to clear its tracks.
//
// Topic: mediasource.lifecycle.stopped
func (s *WebServer) handleMediaSourceStopped(event eventbus.MediaSourceStopped) {
	s.logger.Debug("Received MediaSourceStopped, relaying to WHEP handler.", "reason", event.Reason)
	if s.whepHandler != nil {
		s.whepHandler.SetTracks(nil, nil)
	}
}

// =============================================================================
// Event Handlers (WebSocket Control)
// =============================================================================

// handleSendMessageToClient receives a request to send a message to a specific
// WebSocket client and delegates to the internal WebSocket handler.
//
// Topic: websocket.command.sendToClient
func (s *WebServer) handleSendMessageToClient(event eventbus.WebSocketSendMessageToClient) {
	s.logger.Debug("Relaying message to specific WebSocket client",
		"clientID", event.ClientID,
		"messageType", event.MessageType)

	if s.wsHandler != nil {
		s.wsHandler.SendToClient(event.ClientID, event.MessageType, event.Payload)
	}
}

// handleBroadcastMessage receives a request to broadcast a message to all
// WebSocket clients and delegates to the internal WebSocket handler.
//
// Topic: websocket.command.broadcast
func (s *WebServer) handleBroadcastMessage(event eventbus.WebSocketBroadcastMessage) {
	s.logger.Debug("Broadcasting message to all WebSocket clients",
		"messageType", event.MessageType)

	if s.wsHandler != nil {
		s.wsHandler.Broadcast(event.MessageType, event.Payload)
	}
}

// handleOBSProgramChanged broadcasts OBS program changes to all WebSocket clients.
//
// Topic: obs.program.changed
func (s *WebServer) handleOBSProgramChanged(event eventbus.OBSProgramChanged) {
	prevTitle := "<none>"
	if event.PreviousProgram != nil && event.PreviousProgram.Title != "" {
		prevTitle = event.PreviousProgram.Title
	}
	currTitle := "<none>"
	if event.CurrentProgram != nil && event.CurrentProgram.Title != "" {
		currTitle = event.CurrentProgram.Title
	}

	s.logger.Debug("OBS program changed",
		"previousProgram", prevTitle,
		"currentProgram", currTitle,
		"offsetMs", event.SeekOffsetMs)

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal OBSProgramChanged payload", "error", err)
		return
	}

	s.wsHandler.Broadcast("obsProgramChanged", json.RawMessage(payload))
}

// =============================================================================
// Event Handlers (OBS Connection Status)
// =============================================================================

// handleOBSConnected broadcasts OBS connection status to all WebSocket clients.
//
// Topic: obs.system.connected
func (s *WebServer) handleOBSConnected(event eventbus.OBSConnected) {
	s.logger.Debug("OBS connected, broadcasting to clients", "version", event.OBSVersion)

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal OBSConnected payload", "error", err)
		return
	}

	s.wsHandler.Broadcast("obsConnected", json.RawMessage(payload))
}

// handleOBSDisconnected broadcasts OBS disconnection status to all WebSocket clients.
//
// Topic: obs.system.disconnected
func (s *WebServer) handleOBSDisconnected(event eventbus.OBSDisconnected) {
	s.logger.Debug("OBS disconnected, broadcasting to clients")

	if s.wsHandler == nil {
		return
	}

	// Create a simpler payload without the error object (errors don't marshal well to JSON)
	simplePayload := map[string]interface{}{
		"timestamp": event.Timestamp,
		"reason":    "disconnected",
	}

	payload, err := json.Marshal(simplePayload)
	if err != nil {
		s.logger.Error("Failed to marshal OBSDisconnected payload", "error", err)
		return
	}

	s.wsHandler.Broadcast("obsDisconnected", json.RawMessage(payload))
}

// =============================================================================
// Event Handlers (OBS VirtualCam Status)
// =============================================================================

// handleVirtualCamStarted broadcasts VirtualCam start event to all WebSocket clients.
// This indicates that video stream is now available for preview.
//
// Topic: obs.virtualcam.started
func (s *WebServer) handleVirtualCamStarted(event eventbus.OBSVirtualCamStarted) {
	s.logger.Debug("VirtualCam started, broadcasting to clients")

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal VirtualCamStarted payload", "error", err)
		return
	}

	s.wsHandler.Broadcast("virtualCamStarted", json.RawMessage(payload))
}

// handleVirtualCamStopped broadcasts VirtualCam stop event to all WebSocket clients.
// This indicates that video stream is no longer available.
//
// Topic: obs.virtualcam.stopped
func (s *WebServer) handleVirtualCamStopped(event eventbus.OBSVirtualCamStopped) {
	s.logger.Debug("VirtualCam stopped, broadcasting to clients")

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal VirtualCamStopped payload", "error", err)
		return
	}

	s.wsHandler.Broadcast("virtualCamStopped", json.RawMessage(payload))
}

// =============================================================================
// Event Handlers (Status Response)
// =============================================================================

// handleStatusResponse sends the system status to the specific client that requested it.
//
// Topic: webserver.response.status
func (s *WebServer) handleStatusResponse(event eventbus.StatusResponse) {
	s.logger.Debug("Received status response, sending to client", "clientID", event.ClientID)

	if s.wsHandler == nil {
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"obsConnected":     event.OBSConnected,
		"obsVersion":       event.OBSVersion,
		"virtualCamActive": event.VirtualCamActive,
	})
	if err != nil {
		s.logger.Error("Failed to marshal status response", "error", err)
		return
	}

	s.wsHandler.SendToClient(event.ClientID, "currentStatus", json.RawMessage(payload))
}
