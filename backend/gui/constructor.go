// backend/gui/constructor.go
//
// GUI module constructor, configuration, and public API.
//
// Contents:
// - Type definitions
// - Constructor (New) - calls subscribeToEvents()
// - Public API methods
// - Internal logging helpers

package gui

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"scenescheduler/backend/eventbus"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// GUI manages the graphical user interface for the Scene Scheduler.
// It provides a desktop window with real-time status monitoring,
// program schedule display, and activity logging.
// The GUI operates as an observer - it does not control the backend,
// only displays its state.
type GUI struct {
	// --- Core Dependencies ---
	eventBus *eventbus.EventBus // EventBus for module communication

	// --- Fyne Core ---
	fyneApp fyne.App    // Fyne application instance
	window  fyne.Window // Main application window

	// --- Lifecycle Management ---
	ctx              context.Context         // Module context for lifecycle management
	cancelCtx        context.CancelFunc      // Context cancellation function
	stopOnce         sync.Once               // Ensures Stop() executes only once
	cleanupOnce      sync.Once               // Ensures cleanup() executes only once
	unsubscribeFuncs []eventbus.UnsubscribeFunc // Event subscriptions to clean up

	// --- UI Widgets (Direct References) ---
	connectionLabel        *widget.Label // OBS connection status display
	clockLabel             *widget.Label // Current time display
	webServerStatusLabel   *widget.Label // Web server status display
	webServerUsersLabel    *widget.Label // WebSocket client count display
	livePreviewStatusLabel *widget.Label // Live preview/media source status
	livePreviewUsersLabel  *widget.Label // WebRTC connection count display
	currentProgramCard     *widget.Card  // Current program information card
	nextProgramCard        *widget.Card  // Next program information card
	logListWidget          *widget.List  // Activity log list widget

	// --- Data Binding (Only for Dynamic List) ---
	logListBinding binding.StringList // Binding for dynamic log messages

	// --- Internal State (Protected by Mutex) ---
	mu                   sync.RWMutex // Protects program tracking state
	lastCurrentProgramID string       // Last displayed current program ID
	lastNextProgramID    string       // Last displayed next program ID
}

// =============================================================================
// CONSTRUCTOR
// =============================================================================

// New creates and initializes a new GUI instance.
// It sets up the window, widgets, layout, and event subscriptions.
// The GUI is immediately ready to receive events after this returns.
//
// Parameters:
//   - appCtx: Parent context for lifecycle management
//   - eventBus: EventBus instance for inter-module communication
//
// Returns:
//   - *GUI: Configured GUI instance ready to Run()
func New(appCtx context.Context, eventBus *eventbus.EventBus) *GUI {
	// Create Fyne application and window
	fyneAppInstance := app.New()
	fyneAppInstance.Settings().SetTheme(newMyTheme())
	mainWin := fyneAppInstance.NewWindow("Scene Scheduler")

	// Create GUI instance
	g := &GUI{
		eventBus:         eventBus,
		fyneApp:          fyneAppInstance,
		window:           mainWin,
		unsubscribeFuncs: make([]eventbus.UnsubscribeFunc, 0),
		logListBinding:   binding.NewStringList(),
	}

	// Create derived context for GUI lifecycle
	g.ctx, g.cancelCtx = context.WithCancel(appCtx)

	// Initialize UI widgets with default values
	g.initializeWidgets()

	// Build and set the UI layout
	uiContent := g.buildLayout()
	g.window.SetContent(uiContent)
	g.window.Resize(fyne.NewSize(800, 600))
	g.window.CenterOnScreen()
	g.window.SetMaster()

	// Subscribe to events before returning
	// This ensures no events are missed during startup
	g.subscribeToEvents()

	return g
}

// =============================================================================
// PUBLIC API METHODS
// =============================================================================

// AddLogMessage adds a message to the log list in the UI.
// This method is specifically designed to be called from the central logger.
// The operation is thread-safe and causes the log view to scroll to the bottom.
//
// Parameters:
//   - message: Log message to display in the activity log
func (g *GUI) AddLogMessage(message string) {
	if g.logListBinding == nil {
		return // Guard against calls during shutdown
	}
	fyne.Do(func() {
		_ = g.logListBinding.Append(message)
	})
}

// =============================================================================
// INTERNAL LOGGING HELPERS
// =============================================================================

// logInfo prints an informational message to stdout.
// This is for internal GUI module logging only.
// The main application logger uses AddLogMessage.
func (g *GUI) logInfo(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stdout, "[%s] INFO [gui]: %s\n", timestamp, formattedMsg)
}

// logError prints an error message to stderr.
// This is for internal GUI module logging only.
func (g *GUI) logError(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stderr, "[%s] ERROR [gui]: %s\n", timestamp, formattedMsg)
}