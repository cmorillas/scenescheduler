// backend/gui/theme.go
//
// Custom visual theme for the Fyne application.
//
// Contents:
// - Theme struct and constructor
// - Color overrides

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// =============================================================================
// THEME STRUCT AND CONSTRUCTOR
// =============================================================================

// myTheme is a custom dark theme for the Scene Scheduler GUI.
// It extends Fyne's default dark theme with custom color choices
// for better visibility and aesthetic appeal.
type myTheme struct {
	fyne.Theme
}

// newMyTheme creates and returns a new instance of the custom theme.
// The theme is based on Fyne's DarkTheme with color overrides.
func newMyTheme() fyne.Theme {
	return &myTheme{Theme: theme.DarkTheme()}
}

// =============================================================================
// COLOR OVERRIDES
// =============================================================================

// Color returns the color for a given theme color name and variant.
// It overrides specific colors from the base dark theme while delegating
// all other color requests to the underlying theme.
func (t *myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground:
		// Text color - slightly brighter than default for better readability
		return color.NRGBA{R: 210, G: 210, B: 210, A: 255}

	case theme.ColorNameBackground:
		// General background - dark gray
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}

	case theme.ColorNameButton:
		// Button color when not focused - blue accent
		return color.NRGBA{R: 12, G: 110, B: 253, A: 255}

	case theme.ColorNameHover:
		// Hover color for buttons and labels - slightly darker blue
		return color.NRGBA{R: 11, G: 94, B: 215, A: 255}

	case theme.ColorNameSelection:
		// Selection color for lists and text - medium gray
		return color.NRGBA{R: 100, G: 100, B: 100, A: 255}

	case theme.ColorNameSeparator:
		// Separator lines between UI elements - matches background for subtle separation
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}

	default:
		// Delegate all other colors to the base dark theme
		return t.Theme.Color(name, variant)
	}
}