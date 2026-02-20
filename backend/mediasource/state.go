// backend/mediasource/state.go
//
// This file contains thread-safe methods for accessing and modifying the
// internal state of the Manager, including the FSM state and the active feed.
//
// Contents:
// - State Getters and Setters
// - Feed Getters and Setters

package mediasource

import "scenescheduler/backend/mediasource/internal/feed"

// ============================================================================
// State Getters and Setters
// ============================================================================

// getState returns the current FSM state in a thread-safe manner.
func (m *Manager) getState() State {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()
	return m.state
}

// setState updates the FSM state and notifies the runner loop.
func (m *Manager) setState(s State) {
	m.stateMu.Lock()
	oldState := m.state
	if oldState == s {
		m.stateMu.Unlock()
		return
	}
	m.state = s
	m.stateMu.Unlock()

	m.logger.Debug("State transition", "from", oldState, "to", s)
	m.signalStateChange()
}

// signalStateChange notifies the runner loop that a state change has occurred.
// This is non-blocking.
func (m *Manager) signalStateChange() {
	select {
	case m.stateSignal <- struct{}{}:
	default: // Signal already pending
	}
}

// ============================================================================
// Feed Getters and Setters
// ============================================================================

// getFeed returns the current active feed in a thread-safe manner.
func (m *Manager) getFeed() *feed.Feed {
	m.feedMu.RLock()
	defer m.feedMu.RUnlock()
	return m.feed
}

// setFeed sets the current active feed in a thread-safe manner.
func (m *Manager) setFeed(f *feed.Feed) {
	m.feedMu.Lock()
	defer m.feedMu.Unlock()
	m.feed = f
}

// getAndClearFeed atomically retrieves the current feed and sets the internal
// reference to nil. This prevents race conditions during feed release.
func (m *Manager) getAndClearFeed() *feed.Feed {
	m.feedMu.Lock()
	defer m.feedMu.Unlock()
	f := m.feed
	m.feed = nil
	return f
}
