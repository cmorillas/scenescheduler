// backend/mediasource/runner.go
//
// This file defines the lifecycle orchestration for the Manager. It provides
// the entry points to start and stop the service, runs the FSM loop, and
// manages cleanup at shutdown.
//
// Contents:
// - Public Lifecycle Methods (Run, Stop)
// - Internal Lifecycle Helpers (cleanup)

package mediasource

import (
	"context"
	"time"
)

// Timeout for final cleanup operations
const cleanupTimeout = 5 * time.Second

// ============================================================================
// Public Lifecycle Methods
// ============================================================================

// Run starts the Manager's lifecycle. It initializes the FSM and blocks
// until the context is cancelled.
// The context for this module already exists from the constructor.
func (m *Manager) Run() {
	m.logger.Debug("MediaSource Manager Runner service starting.")

	// Trigger initial state processing.
	m.signalStateChange()

	// Run the FSM loop (blocks until shutdown).
	m.logger.Debug("FSM loop started.", "initialState", m.getState())
	defer m.cleanup()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Debug("FSM loop stopping.")
			return
		case <-m.stateSignal:
			state := m.getState()
			m.logger.Debug("FSM processing state", "state", state)
			m.handleState(state)
		}
	}
}

// Stop gracefully shuts down the manager by cancelling its context.
// This method is idempotent and can be called multiple times safely.
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		m.logger.Info("Requesting to stop MediaSource Manager service.")
		if m.cancelCtx != nil {
			m.cancelCtx()
		}
	})
}

// ============================================================================
// Internal Lifecycle Helpers
// ============================================================================

// cleanup is called once when Run() exits. It ensures all resources are released.
func (m *Manager) cleanup() {
	m.cleanupOnce.Do(func() {
		m.logger.Debug("Cleaning up MediaSource Manager resources.")
		m.unsubscribeAllEvents()

		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()
		m.releaseActiveFeed(ctx, "manager shutdown")
	})
}