// backend/gui/events.go
//
// EventBus subscription management and event handlers.
//
// Contents:
// - Subscription setup (called from New)
// - Event handlers
// - Unsubscribe cleanup

package gui

import (
	"fmt"
	"strings"

	"scenescheduler/backend/eventbus"
)

// =============================================================================
// EVENT SUBSCRIPTION MANAGEMENT
// =============================================================================

// subscribeToEvents sets up all event bus subscriptions for the GUI.
// This is called from New() to ensure the module is ready immediately.
func (g *GUI) subscribeToEvents() {
	g.logInfo("Subscribing GUI to application events.")

	var unsubscribe eventbus.UnsubscribeFunc
	var err error

	// OBS CONNECTION EVENTS
	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleOBSConnected)
	g.addUnsubscriber(unsubscribe, err, "OBSConnected")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleOBSDisconnected)
	g.addUnsubscriber(unsubscribe, err, "OBSDisconnected")

	// WEB SERVER STATUS EVENTS
	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleWebServerStarted)
	g.addUnsubscriber(unsubscribe, err, "WebServerStarted")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleWebServerStopped)
	g.addUnsubscriber(unsubscribe, err, "WebServerStopped")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleWebSocketClientsChanged)
	g.addUnsubscriber(unsubscribe, err, "WebSocketClientsChanged")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleWebRTCConnectionsChanged)
	g.addUnsubscriber(unsubscribe, err, "WebRTCConnectionsChanged")

	// MEDIA SOURCE STATUS EVENTS
	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleMediaSourceReady)
	g.addUnsubscriber(unsubscribe, err, "MediaSourceReady")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleMediaSourceStopped)
	g.addUnsubscriber(unsubscribe, err, "MediaSourceStopped")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleMediaSourceLost)
	g.addUnsubscriber(unsubscribe, err, "MediaSourceLost")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleMediaAcquireFailed)
	g.addUnsubscriber(unsubscribe, err, "MediaAcquireFailed")

	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleMediaSourceUnrecoverable)
	g.addUnsubscriber(unsubscribe, err, "MediaSourceUnrecoverable")

	// SCHEDULER EVENTS
	unsubscribe, err = eventbus.Subscribe(g.eventBus, "Gui", g.handleTargetProgramState)
	g.addUnsubscriber(unsubscribe, err, "TargetProgramState")
}

// unsubscribeAllEvents cleans up all event bus subscriptions.
// Called during cleanup to prevent memory leaks.
func (g *GUI) unsubscribeAllEvents() {
	g.logInfo("Unsubscribing from all application events.")
	for _, unsubscribe := range g.unsubscribeFuncs {
		if unsubscribe != nil {
			unsubscribe()
		}
	}
	g.unsubscribeFuncs = nil
}

// addUnsubscriber is a helper to reduce boilerplate in the subscription process.
func (g *GUI) addUnsubscriber(unsub eventbus.UnsubscribeFunc, err error, eventName string) {
	if err != nil {
		g.logError("Failed to subscribe to %s: %v", eventName, err)
	} else {
		g.unsubscribeFuncs = append(g.unsubscribeFuncs, unsub)
	}
}

// =============================================================================
// EVENT HANDLERS - OBS CONNECTION
// =============================================================================

// handleOBSConnected updates the UI connection status when OBS connects.
// Displays the OBS version in the connection status label.
//
// Topic: obs.system.connected
func (g *GUI) handleOBSConnected(event eventbus.OBSConnected) {
	g.updateConnectionStatus(fmt.Sprintf("Connected (OBS v%s)", event.OBSVersion))
}

// handleOBSDisconnected updates the UI connection status when OBS disconnects.
// Displays error information if the disconnection was due to an error.
//
// Topic: obs.system.disconnected
func (g *GUI) handleOBSDisconnected(event eventbus.OBSDisconnected) {
	statusText := "Disconnected"
	if event.Error != nil {
		statusText = fmt.Sprintf("Disconnected (Error: %v)", event.Error)
	}
	g.updateConnectionStatus(statusText)
}

// =============================================================================
// EVENT HANDLERS - WEB SERVER
// =============================================================================

