// backend/obsclient/types.go
//
// This file defines the core data structures, enums, and errors for the
// OBSClient module.
//
// Contents:
// - Module-level Errors
// - Finite State Machine (FSM) Types
// - Connection Session Types

package obsclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreykaipov/goobs"
)

// ============================================================================
// MODULE-LEVEL ERRORS
// ============================================================================

// ErrNotConnected is returned when an operation requires a connection to OBS,
// but the client is not in the 'Connected' state.
var ErrNotConnected = errors.New("obsclient: not connected")

// ============================================================================
// FINITE STATE MACHINE (FSM) TYPES
// ============================================================================

// State represents the different operational states of the OBS client's FSM.
type State int

const (
	// StateDisconnected is the initial and final state.
	StateDisconnected State = iota
	// StateConnecting indicates an active connection attempt.
	StateConnecting
	// StateConnected indicates a stable, active connection.
	StateConnected
	// StateReconnecting indicates a lost connection and a pending retry.
	StateReconnecting
)

// String provides a human-readable representation of the State.
func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateReconnecting:
		return "Reconnecting"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// ============================================================================
// CONNECTION SESSION TYPES
// ============================================================================

// connection holds all resources for a single, active connection session.
type connection struct {
	client     *goobs.Client
	ctx        context.Context
	cancelCtx  context.CancelFunc
	obsVersion string
}
