// backend/scheduler/types.go
//
// Core data structures for the Scheduler module.
//
// Contents:
// - Schedule Types
// - Program Types
// - Timing and Recurrence Types
// - Behavior Types

package scheduler

import "time"

// ============================================================================
// SCHEDULE TYPES
// ============================================================================

// Schedule is the root object representing a full scheduling configuration.
// It maps directly to the schedule.json file.
type Schedule struct {
	Version      string             `json:"version"`      // Schema version
	ScheduleName string             `json:"scheduleName"` // Human-readable name of the schedule
	Programs     []ScheduledProgram `json:"schedule"`     // List of programs (events)
}

// ============================================================================
// PROGRAM TYPES
// ============================================================================

// ScheduledProgram defines a single scheduled event, including metadata, source, timing,
// and behavior configuration. This is the internal domain model for the scheduler.
type ScheduledProgram struct {
	ID       string   `json:"id"`      // Unique identifier for the program
	Title    string   `json:"title"`   // Human-readable title displayed in UI
	Enabled  bool     `json:"enabled"` // Whether the program is active and should be scheduled
	General  General  `json:"general"` // Visual metadata for frontend display
	Source   Source   `json:"source"`  // OBS input source configuration
	Timing   Timing   `json:"timing"`  // Scheduling details
	Behavior Behavior `json:"behavior"`// Runtime behavior at start/end
}

// General stores metadata for program visualization in the frontend calendar.
type General struct {
	Description     string   `json:"description"`     // Free-text description of the program
	Tags            []string `json:"tags"`            // Keywords for searching/categorization
	ClassNames      []string `json:"classNames"`      // Custom CSS classes for styling
	TextColor       string   `json:"textColor"`       // Calendar text color
	BackgroundColor string   `json:"backgroundColor"` // Calendar background color
	BorderColor     string   `json:"borderColor"`     // Calendar border color
}

// Source defines an OBS input that should be activated during the program.
type Source struct {
	Name          string                 `json:"name"`          // Technical input name in OBS
	InputKind     string                 `json:"inputKind"`     // OBS input type (e.g. browser_source, ffmpeg_source)
	URI           string                 `json:"uri"`           // Path or URL for the source
	InputSettings map[string]interface{} `json:"inputSettings"` // Type-specific OBS input settings
	Transform     map[string]interface{} `json:"transform"`     // Transform properties (position, size, crop)
}

// ============================================================================
// TIMING AND RECURRENCE TYPES
// ============================================================================

// Timing defines when the program should run, either once or recurrently.
type Timing struct {
	Start       time.Time  `json:"start"`       // ISO 8601 format for single events
	End         time.Time  `json:"end"`         // ISO 8601 format for single events
	IsRecurring bool       `json:"isRecurring"` // Whether the program repeats
	Recurrence  Recurrence `json:"recurrence"`  // Recurrence rule if repeating
}

// Recurrence defines the rule for repeating programs.
type Recurrence struct {
	DaysOfWeek []string `json:"daysOfWeek"` // e.g., ["MON", "WED", "FRI"]
	StartRecur string   `json:"startRecur"` // YYYY-MM-DD
	EndRecur   string   `json:"endRecur"`   // YYYY-MM-DD
}

// ============================================================================
// BEHAVIOR TYPES
// ============================================================================

// Behavior defines how the program should behave during and after execution.
type Behavior struct {
	OnEndAction    string `json:"onEndAction"`    // Action after program ends (hide, none, stop)
	PreloadSeconds int    `json:"preloadSeconds"` // Seconds before start to preload the source
}