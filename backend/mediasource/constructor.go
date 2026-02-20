// backend/mediasource/constructor.go
//
// This file defines the Manager struct for the mediasource module, its
// constructor (New), and all module-level state.
//
// Contents:
// - Manager Struct Definition
// - Constructor (New)

package mediasource

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/mediadevices"
	"scenescheduler/backend/config"
	"scenescheduler/backend/eventbus"
	"scenescheduler/backend/logger"
	"scenescheduler/backend/mediasource/internal/feed"

	// Drivers for media device discovery
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
)

// ============================================================================
// Manager Struct Definition
// ============================================================================

// Manager orchestrates the media source lifecycle using a Finite State Machine.
// It coordinates device acquisition, active monitoring, and graceful cleanup.
type Manager struct {
	// --- Dependencies (immutable) ---
	logger        *logger.Logger
	config        *config.MediaSourceConfig
	eventBus      *eventbus.EventBus
	codecSelector *mediadevices.CodecSelector

	// --- FSM State (guarded by stateMu) ---
	stateMu     sync.RWMutex
	state       State
	stateSignal chan struct{}

	// --- Active Feed (guarded by feedMu) ---
	feedMu sync.RWMutex
	feed   *feed.Feed

	// --- Lifecycle Management ---
	ctx              context.Context
	cancelCtx        context.CancelFunc
	stopOnce         sync.Once
	cleanupOnce      sync.Once
	unsubscribeFuncs []eventbus.UnsubscribeFunc
}

// ============================================================================
// Constructor (New)
// ============================================================================

// New creates a new mediasource Manager.
// The manager is immediately ready to receive events after this returns.
// It expands the config, creates the codec selector, and subscribes to events.
// Panics if codec selector creation fails (unrecoverable configuration error).
//
// Parameters:
//   - appCtx: Parent context for lifecycle management
//   - log: Logger instance
//   - cfg: Media source configuration
//   - bus: EventBus for inter-module communication
//
// Returns:
//   - *Manager: Configured Manager instance ready to Run()
func New(appCtx context.Context, log *logger.Logger, cfg *config.MediaSourceConfig, bus *eventbus.EventBus) *Manager {
	log = log.WithModule("mediasource")

	// Expand config with quality profile parameters
	applyQualityProfile(cfg)

	// Create codec selector with quality-tuned parameters
	codecSelector, err := createCodecSelector(cfg)
	if err != nil {
		log.Error("Failed to create codec selector, panicking", "error", err)
		panic(fmt.Sprintf("mediasource: failed to create codec selector: %v", err))
	}

	m := &Manager{
		logger:        log,
		config:        cfg,
		eventBus:      bus,
		codecSelector: codecSelector,
		state:         StateInactive,
		stateSignal:   make(chan struct{}, 1),
	}

	// Create derived context for this module's lifecycle
	m.ctx, m.cancelCtx = context.WithCancel(appCtx)

	// CRITICAL: Subscribe to events before returning
	// Handlers can now safely use m.ctx
	m.subscribeToEvents()

	return m
}