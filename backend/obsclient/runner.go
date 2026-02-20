// backend/obsclient/runner.go
//
// This file defines the lifecycle orchestration for the OBSClient module.
// The runner is responsible for starting, supervising, and gracefully
// stopping the OBSClient. It manages the Finite State Machine (FSM) loop.
//
// Contents:
// - Public Lifecycle Methods (Run, Stop)
// - Internal Lifecycle Helpers (cleanup)

package obsclient

import (
	
)

// ============================================================================
// PUBLIC LIFECYCLE METHODS
// ============================================================================

// Run starts the OBS client lifecycle, including the FSM loop.
// It blocks until the context is canceled.
// The context for this module already exists from the constructor.
func (c *OBSClient) Run() {
	defer c.cleanup()

	c.logger.Debug("OBS client runner starting.")

	c.signalFSM()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.InfoGui("OBS client shutting down.")
			return
		case <-c.signalCh:
			// The switch uses the public GetState accessor for thread-safety.
			switch c.GetState() {
			case StateDisconnected:
				go c.handleDisconnectedState()
			case StateConnecting:
				go c.handleConnectingState()
			case StateConnected:
				go c.handleConnectedState()
			case StateReconnecting:
				go c.handleReconnectingState()
			}
		}
	}
}

// Stop gracefully shuts down the client by cancelling its context.
func (c *OBSClient) Stop() {
	c.stopOnce.Do(func() {
		c.logger.InfoGui("Stop requested for OBS client.")
		if c.cancelCtx != nil {
			c.cancelCtx()
		}
	})
}

// ============================================================================
// INTERNAL LIFECYCLE HELPERS
// ============================================================================

// cleanup is the internal finalizer for the module, ensuring all resources
// are released. It is idempotent.
func (c *OBSClient) cleanup() {
	c.cleanupOnce.Do(func() {
		c.logger.Debug("Cleaning up OBS client resources.")
		c.unsubscribeAllEvents()
		// Ensure a final disconnect call to clean up any active session.
		c.disconnect("client shutdown")
	})
}