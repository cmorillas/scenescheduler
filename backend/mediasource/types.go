// backend/mediasource/types.go
//
// This file defines core data structures for the mediasource Manager,
// including custom types, enums, errors, and proxies.
//
// Contents:
// - Constants
// - State Machine Types
// - Custom Errors
// - Track Protection Proxy

package mediasource

import (
	"errors"
	"fmt"
	"time"

	"github.com/pion/webrtc/v4"
)

// ============================================================================
// Constants
// ============================================================================

// Timeout for feed release operations.
const releaseTimeout = 5 * time.Second

// ============================================================================
// State Machine Types
// ============================================================================

// State represents a state in the Manager's FSM.
type State int

const (
	StateInactive State = iota
	StateAcquiring
	StateActive
	StateReleasing
	StateFailed
)

// String provides a human-readable name for a State.
func (s State) String() string {
	switch s {
	case StateInactive:
		return "Inactive"
	case StateAcquiring:
		return "Acquiring"
	case StateActive:
		return "Active"
	case StateReleasing:
		return "Releasing"
	case StateFailed:
		return "Failed"
	default:
		return fmt.Sprintf("UnknownState(%d)", int(s))
	}
}

// ============================================================================
// Custom Errors
// ============================================================================

var (
	ErrNotActive         = errors.New("mediasource: not active")
	ErrAcquisitionFailed = errors.New("mediasource: acquisition failed")
)

// ============================================================================
// Track Protection Proxy
// ============================================================================

// TrackProxy wraps a webrtc.TrackLocal to prevent external consumers
// from closing the Manager's shared tracks.
type TrackProxy struct {
	webrtc.TrackLocal
}

// Close overrides the embedded track's Close method to be a no-op.
func (p *TrackProxy) Close() error {
	return nil // Intentionally prevent external closure
}

// OnEnded overrides the embedded track's OnEnded method to be a no-op.
func (p *TrackProxy) OnEnded(f func(error)) {
	// Intentionally prevent external event handler manipulation
}

