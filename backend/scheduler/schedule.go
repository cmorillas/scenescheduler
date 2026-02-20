// backend/scheduler/schedule.go
//
// Schedule management and lifecycle operations.
//
// Contents:
// - Schedule Reloading
// - Default Source Setup

package scheduler

// ============================================================================
// SCHEDULE RELOADING
// ============================================================================

// reloadSchedule reloads the schedule from disk and triggers evaluation.
// This is called on startup and when the FileWatcher detects changes.
func (s *Scheduler) reloadSchedule() {
	s.logger.InfoGui("Reloading schedule from file", "path", s.paths.Schedule)

	newSchedule, err := s.loadScheduleFromFile()
	if err != nil {
		s.logger.Error("Failed to reload schedule, keeping existing schedule active", "error", err)
		return
	}

	s.mu.Lock()
	s.schedule = newSchedule
	s.mu.Unlock()

	s.logger.InfoGui("Successfully reloaded schedule into memory", "program_count", len(newSchedule.Programs))

	// Always trigger an immediate evaluation after reload
	s.evaluateAndSwitch()
}

// ============================================================================
// DEFAULT SOURCE SETUP
// ============================================================================

// setupDefaultSource prepares the default source from configuration.
// This is called once during startup in Run().
func (s *Scheduler) setupDefaultSource() {
	if s.config.DefaultSource.Name == "" {
		s.logger.Debug("No default source configured")
		return
	}

	s.logger.Debug("Default source configured",
		"name", s.config.DefaultSource.Name,
		"inputKind", s.config.DefaultSource.InputKind)

	// The default source is converted to a Program on-demand in defaultSourceToProgram()
	// during evaluation, so no additional setup is needed here.
}