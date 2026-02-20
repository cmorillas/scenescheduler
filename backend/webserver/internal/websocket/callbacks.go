// backend/webserver/internal/websocket/callbacks.go
//
// Callback definitions for parent (WebServer) to handle internal events.
// This follows the Ambassador pattern - the internal component notifies
// its parent via callbacks, and the parent translates to EventBus events.

package websocket

import (
	"encoding/json"
)

// Callbacks defines all the ways the Handler notifies its parent (WebServer).
// The parent implements these callbacks to translate internal events into
// EventBus publications or other actions.
type Callbacks struct {
	// Connection lifecycle callbacks
	OnClientConnected    func(clientID, ip, userAgent string)
	OnClientDisconnected func(clientID, ip, userAgent, reason string)
	OnClientsChanged     func(count int)

	// Message routing callbacks (client requests)
	OnGetSchedule    func(clientID string)
	OnCommitSchedule func(clientID string, payload json.RawMessage)
	OnGetStatus      func(clientID string)

	// Source preview callbacks
	OnStartPreview func(clientID, remoteAddr string, payload json.RawMessage)
	OnStopPreview  func(clientID, remoteAddr string)
}
