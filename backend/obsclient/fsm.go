// backend/obsclient/fsm.go
//
// This file contains the implementation for the Finite State Machine (FSM).
// It includes the logic handlers for each state and the helpers to manage
// state transitions.
//
// Contents:
// - FSM State Handler Implementations
// - FSM Helper Implementations
// - Public State Accessors

package obsclient

import (
	"fmt"
	"time"
)

// ============================================================================
// FSM STATE HANDLER IMPLEMENTATIONS
// ============================================================================

// handleDisconnectedState is the logic for the Disconnected state.
func (c *OBSClient) handleDisconnectedState() {
	c.logger.Debug("State: Disconnected. Transitioning to Connecting.")
	c.setState(StateConnecting)
}

// handleConnectingState is the logic for the Connecting state.
func (c *OBSClient) handleConnectingState() {
	c.logger.Debug("State: Connecting. Attempting to establish connection...")
	if err := c.connect(); err != nil {
		c.logger.Warn("Connection attempt failed.", "error", err)
		c.setState(StateReconnecting)
		return
	}
	c.setState(StateConnected)
}

// handleConnectedState is the logic for the Connected state.
func (c *OBSClient) handleConnectedState() {
	c.logger.Debug("State: Connected. Performing initial setup...")
	if err := c.setupScene(); err != nil {
		c.logger.Error("Critical scene setup failed. Disconnecting.", "error", err)
		c.disconnect(fmt.Sprintf("scene setup failed: %v", err))
		c.setState(StateReconnecting)
		return
	}

	// After setup, check the initial state of OBS components to sync up.
	go c.checkInitialState()

	c.stateMu.RLock()
	if c.connection == nil {
		c.stateMu.RUnlock()
		c.logger.Error("Entered Connected state with a nil connection. Reconnecting.")
		c.setState(StateReconnecting)
		return
	}
	sessionCtx := c.connection.ctx
	c.stateMu.RUnlock()

	go c.monitorConnection()
	go c.startOBSEventListener()

	// Block here until the session context is cancelled (e.g., by disconnection).
	<-sessionCtx.Done()
	c.setState(StateReconnecting)
}

// handleReconnectingState is the logic for the Reconnecting state.
func (c *OBSClient) handleReconnectingState() {
	interval := time.Duration(c.config.ReconnectInterval) * time.Second
	c.logger.Debug("State: Reconnecting. Waiting before next attempt...", "interval", interval)

	select {
	case <-time.After(interval):
		c.setState(StateDisconnected)
	case <-c.ctx.Done():
		// Shutdown requested during reconnect wait.
	}
}

// ============================================================================
// FSM HELPER IMPLEMENTATIONS
// ============================================================================

// setState safely transitions the client to a new state and signals the FSM.
func (c *OBSClient) setState(newState State) {
	c.stateMu.Lock()
	oldState := c.state
	if oldState == newState {
		c.stateMu.Unlock()
		return
	}
	c.state = newState
	c.stateMu.Unlock()

	c.logger.Debug("Client state transitioned", "from", oldState, "to", newState)
	c.signalFSM()
}

// signalFSM sends a non-blocking signal to the FSM channel to wake it up.
func (c *OBSClient) signalFSM() {
	select {
	case c.signalCh <- struct{}{}:
	default: // Channel buffer is already full, signal is pending.
	}
}

// ============================================================================
// PUBLIC STATE ACCESSORS
// ============================================================================

// GetState returns the current state of the client in a thread-safe manner.
func (c *OBSClient) GetState() State {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}
