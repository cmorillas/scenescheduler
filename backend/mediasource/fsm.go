// backend/mediasource/fsm.go
//
// This file contains the Finite State Machine (FSM) logic for the Manager.
// It includes the main state dispatcher and the implementation for each
// state handler.
//
// Contents:
// - State Dispatcher
// - State Handler Implementations

package mediasource

import (
	"context"
)

// ============================================================================
// State Dispatcher
// ============================================================================

// handleState is the main dispatcher for the FSM. It executes the appropriate
// logic based on the current state.
func (m *Manager) handleState(state State) {
	switch state {
	case StateInactive:
		// Passive state, no action needed.
	case StateAcquiring:
		m.doHandleAcquiring()
	case StateActive:
		m.doHandleActive()
	case StateReleasing:
		m.doHandleReleasing()
	case StateFailed:
		m.doHandleFailed()
	}
}

// ============================================================================
// State Handler Implementations
// ============================================================================

// doHandleAcquiring attempts to acquire media and transitions state accordingly.
func (m *Manager) doHandleAcquiring() {
	m.logger.Debug("FSM: Handling Acquiring state")

	// Release any stale feed before acquiring a new one.
	m.releaseActiveFeed(context.Background(), "stale feed cleanup")

	err := m.performAcquisition()
	if err != nil {
		m.logger.Error("Acquisition failed", "error", err)
		m.publishMediaAcquireFailed(err.Error())
		m.setState(StateFailed)
		return
	}

	m.logger.Debug("Acquisition successful")
	m.publishMediaReady()
	m.setState(StateActive)
}

// doHandleActive starts monitoring the acquired feed.
func (m *Manager) doHandleActive() {
	m.logger.Debug("FSM: Handling Active state, starting feed monitoring")
	feed := m.getFeed()
	if feed == nil {
		m.logger.Error("FSM bug: Active state entered with nil feed")
		m.setState(StateFailed)
		return
	}

	feed.StartMonitoring(func(reason string) {
		m.logger.Error("Feed reported total failure", "reason", reason)
		m.publishMediaLost(reason)
		m.setState(StateReleasing)
	})
}

// doHandleReleasing performs graceful cleanup and transitions to Inactive.
func (m *Manager) doHandleReleasing() {
	m.logger.Debug("FSM: Handling Releasing state")

	ctx, cancel := context.WithTimeout(context.Background(), releaseTimeout)
	defer cancel()

	m.releaseActiveFeed(ctx, "state transition to releasing")
	m.setState(StateInactive)
}

// doHandleFailed performs cleanup and parks the FSM.
func (m *Manager) doHandleFailed() {
	m.logger.Debug("FSM: Handling Failed state, cleaning up")

	ctx, cancel := context.WithTimeout(context.Background(), releaseTimeout)
	defer cancel()

	m.releaseActiveFeed(ctx, "failure cleanup")
	// The FSM remains in the Failed state, awaiting external intervention.
}
