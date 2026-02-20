// backend/obsclient/connection.go
//
// This file contains methods related to managing the connection session with
// the OBS websocket server. It handles connecting, disconnecting, monitoring,
// and ingesting events from OBS.
//
// Contents:
// - Connection Management
// - Session Goroutines
// - Initial State Synchronization
// - Internal Helpers

package obsclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/gorilla/websocket"
	"scenescheduler/backend/eventbus"
)

const (
	// Interval to send a 'ping' to OBS to ensure the connection is alive.
	keepAliveInterval = 10 * time.Second
)

// ============================================================================
// CONNECTION MANAGEMENT
// ============================================================================

// connect establishes a new connection, handling the full handshake.
func (c *OBSClient) connect() error {
	c.logger.Debug("Attempting to connect to OBS...", "host", c.config.Host, "port", c.config.Port)

	dialer := &websocket.Dialer{
		NetDialContext:   (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		HandshakeTimeout: 5 * time.Second,
	}

	client, err := goobs.New(
		fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		goobs.WithPassword(c.config.Password),
		goobs.WithResponseTimeout(5*time.Second),
		goobs.WithDialer(dialer),
	)
	if err != nil {
		return fmt.Errorf("failed to connect and authenticate: %w", err)
	}

	versionResp, err := client.General.GetVersion()
	if err != nil {
		_ = client.Disconnect()
		return fmt.Errorf("failed to get version info: %w", err)
	}

	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	if c.connection != nil {
		_ = client.Disconnect() // A connection was established concurrently, discard this one.
		return nil
	}

	connCtx, connCancel := context.WithCancel(c.ctx)
	c.connection = &connection{
		client:     client,
		ctx:        connCtx,
		cancelCtx:  connCancel,
		obsVersion: versionResp.ObsVersion,
	}

	c.logger.InfoGui("Successfully connected to OBS.", "obsVersion", versionResp.ObsVersion)
	eventbus.Publish(c.bus, eventbus.OBSConnected{
		OBSVersion: versionResp.ObsVersion,
		Timestamp:  time.Now(),
	})

	return nil
}

// disconnect closes the current connection and cleans up its state.
func (c *OBSClient) disconnect(reason string) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	if c.connection == nil {
		return
	}

	c.logger.InfoGui("Disconnecting from OBS", "reason", reason)

	if c.connection.cancelCtx != nil {
		c.connection.cancelCtx()
	}
	if c.connection.client != nil {
		_ = c.connection.client.Disconnect()
	}

	c.connection = nil

	eventbus.Publish(c.bus, eventbus.OBSDisconnected{
		Error:     errors.New(reason),
		Timestamp: time.Now(),
	})
}

// ============================================================================
// SESSION GOROUTINES
// ============================================================================

// monitorConnection watches an active connection by periodically pinging it.
func (c *OBSClient) monitorConnection() {
	c.logger.Debug("Health monitor started.")
	defer c.logger.Debug("Health monitor stopped.")

	client, sessionCtx := c.getActiveClientAndContext()
	if client == nil {
		return
	}

	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := client.General.GetStats(); err != nil {
				c.disconnect(fmt.Sprintf("health check failed: %v", err))
				return
			}
		case <-sessionCtx.Done():
			return
		}
	}
}

// startOBSEventListener starts the loop to process incoming events from OBS.
func (c *OBSClient) startOBSEventListener() {
	client, sessionCtx := c.getActiveClientAndContext()
	if client == nil {
		return
	}

	c.logger.Debug("Starting OBS event listener.")
	defer c.logger.Debug("Stopping OBS event listener.")

	for {
		select {
		case event, ok := <-client.IncomingEvents:
			if !ok {
				c.disconnect("event channel closed")
				return
			}
			c.processOBSEvent(event)
		case <-sessionCtx.Done():
			return
		}
	}
}

// processOBSEvent identifies and translates an OBS event for the event bus.
func (c *OBSClient) processOBSEvent(event any) {
	if event == nil {
		return
	}
	switch e := event.(type) {
	case *events.VirtualcamStateChanged:
		if e.OutputActive {
			c.logger.Debug("Event from OBS: Virtualcam Started")
			eventbus.Publish(c.bus, eventbus.OBSVirtualCamStarted{Timestamp: time.Now()})
		} else {
			c.logger.Debug("Event from OBS: Virtualcam Stopped")
			eventbus.Publish(c.bus, eventbus.OBSVirtualCamStopped{Timestamp: time.Now()})
		}
	default:
		// Other events can be handled here.
	}
}

// ============================================================================
// INITIAL STATE SYNCHRONIZATION
// ============================================================================

// checkInitialState queries OBS for initial statuses and publishes synthetic events.
func (c *OBSClient) checkInitialState() {
	client, _ := c.getActiveClientAndContext()
	if client == nil {
		c.logger.Error("Cannot check initial state: client not available")
		return
	}

	// Check Virtual Camera status.
	resp, err := client.Outputs.GetVirtualCamStatus()
	if err != nil {
		c.logger.Warn("Could not get initial virtual camera status", "error", err)
		return
	}

	if resp.OutputActive {
		c.logger.Debug("Initial state check: Virtual camera is already active. Publishing synthetic start event.")
		eventbus.Publish(c.bus, eventbus.OBSVirtualCamStarted{Timestamp: time.Now()})
	} else {
		c.logger.Debug("Initial state check: Virtual camera is inactive.")
	}
}

// ============================================================================
// PUBLIC API FOR STATUS QUERIES
// ============================================================================

// ConnectionStatus contains the current state of OBS connection and virtual camera.
type ConnectionStatus struct {
	IsConnected      bool
	OBSVersion       string
	VirtualCamActive bool
}

// GetCurrentStatus returns the current connection state and virtual camera status.
// This is thread-safe and can be called from any goroutine.
func (c *OBSClient) GetCurrentStatus() ConnectionStatus {
	client, _ := c.getActiveClientAndContext()

	status := ConnectionStatus{
		IsConnected:      client != nil,
		OBSVersion:       "",
		VirtualCamActive: false,
	}

	if client == nil {
		return status
	}

	// Get OBS version
	c.stateMu.RLock()
	if c.connection != nil {
		status.OBSVersion = c.connection.obsVersion
	}
	c.stateMu.RUnlock()

	// Check virtual camera status
	resp, err := client.Outputs.GetVirtualCamStatus()
	if err != nil {
		c.logger.Warn("Failed to get virtual camera status", "error", err)
		return status
	}

	status.VirtualCamActive = resp.OutputActive
	return status
}

// ============================================================================
// INTERNAL HELPERS
// ============================================================================

// getActiveClientAndContext is a thread-safe helper to get the active client
// and its associated context.
func (c *OBSClient) getActiveClientAndContext() (*goobs.Client, context.Context) {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	if c.connection == nil || c.connection.client == nil {
		return nil, nil
	}
	return c.connection.client, c.connection.ctx
}

