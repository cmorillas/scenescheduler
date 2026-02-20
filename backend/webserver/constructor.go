// backend/webserver/constructor.go
//
// This file defines the WebServer module's struct and its constructor.
// It is responsible for initializing the module and all its dependencies,
// including internal handlers for WebSocket and WHEP.
//
// Contents:
// - Struct Definition
// - Constructor (New)

package webserver

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"scenescheduler/backend/config"
	"scenescheduler/backend/eventbus"
	"scenescheduler/backend/logger"
	"scenescheduler/backend/webserver/internal/sourcepreview"
	"scenescheduler/backend/webserver/internal/websocket"
	"scenescheduler/backend/webserver/internal/whep"
)

// =============================================================================
// Struct Definition
// =============================================================================

// WebServer orchestrates the HTTP server, WebSocket gateway, and WHEP streaming.
// It acts as the primary bridge between browser clients and the backend system.
type WebServer struct {
	// --- Dependencies ---
	config *config.WebServerConfig
	logger *logger.Logger
	bus    *eventbus.EventBus

	// --- Internal Components ---
	httpServer        *http.Server
	wsHandler         *websocket.Handler
	whepHandler       *whep.Handler
	previewManager    *sourcepreview.Manager

	// --- Lifecycle Management ---
	ctx              context.Context
	cancel           context.CancelFunc
	stopOnce         sync.Once
	cleanupOnce      sync.Once
	unsubscribeFuncs []eventbus.UnsubscribeFunc
}

// =============================================================================
// Constructor
// =============================================================================

// New creates a new, fully initialized WebServer instance.
// The server is immediately ready to receive events after this returns.
// It sets up internal handlers, configures HTTP routes, and subscribes to
// necessary events before returning, ensuring the module is ready on creation.
//
// Parameters:
//   - appCtx: Parent context for lifecycle management
//   - log: Logger instance
//   - cfg: Web server configuration
//   - bus: EventBus for inter-module communication
//   - staticFiles: Embedded filesystem for static frontend files
//
// Returns:
//   - *WebServer: Configured WebServer instance ready to Run()
func New(
	appCtx context.Context,
	log *logger.Logger,
	cfg *config.WebServerConfig,
	bus *eventbus.EventBus,
	staticFiles fs.FS,
) *WebServer {
	log = log.WithModule("webserver")

	// Create the WebServer instance.
	server := &WebServer{
		config:           cfg,
		logger:           log,
		bus:              bus,
		unsubscribeFuncs: make([]eventbus.UnsubscribeFunc, 0, 4),
	}

	// Create derived context for this module's lifecycle
	server.ctx, server.cancel = context.WithCancel(appCtx)

	// Create internal handlers, injecting dependencies.
	// Both WHEP and WebSocket handlers use callbacks to notify of events,
	// keeping them decoupled from the EventBus.
	server.wsHandler = websocket.New(log, server.createWebSocketCallbacks(bus))
	server.whepHandler = whep.New(log, func(count int) {
		eventbus.Publish(bus, eventbus.WebRTCConnectionsChanged{
			Timestamp: time.Now(),
			Count:     count,
		})
	})

	// Create preview manager (may fail non-fatally if binary not found)
	previewMgr, err := sourcepreview.New(log, cfg.HlsPath)
	if err != nil {
		log.Error("Failed to create preview manager", "error", err)
		// Continue anyway - preview requests will fail gracefully
	}
	server.previewManager = previewMgr

	// Setup HTTP routing and middleware.
	mux := server.setupRouter(staticFiles)
	server.httpServer = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// CRITICAL: Subscribe to events before returning to prevent race conditions.
	// Handlers can now safely use server.ctx
	server.subscribeToEvents()

	return server
}

// createWebSocketCallbacks creates the callback functions for the WebSocket handler.
// This is the Ambassador pattern - the internal handler notifies the parent (WebServer)
// via callbacks, and the parent translates these to EventBus events.
func (ws *WebServer) createWebSocketCallbacks(bus *eventbus.EventBus) websocket.Callbacks {
	return websocket.Callbacks{
		// Connection lifecycle callbacks
		OnClientConnected: func(clientID, ip, userAgent string) {
			eventbus.Publish(bus, eventbus.WebSocketClientConnected{
				ClientID:  clientID,
				IP:        ip,
				UserAgent: userAgent,
				Timestamp: time.Now(),
			})
		},
		OnClientDisconnected: func(clientID, ip, userAgent, reason string) {
			// Cleanup any active previews for this client
			ws.handleClientDisconnected(clientID, ip, userAgent, reason)

			// Publish disconnect event to EventBus
			eventbus.Publish(bus, eventbus.WebSocketClientDisconnected{
				ClientID:  clientID,
				IP:        ip,
				UserAgent: userAgent,
				Reason:    reason,
				Timestamp: time.Now(),
			})
		},
		OnClientsChanged: func(count int) {
			eventbus.Publish(bus, eventbus.WebSocketClientsChanged{
				Count:     count,
				Timestamp: time.Now(),
			})
		},

		// Message routing callbacks (client requests)
		OnGetSchedule: func(clientID string) {
			eventbus.Publish(bus, eventbus.GetScheduleRequested{
				ClientID: clientID,
			})
		},
		OnCommitSchedule: func(clientID string, payload json.RawMessage) {
			eventbus.Publish(bus, eventbus.CommitScheduleRequested{
				ClientID: clientID,
				Payload:  payload,
			})
		},
		OnGetStatus: func(clientID string) {
			eventbus.Publish(bus, eventbus.GetStatusRequested{
				ClientID: clientID,
			})
		},

		// Source preview callbacks
		OnStartPreview: func(clientID, remoteAddr string, payload json.RawMessage) {
			ws.handleStartPreview(clientID, remoteAddr, payload)
		},
		OnStopPreview: func(clientID, remoteAddr string) {
			ws.handleStopPreview(clientID, remoteAddr)
		},
	}
}