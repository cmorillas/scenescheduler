// backend/scheduler/helpers.go
//
// Stateless helper functions for the Scheduler module.
// These are pure functions without receiver that can be reused across the module.
//
// Contents:
// - Constants and Maps
// - Program Lookup Helpers
// - Program Time Helpers
// - Miscellaneous Helpers

package scheduler

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// CONSTANTS AND MAPS
// ============================================================================

const (
	DefaultProgramID = "default-source"
	NoProgramTitle   = "<none>"

	isoFormat  = "2006-01-02T15:04:05Z" // RFC3339 format for UTC
	dateFormat = "2006-01-02"
)

var weekDaysMap = map[string]time.Weekday{
	"SUN": time.Sunday,
	"MON": time.Monday,
	"TUE": time.Tuesday,
	"WED": time.Wednesday,
	"THU": time.Thursday,
	"FRI": time.Friday,
	"SAT": time.Saturday,
}

// ============================================================================
// PROGRAM LOOKUP HELPERS
// ============================================================================

// findProgramAtTime returns the program active at the given time `t`,
// correctly handling both single and recurring events. For recurring events,
// it treats the stored times as LOCAL time templates, ignoring timezone information.
func findProgramAtTime(programs []ScheduledProgram, t time.Time) *ScheduledProgram {
	// Work in local time for all comparisons
	tLocal := t.Local()

	for i := range programs {
		p := &programs[i]
		if !p.Enabled {
			continue
		}

		if p.Timing.IsRecurring {
			// For recurring events: check today and yesterday (for overnight events)
			for dayOffset := 0; dayOffset >= -1; dayOffset-- {
				checkDay := tLocal.AddDate(0, 0, dayOffset)

				// 1. Check if recurrence rule is active for this date
				checkDayStr := checkDay.Format(dateFormat)
				if p.Timing.Recurrence.StartRecur != "" && checkDayStr < p.Timing.Recurrence.StartRecur {
					continue
				}
				if p.Timing.Recurrence.EndRecur != "" && checkDayStr > p.Timing.Recurrence.EndRecur {
					continue
				}

				// 2. Check if day of week matches
				dayMatches := false
				checkWeekday := checkDay.Weekday()
				for _, dayStr := range p.Timing.Recurrence.DaysOfWeek {
					if mappedDay, ok := weekDaysMap[strings.ToUpper(dayStr)]; ok && mappedDay == checkWeekday {
						dayMatches = true
						break
					}
				}
				if !dayMatches {
					continue
				}

				// 3. Extract ONLY the time part from the templates, ignore date and timezone
				templateStart := p.Timing.Start
				templateEnd := p.Timing.End
				if templateStart.IsZero() || templateEnd.IsZero() {
					continue
				}

				// Build local times using only the H:M:S from templates
				eventStart := time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(),
					templateStart.Hour(), templateStart.Minute(), templateStart.Second(), 0, time.Local)

				eventEnd := time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(),
					templateEnd.Hour(), templateEnd.Minute(), templateEnd.Second(), 0, time.Local)

				// Handle overnight events (end time is before or equal to start)
				if eventEnd.Before(eventStart) || eventEnd.Equal(eventStart) {
					eventEnd = eventEnd.Add(24 * time.Hour)
				}

				// 4. Check if current time is within [eventStart, eventEnd)
				if (tLocal.After(eventStart) || tLocal.Equal(eventStart)) && tLocal.Before(eventEnd) {
					return p
				}
			}
		} else {
			// Non-recurring: use absolute timestamps in local time
			start := p.Timing.Start.Local()
			end := p.Timing.End.Local()
			if !start.IsZero() && !end.IsZero() &&
				(tLocal.After(start) || tLocal.Equal(start)) && tLocal.Before(end) {
				return p
			}
		}
	}
	return nil
}

// findNextProgramAfter returns the first program that starts after the given time.
// For recurring events, it calculates the next occurrence within a 7-day lookahead window.
func findNextProgramAfter(programs []ScheduledProgram, after time.Time) *ScheduledProgram {
	afterLocal := after.Local()
	var next *ScheduledProgram
	var nextStartTime time.Time

	for i := range programs {
		p := &programs[i]
		if !p.Enabled {
			continue
		}

		if p.Timing.IsRecurring {
			// For recurring events: find the next occurrence in the next 7 days
			candidateStart := findNextRecurrenceAfter(p, afterLocal)
			if !candidateStart.IsZero() {
				if next == nil || candidateStart.Before(nextStartTime) {
					next = p
					nextStartTime = candidateStart
				}
			}
		} else {
			// Non-recurring: simple timestamp comparison
			start := p.Timing.Start.Local()
			if !start.IsZero() && start.After(afterLocal) {
				if next == nil || start.Before(nextStartTime) {
					next = p
					nextStartTime = start
				}
			}
		}
	}
	return next
}

