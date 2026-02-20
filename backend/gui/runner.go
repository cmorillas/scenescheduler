// backend/gui/runner.go
//
// Lifecycle orchestration for the GUI module.
//
// Contents:
// - Public lifecycle methods (Run, Stop)
// - Internal lifecycle helpers (cleanup)

package gui

import (
	"fyne.io/fyne/v2"
)

// =============================================================================
// PUBLIC LIFECYCLE METHODS
// =============================================================================

// Run starts the main GUI event loop and associated goroutines.
// It blocks until the GUI is closed.
//
// This method:
// - Sets up the window close handler
// - Starts the clock update goroutine
// - Monitors context cancellation
// - Blocks on Fyne's ShowAndRun() loop
func (g *GUI) Run() {
	g.logInfo("GUI runner started.")

	defer g.cleanup()

	// Set the window close handler to cancel the GUI's context
	g.window.SetOnClosed(func() {
		g.logInfo("Window close event detected, stopping GUI.")
		g.Stop()
	})

	// Start clock update goroutine
	g.startClock()

	// Goroutine to monitor context cancellation and ensure the Fyne app quits
	go func() {
		<-g.ctx.Done()
		g.logInfo("GUI context cancelled. Signalling Fyne app to quit.")
		fyne.Do(func() {
			if g.fyneApp != nil {
				g.fyneApp.Quit()
			}
		})
	}()

	g.logInfo("Starting Fyne's blocking ShowAndRun() loop.")
	g.window.ShowAndRun()
	g.logInfo("Fyne's ShowAndRun loop has finished.")
}

// Stop gracefully shuts down the GUI by cancelling its context.
// This method is idempotent and can be called multiple times safely.
func (g *GUI) Stop() {
	g.stopOnce.Do(func() {
		g.logInfo("Stop requested for GUI.")
		if g.cancelCtx != nil {
			g.cancelCtx()
		}
	})
}

// =============================================================================
// INTERNAL LIFECYCLE HELPERS
// =============================================================================

// cleanup is the internal finalizer for the module.
// It orchestrates the graceful shutdown of all module resources.
// This method is idempotent and can be called multiple times safely.
func (g *GUI) cleanup() {
	g.cleanupOnce.Do(func() {
		g.logInfo("Cleaning up GUI resources.")
		g.unsubscribeAllEvents()
	})
}