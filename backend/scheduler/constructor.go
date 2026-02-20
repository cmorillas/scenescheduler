// backend/scheduler/constructor.go
//
// Scheduler module constructor and configuration.
//
// Contents:
// - Scheduler Struct Definition
// - Constructor (New)

package scheduler

import (
	"context"
	"sync"

	"scenescheduler/backend/config"
	"scenescheduler/backend/eventbus"
	"scenescheduler/backend/logger"
)

// ============================================================================
// SCHEDULER STRUCT DEFINITION
// ============================================================================

// Scheduler evaluates schedule.json and publishes target program state.
// It does not control OBS directly, only declares desired state.
type Scheduler struct {
	// --- Configuration ---
	logger *logger.Logger
	bus    *eventbus.EventBus
	paths  *config.PathsConfig
	config *config.SchedulerConfig

	// --- Lifecycle Management ---
	ctx       context.Context
	cancelCtx context.CancelFunc

	// --- Idempotency Protection ---
	stopOnce    sync.Once
	cleanupOnce sync.Once

	// --- Event Subscriptions ---
	unsubscribeFuncs []func()

	// --- Internal State (protected by mutex) ---
	mu       sync.RWMutex
	schedule *Schedule // Current loaded schedule

	// --- Internal Components ---
	fileWatcher *fileWatcher // Watches schedule.json for changes
}

// ============================================================================
// CONSTRUCTOR
// ============================================================================

// New creates a new Scheduler instance with the provided dependencies.
// The scheduler is immediately ready to receive events after this returns.
//
// Parameters:
//   - appCtx: Parent context for lifecycle management
//   - log: Logger instance
//   - pathsCfg: File paths configuration
//   - schedulerCfg: Scheduler-specific configuration
//   - bus: EventBus for inter-module communication
//
// Returns:
//   - *Scheduler: Configured Scheduler instance ready to Run()
//
// CRITICAL: This constructor subscribes to events before returning to prevent
// race conditions where events could arrive before subscription is complete.
func New(
	appCtx context.Context,
	log *logger.Logger,
	pathsCfg *config.PathsConfig,
	schedulerCfg *config.SchedulerConfig,
	bus *eventbus.EventBus,
) *Scheduler {
	s := &Scheduler{
		logger:           log.WithModule("scheduler"),
		bus:              bus,
		paths:            pathsCfg,
		config:           schedulerCfg,
		unsubscribeFuncs: make([]func(), 0),
	}

	// Create derived context for this module's lifecycle
	s.ctx, s.cancelCtx = context.WithCancel(appCtx)

	// CRITICAL: Subscribe to events before returning to prevent race conditions
	// Handlers can now safely use s.ctx
	s.subscribeToEvents()

	return s
}