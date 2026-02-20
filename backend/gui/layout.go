// backend/gui/layout.go
//
// Visual layout construction for the GUI.
//
// Contents:
// - Main layout builder
// - Layout component helpers

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// =============================================================================
// MAIN LAYOUT BUILDER
// =============================================================================

// buildLayout constructs the complete visual hierarchy of the GUI.
// Called once from the constructor after widgets are initialized.
//
// Layout structure:
// - Top section: Status rows (OBS, WebServer, LivePreview)
// - Middle section: Program panels (Current, Next)
// - Bottom section: Activity log list
//
// Returns the root container ready to be set as window content.
func (g *GUI) buildLayout() fyne.CanvasObject {
	// --- Connection Status Row ---
	connectionLabel := widget.NewLabelWithStyle("OBS Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	connectionHBox := container.NewHBox(
		connectionLabel,
		g.connectionLabel,
		layout.NewSpacer(),
		g.clockLabel,
	)

	// --- Web Server Status Row ---
	webServerStatusLabel := widget.NewLabelWithStyle("Web Server:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	webServerUsersLabel := widget.NewLabelWithStyle("WebSockets:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	webServerHBox := container.NewHBox(
		webServerStatusLabel,
		g.webServerStatusLabel,
		layout.NewSpacer(),
		webServerUsersLabel,
		g.webServerUsersLabel,
	)

	// --- Live Preview Status Row ---
	livePreviewStatusLabel := widget.NewLabelWithStyle("Live Preview:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	livePreviewUsersLabel := widget.NewLabelWithStyle("WebRTCs:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	livePreviewHBox := container.NewHBox(
		livePreviewStatusLabel,
		g.livePreviewStatusLabel,
		layout.NewSpacer(),
		livePreviewUsersLabel,
		g.livePreviewUsersLabel,
	)

	// --- Program Panels Section ---
	programsTitle := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Current Program", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Next Program", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	programsPanels := container.NewGridWithColumns(2, g.currentProgramCard, g.nextProgramCard)

	// --- Activity Log Section ---
	logsTitle := widget.NewLabelWithStyle("Activity Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	g.logListWidget = widget.NewListWithData(
		g.logListBinding,
		g.createLogItemTemplate,
		g.updateLogItemFromData,
	)

	// Auto-scroll to bottom when new log entries are added
	g.logListBinding.AddListener(binding.NewDataListener(func() {
		fyne.Do(func() {
			if g.logListWidget != nil {
				g.logListWidget.ScrollToBottom()
			}
		})
	}))

	logCard := widget.NewCard("", "", g.logListWidget)

	// --- Compose Final Layout ---
	topLayout := container.NewVBox(
		connectionHBox,
		webServerHBox,
		livePreviewHBox,
		newSpacer(0, theme.Padding()*2),
		programsTitle,
		programsPanels,
		newSpacer(0, theme.Padding()),
		logsTitle,
	)

	mainLayout := container.NewBorder(topLayout, nil, nil, nil, logCard)
	paddedLayout := container.NewPadded(mainLayout)

	return paddedLayout
}

// =============================================================================
// LAYOUT COMPONENT HELPERS
// =============================================================================

// createLogItemTemplate creates a template widget for log list items.
// This is called by Fyne's list widget to create reusable item templates.
// Uses canvas.Text instead of widget.Label for compact spacing.
func (g *GUI) createLogItemTemplate() fyne.CanvasObject {
	return canvas.NewText("", theme.ForegroundColor())
}

// updateLogItemFromData updates a log list item widget with data from the binding.
// This is called by Fyne's list widget when an item needs to be displayed.
func (g *GUI) updateLogItemFromData(item binding.DataItem, obj fyne.CanvasObject) {
	str, _ := item.(binding.String).Get()
	obj.(*canvas.Text).Text = str
	obj.Refresh()
}