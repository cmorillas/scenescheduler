// backend/mediasource/feed.go
//
// This file contains methods for managing the internal feed component.
// It acts as a bridge between the Manager's FSM and the feed's lifecycle.
//
// Contents:
// - Feed Lifecycle Methods
// - FSM Trigger Methods

package mediasource

import (
	"context"
	"fmt"
	"time"

	"scenescheduler/backend/mediasource/internal/feed"
)

const acquisitionTimeout = 10 * time.Second

// ============================================================================
// Feed Lifecycle Methods
// ============================================================================

// performAcquisition creates and acquires a new feed.
func (m *Manager) performAcquisition() error {
	m.logger.Debug("Creating and acquiring new feed")

	newFeed := feed.New(*m.config, m.codecSelector, m.logger)

	if err := newFeed.Acquire(m.ctx); err != nil {
		return fmt.Errorf("feed acquisition failed: %w", err)
	}

	m.setFeed(newFeed)
	m.logger.InfoGui("Feed acquired successfully",
		"videoDevice", newFeed.GetVideoDeviceName(),
		"audioDevice", newFeed.GetAudioDeviceName())

	return nil
}

// releaseActiveFeed safely releases the current feed. This is idempotent.
func (m *Manager) releaseActiveFeed(ctx context.Context, reason string) {
	activeFeed := m.getAndClearFeed()
	if activeFeed == nil {
		return
	}

	m.logger.Debug("Releasing active feed", "reason", reason)

	timeout := releaseTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	if err := activeFeed.Release(timeout); err != nil {
		m.logger.Error("Error during feed release", "error", err)
	} else {
		m.logger.Debug("Feed released successfully")
	}

	m.publishMediaStopped(reason)
}

// ============================================================================
// FSM Trigger Methods
// ============================================================================

// requestAcquisition initiates a device acquisition attempt by changing state.
func (m *Manager) requestAcquisition() {
	currentState := m.getState()
	switch currentState {
	case StateActive, StateAcquiring:
		m.logger.Debug("Ignoring acquisition request, already busy", "state", currentState)
	case StateInactive, StateFailed:
		m.logger.Debug("Acquisition requested, transitioning to Acquiring", "fromState", currentState)
		m.setState(StateAcquiring)
	case StateReleasing:
		m.logger.Debug("Acquisition requested while releasing, will re-trigger later")
	}
}

// requestRelease initiates graceful release of the active feed by changing state.
func (m *Manager) requestRelease(reason string) {
	m.logger.InfoGui("Release requested", "reason", reason)
	currentState := m.getState()

	switch currentState {
	case StateActive, StateAcquiring:
		m.setState(StateReleasing)
	case StateFailed:
		m.logger.Debug("Releasing from failed state")
		m.setState(StateReleasing)
	default:
		m.logger.Debug("Ignoring release request, not active", "state", currentState)
	}
}
