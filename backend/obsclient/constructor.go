// backend/obsclient/constructor.go
//
// This file defines the OBSClient struct, its state, and the constructor. It is
// the only file where the module's state should be defined.
//
// Contents:
// - OBSClient Struct Definition
// - Constructor (New)

package obsclient

import (
	"context"
	"sync"

	"scenescheduler/backend/config"
	"scenescheduler/backend/eventbus"
	"scenescheduler/backend/logger"
	"scenescheduler/backend/obsclient/internal/switcher"
)

// ============================================================================
// OBSCLIENT STRUCT DEFINITION
// ============================================================================

// OBSClient manages the connection and lifecycle with OBS. It acts as a high-
// level orchestrator, delegating complex tasks to specialized internal components.
type OBSClient struct {
	// --- Configuration and Dependencies ---
	logger *logger.Logger
	config *config.OBSConfig
	bus    *eventbus.EventBus

	// --- Internal Components ---
	switcher *switcher.Switcher

	// --- Lifecycle Management ---
	ctx              context.Context
	cancelCtx        context.CancelFunc
	signalCh         chan struct{}
	unsubscribeFuncs []func()
	stopOnce         sync.Once
	cleanupOnce      sync.Once

	// --- Synchronization ---
	stateMu  sync.RWMutex // Protects state, connection, and activeProgram
	switchMu sync.Mutex   // Serializes all convergence operations to prevent races

	// --- Internal State (protected by stateMu) ---
	state         State
	connection    *connection
	activeProgram *eventbus.Program // Holds the currently active program
}

// ============================================================================
// CONSTRUCTOR
// ============================================================================

// New creates and configures a new OBS client module and its internal components.
// The client is immediately ready to receive events after this returns.
//
// Parameters:
//   - appCtx: Parent context for lifecycle management
//   - log: Logger instance
//   - cfg: OBS connection configuration
//   - bus: EventBus for inter-module communication
//
// Returns:
//   - *OBSClient: Configured OBSClient instance ready to Run()
func New(appCtx context.Context, log *logger.Logger, cfg *config.OBSConfig, bus *eventbus.EventBus) *OBSClient {
	c := &OBSClient{
		logger:           log.WithModule("obsclient"),
		config:           cfg,
		bus:              bus,
		switcher:         switcher.New(log, cfg),
		signalCh:         make(chan struct{}, 1),
		unsubscribeFuncs: make([]func(), 0),
		state:            StateDisconnected,
		activeProgram:    nil, // Starts with no active program
	}

	// Create derived context for this module's lifecycle
	c.ctx, c.cancelCtx = context.WithCancel(appCtx)

	// Subscribe to events immediately upon creation to prevent race conditions.
	// Handlers can now safely use c.ctx
	c.subscribeToEvents()

	return c
}