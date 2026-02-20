// backend/obsclient/helpers.go
//
// This file contains stateless helper functions for the OBSClient module.
//
// Contents:
// - Logging and Display Helpers

package obsclient

import (
	"fmt"
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// LOGGING AND DISPLAY HELPERS
// ============================================================================

// getProgramTitle returns a display-friendly name for a program data object,
// primarily used for logging.
func getProgramTitle(p *eventbus.Program) string {
	if p == nil {
		return "<none>"
	}
	if p.Title != "" {
		return p.Title
	}
	return fmt.Sprintf("Untitled (%s)", p.ID)
}