// findNextRecurrenceAfter finds the next occurrence of a recurring program after the given time.
// It checks up to 7 days ahead to find the next matching day and time.
func findNextRecurrenceAfter(p *ScheduledProgram, after time.Time) time.Time {
	if p.Timing.Start.IsZero() {
		return time.Time{}
	}

	// Extract the time template (H:M:S only)
	templateStart := p.Timing.Start

	// Check the next 7 days for a matching occurrence
	for dayOffset := 0; dayOffset <= 7; dayOffset++ {
		checkDay := after.AddDate(0, 0, dayOffset)

		// Check if recurrence rule is active for this date
		checkDayStr := checkDay.Format(dateFormat)
		if p.Timing.Recurrence.StartRecur != "" && checkDayStr < p.Timing.Recurrence.StartRecur {
			continue
		}
		if p.Timing.Recurrence.EndRecur != "" && checkDayStr > p.Timing.Recurrence.EndRecur {
			continue
		}

		// Check if this weekday is in the recurrence pattern
		checkWeekday := checkDay.Weekday()
		dayMatches := false
		for _, dayStr := range p.Timing.Recurrence.DaysOfWeek {
			if mappedDay, ok := weekDaysMap[strings.ToUpper(dayStr)]; ok && mappedDay == checkWeekday {
				dayMatches = true
				break
			}
		}
		if !dayMatches {
			continue
		}

		// Build the candidate start time using the template time
		candidateStart := time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(),
			templateStart.Hour(), templateStart.Minute(), templateStart.Second(), 0, time.Local)

		// This is a valid occurrence if it's after the reference time
		if candidateStart.After(after) {
			return candidateStart
		}
	}

	return time.Time{} // No occurrence found in the next 7 days
}

// ============================================================================
// PROGRAM TIME HELPERS
// ============================================================================

// getProgramStartTime returns the effective start time of a program for a given day in local time.
func getProgramStartTime(p *ScheduledProgram, now time.Time) time.Time {
	if p == nil || p.Timing.Start.IsZero() {
		return time.Time{}
	}
	nowLocal := now.Local()

	if p.Timing.IsRecurring {
		templateStart := p.Timing.Start
		return time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
			templateStart.Hour(), templateStart.Minute(), templateStart.Second(), 0, time.Local)
	}
	// For non-recurring events, the start time is absolute in local time.
	return p.Timing.Start.Local()
}

// getProgramEndTime returns the effective end time of a program for a given day in local time.
func getProgramEndTime(p *ScheduledProgram, now time.Time) time.Time {
	if p == nil || p.Timing.Start.IsZero() || p.Timing.End.IsZero() {
		return time.Time{}
	}
	nowLocal := now.Local()

	if p.Timing.IsRecurring {
		templateStart := p.Timing.Start
		templateEnd := p.Timing.End

		todayEnd := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(),
			templateEnd.Hour(), templateEnd.Minute(), templateEnd.Second(), 0, time.Local)

		// Handle overnight programs (e.g., starts 22:00, ends 02:00)
		if templateEnd.Before(templateStart) || templateEnd.Equal(templateStart) {
			todayEnd = todayEnd.Add(24 * time.Hour)
		}
		return todayEnd
	}

	// For non-recurring events, the end time is absolute in local time.
	return p.Timing.End.Local()
}

// ============================================================================
// MISCELLANEOUS HELPERS
// ============================================================================

// isDefaultSource returns true if the given program represents the default source.
func isDefaultSource(p *ScheduledProgram) bool {
	return p != nil && p.ID == DefaultProgramID
}

// getProgramTitle returns the title of a program, or a placeholder if nil.
func getProgramTitle(p *ScheduledProgram) string {
	if p == nil {
		return NoProgramTitle
	}
	if p.Title != "" {
		return p.Title
	}
	return fmt.Sprintf("Untitled (%s)", p.ID)
}