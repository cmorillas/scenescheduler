// backend/scheduler/evaluation.go
//
// Core scheduling evaluation logic.
//
// Contents:
// - Main Evaluation Method
// - Target State Publishing
// - Default Source Handling
// - Helper Methods

package scheduler

import (
	"time"

	"scenescheduler/backend/eventbus"
)

// ============================================================================
// MAIN EVALUATION METHOD
// ============================================================================

// evaluateAndSwitch contains the core scheduling logic executed on every tick.
// It is stateless: it always publishes the current desired state without
// comparing to previous state. The responsibility of deciding whether to act
// has been moved to the OBSClient.
func (s *Scheduler) evaluateAndSwitch() {
	now := time.Now()

	s.mu.RLock()
	currentSchedule := s.schedule
	s.mu.RUnlock()

	var targetProgram *ScheduledProgram
	if currentSchedule != nil {
		targetProgram = findProgramAtTime(currentSchedule.Programs, now)
	}

	// If no scheduled program is active and a default source is configured, use it
	if targetProgram == nil && s.config.DefaultSource.Name != "" {
		targetProgram = s.defaultSourceToProgram()
	}

	// Calculate the next program for informational purposes
	var nextProgram *ScheduledProgram
	if currentSchedule != nil {
		searchStartTime := now
		if targetProgram != nil && !isDefaultSource(targetProgram) {
			searchStartTime = getProgramEndTime(targetProgram, now)
		}
		if !searchStartTime.IsZero() {
			nextProgram = findNextProgramAfter(currentSchedule.Programs, searchStartTime)
		}
	}

	// Calculate seek offset for media that can be seeked
	var seekOffset time.Duration
	if targetProgram != nil && !isDefaultSource(targetProgram) {
		startTime := getProgramStartTime(targetProgram, now)
		if now.After(startTime) {
			seekOffset = now.Sub(startTime)
		}
	}

	// Always publish the desired state. The OBSClient will decide if action is needed.
	eventbus.Publish(s.bus, eventbus.TargetProgramState{
		Timestamp:     now,
		TargetProgram: toExecutableProgram(targetProgram),
		NextProgram:   toExecutableProgram(nextProgram),
		SeekOffset:    seekOffset,
	})
}

// ============================================================================
// TARGET STATE PUBLISHING
// ============================================================================

// toExecutableProgram translates a ScheduledProgram (internal domain model)
// to an eventbus.Program (DTO for execution). This adapter is the anti-corruption
// layer between the scheduler's internal representation and the public event contract.
func toExecutableProgram(p *ScheduledProgram) *eventbus.Program {
	if p == nil {
		return nil
	}
	return &eventbus.Program{
		ID:            p.ID,
		Title:         p.Title,
		SourceName:    p.Source.Name,
		InputKind:     p.Source.InputKind,
		URI:           p.Source.URI,
		InputSettings: p.Source.InputSettings,
		Transform:     p.Source.Transform,
		Start:         p.Timing.Start,
		End:           p.Timing.End,
	}
}

// ============================================================================
// DEFAULT SOURCE HANDLING
// ============================================================================

// defaultSourceToProgram converts the DefaultSource config into a ScheduledProgram struct.
func (s *Scheduler) defaultSourceToProgram() *ScheduledProgram {
	// Assert dynamic types from config to concrete maps when possible.
	var inputSettings map[string]interface{}
	if m, ok := s.config.DefaultSource.InputSettings.(map[string]interface{}); ok {
		inputSettings = m
	}
	var transform map[string]interface{}
	if m, ok := s.config.DefaultSource.Transform.(map[string]interface{}); ok {
		transform = m
	}

	return &ScheduledProgram{
		ID:      DefaultProgramID,
		Title:   "Default Source",
		Enabled: true,
		Source: Source{
			Name:          s.config.DefaultSource.Name,
			InputKind:     s.config.DefaultSource.InputKind,
			URI:           s.config.DefaultSource.URI,
			InputSettings: inputSettings,
			Transform:     transform,
		},
	}
}