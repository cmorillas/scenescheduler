// backend/eventbus/events_scheduler.go
package eventbus

import "time"


// =============================================================================
// Scheduler State Events
// =============================================================================

// TargetProgramState is published by the Scheduler every evaluation cycle.
// It declares the desired state: which program should be visible at the current moment.
// The Scheduler is stateless and always publishes the current target, regardless of
// whether it represents a change. Consuming modules decide if action is needed.
type TargetProgramState struct {
	Timestamp      time.Time
	TargetProgram  *Program
	NextProgram    *Program
	SeekOffset     time.Duration
}

// GetTopic returns the unique topic identifier for this event.
func (e TargetProgramState) GetTopic() string { return "scheduler.state.targetProgram" }