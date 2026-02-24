// backend/webserver/internal/websocket/handler.go
//
// WebSocket handler: manages connection lifecycle and message routing.
//
// This handler is an internal component that follows the Ambassador pattern.
// It has no knowledge of the EventBus and is controlled entirely by its parent
// (the WebServer module) through direct method calls.
//
// Contents:
// - Struct Definition
// - Constructor
// - HTTP Handler
// - Connection Management
// - Message Routing
// - Public API (for parent control)

package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"scenescheduler/backend/logger"

	"github.com/gorilla/websocket"
)

// Disconnect reason constants
const (
	ReasonGracefulShutdown = "graceful_shutdown"
	ReasonClientClosed     = "client_closed"
)

// =============================================================================
// Struct Definition
// =============================================================================

// Handler manages the mechanics of WebSocket connections: lifecycle,
// read/write pumps, and connection registration.
type Handler struct {
	// --- Internal State ---
	connections map[string]*WSConnection
	mu          sync.RWMutex

	// --- Dependencies ---
	logger    *logger.Logger
	callbacks Callbacks
	upgrader  websocket.Upgrader
}

// =============================================================================
// Constructor
// =============================================================================

// New creates a new WebSocket handler instance.
// The handler does NOT subscribe to any events - it is controlled by its parent
// via the provided callbacks.
func New(log *logger.Logger, callbacks Callbacks) *Handler {
	return &Handler{
		connections: make(map[string]*WSConnection),
		logger:      log.WithModule("websocket"),
		callbacks:   callbacks,
		upgrader: websocket.Upgrader{
			CheckOrigin:     func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// =============================================================================
// HTTP Handler
// =============================================================================

// HandleConnection upgrades an HTTP connection and starts the connection pumps.
func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", "error", err, "remoteAddr", r.RemoteAddr)
		return
	}

	wsConn := newWSConnection(conn, r, h)
	h.register(wsConn)

	// Launch pumps - readPump blocks until connection closes
	go wsConn.writePump()
	wsConn.readPump()
}

// =============================================================================
// Connection Management
// =============================================================================

// register adds a connection to the handler's managed pool.
func (h *Handler) register(conn *WSConnection) {
	h.mu.Lock()
	h.connections[conn.ID] = conn
	newCount := len(h.connections)
	h.mu.Unlock()

	h.logger.Info("WebSocket connected", "id", conn.ID, "ip", conn.IP, "total", newCount)

	// Notify parent via callbacks
	if h.callbacks.OnClientConnected != nil {
		h.callbacks.OnClientConnected(conn.ID, conn.IP, conn.UserAgent)
	}
	if h.callbacks.OnClientsChanged != nil {
		h.callbacks.OnClientsChanged(newCount)
	}
}

// unregister removes a connection from the pool and closes its resources.
// Safe to call multiple times (idempotent).
func (h *Handler) unregister(conn *WSConnection, reason string) {
	h.mu.Lock()
	if _, ok := h.connections[conn.ID]; !ok {
		h.mu.Unlock()
		return
	}

	delete(h.connections, conn.ID)
	newCount := len(h.connections)
	close(conn.send)
	_ = conn.conn.Close()
	h.mu.Unlock()

	h.logger.Info("WebSocket disconnected", "id", conn.ID, "ip", conn.IP, "reason", reason, "total", newCount)

	// Notify parent via callbacks
	if h.callbacks.OnClientDisconnected != nil {
		h.callbacks.OnClientDisconnected(conn.ID, conn.IP, conn.UserAgent, reason)
	}
	if h.callbacks.OnClientsChanged != nil {
		h.callbacks.OnClientsChanged(newCount)
	}
}

// CloseAll disconnects all connections, typically during graceful shutdown.
func (h *Handler) CloseAll() {
	h.mu.Lock()
	connectionsToClose := make([]*WSConnection, 0, len(h.connections))
	for _, conn := range h.connections {
		connectionsToClose = append(connectionsToClose, conn)
	}
	h.mu.Unlock()

	h.logger.Debug("Closing all WebSocket connections", "count", len(connectionsToClose))
	for _, conn := range connectionsToClose {
		h.unregister(conn, ReasonGracefulShutdown)
	}
}

// =============================================================================
// Message Routing
// =============================================================================

// routeMessage processes an incoming message and notifies parent via callbacks.
// The parent (WebServer) will then translate these to EventBus events.
func (h *Handler) routeMessage(msg *Message, connID string) {
	switch msg.Action {
	case "getSchedule":
		h.logger.Debug("Routing 'getSchedule' command", "connID", connID)
		if h.callbacks.OnGetSchedule != nil {
			h.callbacks.OnGetSchedule(connID)
		}

	case "commitSchedule":
		h.logger.Debug("Routing 'commitSchedule' command", "connID", connID)
		if h.callbacks.OnCommitSchedule != nil {
			h.callbacks.OnCommitSchedule(connID, msg.Payload)
		}

	case "getStatus":
		h.logger.Debug("Routing 'getStatus' command", "connID", connID)
		if h.callbacks.OnGetStatus != nil {
			h.callbacks.OnGetStatus(connID)
		}

	case "startPreview":
		h.logger.Debug("Routing 'startPreview' command", "connID", connID)
		// Get connection to extract remoteAddr
		h.mu.RLock()
		conn, exists := h.connections[connID]
		h.mu.RUnlock()

		if exists && h.callbacks.OnStartPreview != nil {
			// Use IP as remoteAddr for preview tracking
			h.callbacks.OnStartPreview(connID, conn.IP, msg.Payload)
		}

	case "stopPreview":
		h.logger.Debug("Routing 'stopPreview' command", "connID", connID)
		// Get connection to extract remoteAddr
		h.mu.RLock()
		conn, exists := h.connections[connID]
		h.mu.RUnlock()

		if exists && h.callbacks.OnStopPreview != nil {
			// Use IP as remoteAddr for preview tracking
			h.callbacks.OnStopPreview(connID, conn.IP)
		}

	default:
		h.logger.Warn("Unknown message action", "action", msg.Action, "connID", connID)
	}
}

// =============================================================================
// Public API (for parent control)
// =============================================================================

// SendToClient sends a message to a specific client connection.
// This is called by the parent WebServer in response to EventBus events.
func (h *Handler) SendToClient(clientID, messageType string, payload interface{}) {
	h.mu.RLock()
	conn, exists := h.connections[clientID]
	h.mu.RUnlock()

	if !exists {
		h.logger.Warn("Attempted to send to non-existent connection", "connID", clientID)
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal payload", "error", err, "connID", clientID)
		return
	}

	msg := &Message{Action: messageType, Payload: json.RawMessage(payloadBytes)}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal message", "error", err, "connID", clientID)
		return
	}

	// Protect against send-on-closed-channel panic.
	// Between releasing RLock above and sending below, unregister() may close conn.send.
	defer func() {
		if r := recover(); r != nil {
			h.logger.Warn("Send failed (connection closed concurrently)", "connID", clientID)
		}
	}()

	select {
	case conn.send <- msgBytes:
	default:
		h.logger.Warn("Connection send buffer full, dropping message", "connID", conn.ID)
	}
}

// Broadcast sends a message to all connected clients.
// This is called by the parent WebServer in response to EventBus events.
func (h *Handler) Broadcast(messageType string, payload json.RawMessage) {
	msg := &Message{Action: messageType, Payload: payload}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal broadcast message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conn := range h.connections {
		func(c *WSConnection) {
			defer func() {
				if r := recover(); r != nil {
					h.logger.Warn("Broadcast failed (connection closed concurrently)", "connID", c.ID)
				}
			}()
			select {
			case c.send <- msgBytes:
			default:
				h.logger.Warn("Connection send buffer full, dropping broadcast", "connID", c.ID)
			}
		}(conn)
	}
}
