// backend/mediasource/events.go
//
// This file manages EventBus subscriptions and handlers for the Manager.
//
// Contents:
// - Subscription Management
// - Event Handlers

package mediasource

import (
	"time"
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// Subscription Management
// ============================================================================

// subscribeToEvents wires all event bus subscriptions for the Manager.
// This is called from New() to ensure the module is ready immediately.
func (m *Manager) subscribeToEvents() {
	m.logger.Debug("Subscribing to application events.")

	unsub1, err1 := eventbus.Subscribe(m.eventBus, "MediaSource", m.handleVirtualCamStarted)
	m.addUnsubscriber(unsub1, err1, "OBSVirtualCamStarted")

	unsub2, err2 := eventbus.Subscribe(m.eventBus, "MediaSource", m.handleVirtualCamStopped)
	m.addUnsubscriber(unsub2, err2, "OBSVirtualCamStopped")
}

// unsubscribeAllEvents cleans up all event subscriptions.
func (m *Manager) unsubscribeAllEvents() {
	m.logger.Debug("Unsubscribing from all events.")
	for _, unsubscribe := range m.unsubscribeFuncs {
		unsubscribe()
	}
	m.unsubscribeFuncs = nil
}

// addUnsubscriber is a helper to reduce boilerplate in the subscription process.
func (m *Manager) addUnsubscriber(unsub eventbus.UnsubscribeFunc, err error, name string) {
	if err != nil {
		m.logger.Error("Failed to subscribe to event", "event", name, "error", err)
	} else {
		m.unsubscribeFuncs = append(m.unsubscribeFuncs, unsub)
	}
}

// ============================================================================
// Event Handlers
// ============================================================================

// handleVirtualCamStarted reacts to the OBS virtual camera being activated.
// It triggers a media acquisition attempt after a short delay to allow
// the device to fully initialize.
// Topic: obs.virtualcam.started
func (m *Manager) handleVirtualCamStarted(event eventbus.OBSVirtualCamStarted) {
	m.logger.Debug("Event received: OBSVirtualCamStarted")

	// Launch in background to avoid blocking the EventBus.
	// Wait for device initialization before attempting acquisition.
	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			m.requestAcquisition()
		case <-m.ctx.Done():
			m.logger.Debug("Context cancelled while waiting to acquire device")
			return
		}
	}()
}

// handleVirtualCamStopped reacts to the OBS virtual camera being deactivated.
// It triggers a graceful release of the media feed.
// Topic: obs.virtualcam.stopped
func (m *Manager) handleVirtualCamStopped(event eventbus.OBSVirtualCamStopped) {
	m.logger.Debug("Event received: OBSVirtualCamStopped")
	m.requestRelease("OBS Virtual Camera stopped")
}
