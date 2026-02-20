// backend/webserver/internal/websocket/connection.go
//
// WebSocket connection representation and pump goroutines.

package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Constants for WebSocket timing and limits.
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 16384 // 16 KB
)

// =============================================================================
// WSConnection Definition
// =============================================================================

// WSConnection represents a single WebSocket connection with its metadata
// and communication channels.
type WSConnection struct {
	ID        string
	IP        string
	UserAgent string

	conn    *websocket.Conn
	handler *Handler
	send    chan []byte
}

// Message defines the standard WebSocket message structure.
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// =============================================================================
// Constructor
// =============================================================================

// newWSConnection creates a new connection instance from an HTTP request.
func newWSConnection(conn *websocket.Conn, r *http.Request, handler *Handler) *WSConnection {
	return &WSConnection{
		ID:        generateID(),
		IP:        getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		conn:      conn,
		handler:   handler,
		send:      make(chan []byte, 256),
	}
}

// =============================================================================
// Pump Goroutines
// =============================================================================

// writePump pumps messages from the send channel to the WebSocket connection.
// A single goroutine runs writePump for each connection.
func (c *WSConnection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Handler closed the channel
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.handler.logger.Warn("Failed to write message", "connID", c.ID, "error", err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.handler.logger.Warn("Failed to send ping", "connID", c.ID, "error", err)
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the router.
// It blocks until the connection is closed, ensuring proper cleanup via defer.
func (c *WSConnection) readPump() {
	defer func() {
		c.handler.unregister(c, ReasonClientClosed)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.handler.logger.Warn("Unexpected close error", "connID", c.ID, "error", err)
			}
			break
		}

		c.handler.routeMessage(&msg, c.ID)
	}
}

// =============================================================================
// Helpers
// =============================================================================

// generateID creates a cryptographically secure random ID.
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random ID: %v", err))
	}
	return hex.EncodeToString(bytes)
}

// getClientIP retrieves the real client IP from an HTTP request.
func getClientIP(r *http.Request) string {
	headersToCheck := []string{
		"CF-Connecting-IP",
		"X-Real-IP",
		"X-Forwarded-For",
	}

	for _, header := range headersToCheck {
		ipStr := r.Header.Get(header)
		if ipStr == "" {
			continue
		}

		if header == "X-Forwarded-For" {
			parts := strings.Split(ipStr, ",")
			for _, part := range parts {
				ip := strings.TrimSpace(part)
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		} else {
			ip := strings.TrimSpace(ipStr)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	if ip == "::1" {
		return "127.0.0.1"
	}

	return ip
}