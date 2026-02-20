// backend/scheduler/runner.go
//
// Lifecycle orchestration for the Scheduler module.
//
// Contents:
// - Public Lifecycle Methods (Run, Stop)
// - Internal Lifecycle Helpers

package scheduler

import (
	"time"
)

// ============================================================================
// PUBLIC LIFECYCLE METHODS
// ============================================================================

// Run starts the Scheduler's main evaluation loop.
// It initializes components, loads the schedule, and runs periodic evaluation
// until the context is canceled or Stop() is called.
// The context for this module already exists from the constructor.
//
// The evaluation loop runs every second to check if the target program has changed.
func (s *Scheduler) Run() {
	defer s.cleanup()

	s.logger.Info("Scheduler Runner starting")

	// Initialize default source from configuration
	s.setupDefaultSource()

	// Load initial schedule from disk
	s.reloadSchedule()

	// Start file watcher for hot-reload
	s.initFileWatcher()

	// Create ticker for periodic evaluation (every second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Main evaluation loop
	for {
		select {
		case <-ticker.C:
			s.evaluateAndSwitch()

		case <-s.ctx.Done():
			s.logger.InfoGui("Scheduler context canceled, stopping")
			return
		}
	}
}

// Stop gracefully stops the Scheduler by canceling its context.
// This method is idempotent and can be called multiple times safely.
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		s.logger.InfoGui("Stop requested for Scheduler")
		if s.cancelCtx != nil {
			s.cancelCtx()
		}
	})
}

// ============================================================================
// INTERNAL LIFECYCLE HELPERS
// ============================================================================

// cleanup releases resources and unsubscribes from all event bus topics.
// This method is idempotent and guaranteed to run only once.
func (s *Scheduler) cleanup() {
	s.cleanupOnce.Do(func() {
		s.logger.Debug("Cleaning up Scheduler resources")

		// Stop file watcher
		if s.fileWatcher != nil {
			s.fileWatcher.stop()
		}

		// Unsubscribe from all events
		s.unsubscribeAllEvents()
	})
}