// handleWebServerStarted updates the UI when the web server starts.
// Displays the server address, port, and TLS status.
//
// Topic: webserver.system.started
func (g *GUI) handleWebServerStarted(event eventbus.WebServerStarted) {
	var statusParts []string

	if len(event.IPs) > 0 {
		ipInfo := event.IPs[0]
		if len(event.IPs) > 1 {
			ipInfo += " (and others)"
		}
		statusParts = append(statusParts, fmt.Sprintf("Running at http://%s:%s", ipInfo, event.Port))
	} else {
		statusParts = append(statusParts, fmt.Sprintf("Running on port %s", event.Port))
	}

	if event.UseTLS {
		statusParts = append(statusParts, "(TLS)")
	}

	g.updateWebServerStatus(strings.Join(statusParts, " "))
}

// handleWebServerStopped updates the UI when the web server stops.
// Displays the reason for stopping and resets the user count.
//
// Topic: webserver.system.stopped
func (g *GUI) handleWebServerStopped(event eventbus.WebServerStopped) {
	g.updateWebServerStatus(fmt.Sprintf("Stopped (%s)", event.Reason))
	g.updateWebServerUsers("-")
}

// handleWebSocketClientsChanged updates the WebSocket client count in the UI.
//
// Topic: webserver.websocket.clientsChanged
func (g *GUI) handleWebSocketClientsChanged(event eventbus.WebSocketClientsChanged) {
	g.updateWebServerUsers(fmt.Sprintf("%d", event.Count))
}

// handleWebRTCConnectionsChanged updates the WebRTC connection count in the UI.
//
// Topic: webserver.webrtc.connectionsChanged
func (g *GUI) handleWebRTCConnectionsChanged(event eventbus.WebRTCConnectionsChanged) {
	g.updateLivePreviewUsers(fmt.Sprintf("%d", event.Count))
}

// =============================================================================
// EVENT HANDLERS - MEDIA SOURCE
// =============================================================================

// handleMediaSourceReady updates the UI when media source capture starts.
// Displays the video and audio device names.
//
// Topic: mediasource.state.ready
func (g *GUI) handleMediaSourceReady(event eventbus.MediaSourceReady) {
	g.updateLivePreviewStatus(fmt.Sprintf("Video: %s, Audio: %s", event.VideoDeviceName, event.AudioDeviceName))
}

// handleMediaSourceStopped updates the UI when media source capture stops.
// This is a normal stop, not an error condition.
//
// Topic: mediasource.state.stopped
func (g *GUI) handleMediaSourceStopped(event eventbus.MediaSourceStopped) {
	g.updateLivePreviewStatus("Inactive")
}

// handleMediaSourceLost updates the UI when media source is lost.
// This indicates the source disconnected or became unavailable.
//
// Topic: mediasource.state.lost
func (g *GUI) handleMediaSourceLost(event eventbus.MediaSourceLost) {
	g.updateLivePreviewStatus(fmt.Sprintf("Lost: %s", event.Reason))
}

// handleMediaAcquireFailed updates the UI when media source acquisition fails.
// This indicates an error during the initial connection attempt.
//
// Topic: mediasource.error.acquireFailed
func (g *GUI) handleMediaAcquireFailed(event eventbus.MediaAcquireFailed) {
	g.updateLivePreviewStatus(fmt.Sprintf("Failed: %s", event.Reason))
}

// handleMediaSourceUnrecoverable updates the UI when media source enters
// an unrecoverable error state. This indicates a fatal error that requires
// manual intervention.
//
// Topic: mediasource.error.unrecoverable
func (g *GUI) handleMediaSourceUnrecoverable(event eventbus.MediaSourceUnrecoverable) {
	g.updateLivePreviewStatus(fmt.Sprintf("FATAL ERROR: %s", event.Reason))
}

// =============================================================================
// EVENT HANDLERS - SCHEDULER
// =============================================================================

// handleTargetProgramState updates the GUI with the current and next programs.
// Only updates if the programs have actually changed to avoid unnecessary UI redraws.
// This operation is idempotent - receiving the same state multiple times
// causes no additional side effects.
//
// Topic: scheduler.state.targetProgram
func (g *GUI) handleTargetProgramState(event eventbus.TargetProgramState) {
	if g.shouldUpdateProgramPanels(event) {
		g.updateProgramPanels(event.TargetProgram, event.NextProgram)
	}
}