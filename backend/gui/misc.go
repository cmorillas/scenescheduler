// backend/gui/misc.go
//
// Private implementation methods for GUI operations.
//
// Contents:
// - Initialization helpers
// - Lifecycle methods
// - UI update methods
// - State checking methods

package gui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"

	"scenescheduler/backend/eventbus"
)

// =============================================================================
// INITIALIZATION HELPERS
// =============================================================================

// initializeWidgets creates and initializes all UI widgets with default values.
// Called from constructor to set up the widget references.
func (g *GUI) initializeWidgets() {
	g.connectionLabel = widget.NewLabel("Disconnected")
	g.clockLabel = widget.NewLabel("--:--:--")
	g.clockLabel.TextStyle = fyne.TextStyle{Monospace: true}
	g.webServerStatusLabel = widget.NewLabel("Off")
	g.webServerUsersLabel = widget.NewLabel("-")
	g.livePreviewStatusLabel = widget.NewLabel("Off")
	g.livePreviewUsersLabel = widget.NewLabel("-")

	// Program cards will be created in buildLayout with initial content
	g.currentProgramCard = g.createProgramCard("N/A", "--:--:--", "--:--:--")
	g.nextProgramCard = g.createProgramCard("N/A", "--:--:--", "--:--:--")

	// Log list widget will be created in buildLayout
}

// =============================================================================
// LIFECYCLE METHODS
// =============================================================================

// startClock launches a goroutine to update the GUI's clock every second.
// The goroutine runs until the GUI's context is cancelled.
func (g *GUI) startClock() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		// Immediately set the current time
		fyne.Do(func() {
			g.clockLabel.SetText(time.Now().Format("15:04:05"))
		})

		for {
			select {
			case <-ticker.C:
				fyne.Do(func() {
					g.clockLabel.SetText(time.Now().Format("15:04:05"))
				})
			case <-g.ctx.Done():
				g.logInfo("Clock update goroutine stopping.")
				return
			}
		}
	}()
}

// =============================================================================
// UI UPDATE METHODS - STATUS LABELS
// =============================================================================

// updateConnectionStatus updates the OBS connection status label.
// Thread-safe operation using fyne.Do().
func (g *GUI) updateConnectionStatus(status string) {
	fyne.Do(func() {
		g.connectionLabel.SetText(status)
	})
}

// updateWebServerStatus updates the web server status label.
// Thread-safe operation using fyne.Do().
func (g *GUI) updateWebServerStatus(status string) {
	fyne.Do(func() {
		g.webServerStatusLabel.SetText(status)
	})
}

// updateWebServerUsers updates the WebSocket client count label.
// Thread-safe operation using fyne.Do().
func (g *GUI) updateWebServerUsers(count string) {
	fyne.Do(func() {
		g.webServerUsersLabel.SetText(count)
	})
}

// updateLivePreviewStatus updates the live preview/media source status label.
// Thread-safe operation using fyne.Do().
func (g *GUI) updateLivePreviewStatus(status string) {
	fyne.Do(func() {
		g.livePreviewStatusLabel.SetText(status)
	})
}

// updateLivePreviewUsers updates the WebRTC connection count label.
// Thread-safe operation using fyne.Do().
func (g *GUI) updateLivePreviewUsers(count string) {
	fyne.Do(func() {
		g.livePreviewUsersLabel.SetText(count)
	})
}

// =============================================================================
// UI UPDATE METHODS - PROGRAM PANELS
// =============================================================================

// shouldUpdateProgramPanels checks if the program panels need updating.
// Compares the new program IDs with the last displayed IDs to avoid
// unnecessary UI redraws.
//
// Returns true if either the current or next program has changed.
func (g *GUI) shouldUpdateProgramPanels(event eventbus.TargetProgramState) bool {
	var currentID, nextID string
	if event.TargetProgram != nil {
		currentID = event.TargetProgram.ID
	}
	if event.NextProgram != nil {
		nextID = event.NextProgram.ID
	}

	g.mu.RLock()
	lastCurrentID := g.lastCurrentProgramID
	lastNextID := g.lastNextProgramID
	g.mu.RUnlock()

	return currentID != lastCurrentID || nextID != lastNextID
}

// updateProgramPanels updates both the current and next program display cards.
// Updates internal state tracking to reflect the new programs.
// Thread-safe operation using fyne.Do() and mutex protection.
func (g *GUI) updateProgramPanels(current, next *eventbus.Program) {
	fyne.Do(func() {
		// Rebuild current program card
		g.currentProgramCard.SetContent(g.buildProgramCardContent(current))

		// Rebuild next program card
		g.nextProgramCard.SetContent(g.buildProgramCardContent(next))
	})

	// Update tracked state
	g.mu.Lock()
	if current != nil {
		g.lastCurrentProgramID = current.ID
	} else {
		g.lastCurrentProgramID = ""
	}
	if next != nil {
		g.lastNextProgramID = next.ID
	} else {
		g.lastNextProgramID = ""
	}
	g.mu.Unlock()
}

// =============================================================================
// INTERNAL HELPERS - PROGRAM CARDS
// =============================================================================

// createProgramCard creates a new program card widget with initial values.
// Used during initialization to create the placeholder cards.
func (g *GUI) createProgramCard(title, start, end string) *widget.Card {
	content := g.buildProgramCardContentFromStrings(title, start, end)
	return widget.NewCard("", "", content)
}

// buildProgramCardContent builds the content for a program card from ProgramData.
// Returns a container with the program title and timing information.
// If program is nil, returns placeholder content.
func (g *GUI) buildProgramCardContent(program *eventbus.Program) fyne.CanvasObject {
	if program == nil {
		return g.buildProgramCardContentFromStrings("N/A", "--:--:--", "--:--:--")
	}

	title := program.Title
	if title == "" {
		title = program.SourceName // Fallback to source name if title is empty
	}

	startTime := "--:--:--"
	if !program.Start.IsZero() {
		startTime = program.Start.Format("15:04:05")
	}

	endTime := "--:--:--"
	if !program.End.IsZero() {
		endTime = program.End.Format("15:04:05")
	}

	return g.buildProgramCardContentFromStrings(title, startTime, endTime)
}

// buildProgramCardContentFromStrings builds the content for a program card from strings.
// Creates a container with formatted title and timing labels.
func (g *GUI) buildProgramCardContentFromStrings(title, start, end string) fyne.CanvasObject {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	startLabel := widget.NewLabel(start)
	endLabel := widget.NewLabel(end)

	form := widget.NewForm(
		widget.NewFormItem("Start:", startLabel),
		widget.NewFormItem("End:", endLabel),
	)

	return container.NewVBox(titleLabel, form)
}