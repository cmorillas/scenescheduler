// backend/gui/helpers.go
//
// Stateless utility functions for the GUI module.
// These functions have no receiver and do not depend on GUI state.
//
// Contents:
// - UI helpers

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// =============================================================================
// UI HELPERS
// =============================================================================

// newSpacer creates a transparent rectangle to act as a spacer in layouts.
// Used to add visual spacing between UI components without visible elements.
//
// Parameters:
//   - width: Minimum width of the spacer in pixels
//   - height: Minimum height of the spacer in pixels
//
// Returns a canvas object that can be added to containers for spacing.
func newSpacer(width, height float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(width, height))
	return rect
}