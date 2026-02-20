// backend/obsclient/events.go
//
// This file wires the OBSClient to the application's event bus. Handlers here
// remain minimal: they validate inputs, log context, and delegate to themed
// implementation files (e.g., switcher.go).
//
// Contents:
// - Subscription Management
// - Event Handlers

package obsclient

import (
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// SUBSCRIPTION MANAGEMENT
// ============================================================================

// subscribeToEvents sets up all event bus subscriptions for the client.
// This is called from New() to ensure the module is ready immediately.
func (c *OBSClient) subscribeToEvents() {
	c.logger.Debug("Subscribing to application events.")

	unsub1, err1 := eventbus.Subscribe(c.bus, "ObsClient", c.handleTargetProgramState)
	if err1 != nil {
		c.logger.Error("Failed to subscribe to TargetProgramState", "error", err1)
		return
	}
	c.unsubscribeFuncs = append(c.unsubscribeFuncs, unsub1)

	unsub2, err2 := eventbus.Subscribe(c.bus, "ObsClient", c.handleGetStatusRequested)
	if err2 != nil {
		c.logger.Error("Failed to subscribe to GetStatusRequested", "error", err2)
		return
	}
	c.unsubscribeFuncs = append(c.unsubscribeFuncs, unsub2)
}

// unsubscribeAllEvents cleans up all event bus subscriptions.
func (c *OBSClient) unsubscribeAllEvents() {
	c.logger.Debug("Unsubscribing from all application events.")
	for _, unsubscribe := range c.unsubscribeFuncs {
		if unsubscribe != nil {
			unsubscribe()
		}
	}
	c.unsubscribeFuncs = nil
}

// ============================================================================
// EVENT HANDLERS
// ============================================================================

// handleTargetProgramState processes the desired state declaration from the scheduler.
// It compares the target state with its own internal state and converges if necessary.
//
// Topic:   scheduler.state.targetProgram
func (c *OBSClient) handleTargetProgramState(event eventbus.TargetProgramState) {
	if c.GetState() != StateConnected {
		// Do not log here, as this will be the normal state when OBS is disconnected.
		// It would generate too much noise.
		return
	}

	// Delegate the convergence logic to the dedicated method in switcher.go
	c.convergeToState(event)
}

// handleGetStatusRequested responds to status requests from clients.
// It queries the current OBS connection and VirtualCam state, then publishes a StatusResponse.
//
// Topic:   webserver.command.getStatus
func (c *OBSClient) handleGetStatusRequested(event eventbus.GetStatusRequested) {
	c.logger.Debug("Received status request", "clientID", event.ClientID)

	status := c.GetCurrentStatus()

	eventbus.Publish(c.bus, eventbus.StatusResponse{
		ClientID:         event.ClientID,
		OBSConnected:     status.IsConnected,
		OBSVersion:       status.OBSVersion,
		VirtualCamActive: status.VirtualCamActive,
	})
}
