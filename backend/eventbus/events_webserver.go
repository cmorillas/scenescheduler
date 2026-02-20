// backend/eventbus/events_webserver.go
package eventbus

import (
    "encoding/json"
    "time"
)

// Constants for standard reason fields to ensure consistency across the application.
const (
    ReasonGracefulShutdown = "graceful_shutdown"
    ReasonServerError      = "server_error"
    ReasonClientClosed     = "client_closed"
    ReasonConnectionError  = "connection_error"
)

// =============================================================================
// WebServer Lifecycle Events
// =============================================================================

// WebServerStarted is published when the HTTP server begins listening for connections.
type WebServerStarted struct {
    Port      string
    UseTLS    bool
    IPs       []string
    Timestamp time.Time
}

func (e WebServerStarted) GetTopic() string { return "webserver.lifecycle.started" }

// WebServerStopped is published after the HTTP server has successfully shut down.
type WebServerStopped struct {
    Reason    string
    Timestamp time.Time
}

func (e WebServerStopped) GetTopic() string { return "webserver.lifecycle.stopped" }

// =============================================================================
// WebSocket Events (emitted by WebServer)
// =============================================================================

// WebSocketClientConnected is published when a new client establishes a WebSocket connection.
type WebSocketClientConnected struct {
    ClientID  string
    IP        string
    UserAgent string
    Timestamp time.Time
}

func (e WebSocketClientConnected) GetTopic() string { return "webserver.websocket.clientConnected" }

// WebSocketClientDisconnected is published when a client's WebSocket connection is terminated.
type WebSocketClientDisconnected struct {
    ClientID  string
    IP        string
    UserAgent string
    Reason    string
    Timestamp time.Time
}

func (e WebSocketClientDisconnected) GetTopic() string { return "webserver.websocket.clientDisconnected" }

// WebSocketSendMessageToClient is a command to send a message to a specific client.
type WebSocketSendMessageToClient struct {
    ClientID    string
    MessageType string
    Payload     interface{}
}

func (e WebSocketSendMessageToClient) GetTopic() string { return "webserver.websocket.sendToClient" }

// WebSocketBroadcastMessage is a command to send a message to ALL connected clients.
type WebSocketBroadcastMessage struct {
    MessageType string
    Payload     json.RawMessage
}

func (e WebSocketBroadcastMessage) GetTopic() string { return "webserver.websocket.broadcast" }

// =============================================================================
// Statistics Events (emitted by WebServer)
// =============================================================================

// WebSocketClientsChanged is published by the WebSocketHandler whenever the number
// of connected clients changes.
type WebSocketClientsChanged struct {
    Timestamp time.Time
    Count     int
}

func (e WebSocketClientsChanged) GetTopic() string { return "webserver.stats.websocketClientsChanged" }

// WebRTCConnectionsChanged is published by the WebRTCHandler whenever the number
// of active peer connections changes.
type WebRTCConnectionsChanged struct {
    Timestamp time.Time
    Count     int
}

func (e WebRTCConnectionsChanged) GetTopic() string { return "webserver.stats.webrtcConnectionsChanged" }

// =============================================================================
// Application Command Events (emitted by WebServer)
// =============================================================================

// GetScheduleRequested is a generic command to request the schedule.
// The payload can be enriched with the source if needed.
type GetScheduleRequested struct {
    ClientID string    
}

func (e GetScheduleRequested) GetTopic() string { return "webserver.command.getSchedule" }

// CommitScheduleRequested is a generic command to save the schedule.
type CommitScheduleRequested struct {
    ClientID string
    Payload  json.RawMessage
}

func (e CommitScheduleRequested) GetTopic() string { return "webserver.command.commitSchedule" }

// GetStatusRequested is a command to request the current status of OBS and VirtualCam.
type GetStatusRequested struct {
    ClientID string
}

func (e GetStatusRequested) GetTopic() string { return "webserver.command.getStatus" }

// StatusResponse is sent in response to GetStatusRequested with the current system status.
type StatusResponse struct {
    ClientID         string
    OBSConnected     bool
    OBSVersion       string
    VirtualCamActive bool
}

func (e StatusResponse) GetTopic() string { return "webserver.response.status" }

