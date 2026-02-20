// File: services/websocket.mjs
// This module handles WebSocket communication with the server.
// It uses a simple protocol: { action: "string", payload: {} }
//
// --- Outgoing Actions (Client -> Server) ---
// - getSchedule: Requests the current schedule.
//   => { action: "getSchedule", payload: {} }
// - commitSchedule: Sends the current schedule to be saved.
//   => { action: "commitSchedule", payload: { Schedule 1.0 JSON object } }
//
// --- Incoming Actions (Server -> Client) ---
// - currentSchedule: Carries the full schedule payload from the server.
//   => { action: "currentSchedule", payload: { Schedule 1.0 JSON object } }
// - log: Carries a generic message for logging.
//   => { action: "log", payload: "Server message here..." }

import { setWebSocketStatus, setSchedule, setOBSStatus, setPreviewStatus } from '../shared/app-state.mjs';
import { addLogMessage } from '../shared/ui-updater.mjs';

// ================================
// MODULE STATE (Private)
// ================================
let ws = null;
const reconnectTimeout = 5000; // Reconnect delay: 5 seconds
const url = '/ws'; // This will be dynamically resolved
let pendingScheduleRequest = null; // Track if getSchedule was from user action
let isReconnecting = false; // Track if we're in reconnection mode
let reconnectAttempts = 0; // Count reconnection attempts

// ================================
// PUBLIC API
// ================================

/**
 * Establishes and manages the WebSocket connection.
 */
function connect() {
    // Update state: connecting
    setWebSocketStatus(false, 'connecting', 'Connecting...');

    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const fullUrl = `${proto}//${window.location.host}${url}`;
    ws = new WebSocket(fullUrl);

    ws.onopen = () => {
        // Reset reconnection state
        const wasReconnecting = isReconnecting;
        isReconnecting = false;
        reconnectAttempts = 0;

        // Update state: connected
        setWebSocketStatus(true, 'connected', 'Connected');

        // Log only initial connection or successful reconnection
        if (wasReconnecting) {
            addLogMessage('Reconnected to server', 'info');
        } else {
            addLogMessage('Connected to server', 'info');
        }

        // Dispatch custom event (for backward compatibility)
        document.dispatchEvent(new CustomEvent('ws:statusChange', {
            detail: { text: 'Connected', color: 'green' }
        }));

        // Auto-fetch schedule and status from server on connect
        pendingScheduleRequest = { fromUser: false };
        sendMessage('getSchedule', {});
        sendMessage('getStatus', {});
    };

    ws.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);

            // Handle different message types
            handleMessage(message);

            // Dispatch custom event (for backward compatibility)
            document.dispatchEvent(new CustomEvent('ws:message', { detail: message }));

        } catch (e) {
            console.error('Failed to parse message:', e);
            addLogMessage(`Error parsing message: ${e.message}`, 'error');
        }
    };

    ws.onclose = () => {
        // Update state: disconnected
        setWebSocketStatus(false, 'disconnected', 'Disconnected');

        // When WebSocket disconnects, we don't know the real state of OBS and preview
        // So we should set them to unknown/unavailable state
        setOBSStatus(false, { unknown: true });
        setPreviewStatus('unknown');

        // Mark as reconnecting and increment attempts
        isReconnecting = true;
        reconnectAttempts++;

        // Log only the first disconnection, not every retry
        if (reconnectAttempts === 1) {
            addLogMessage('Disconnected from server, reconnecting...', 'warning');
        }

        // Dispatch custom event (for backward compatibility)
        document.dispatchEvent(new CustomEvent('ws:statusChange', {
            detail: { text: 'Disconnected', color: 'red' }
        }));

        // Attempt reconnection
        setTimeout(connect, reconnectTimeout);
    };

    ws.onerror = (error) => {
        // Only log errors if not in reconnection mode
        // During reconnection, errors are expected and don't need console spam
        if (!isReconnecting) {
            console.error('WebSocket error:', error);
            setWebSocketStatus(false, 'error', 'Connection error');
            addLogMessage('Connection error', 'error');
        }
        // If reconnecting, errors are silently ignored (onclose will handle reconnection)
    };
}

/**
 * Handle incoming WebSocket messages
 * @param {Object} message - Parsed message object
 */
function handleMessage(message) {
    const { action, payload } = message;

    switch (action) {
        case 'currentSchedule':
            // Update schedule in state with context about the request
            const options = pendingScheduleRequest || {};
            pendingScheduleRequest = null; // Clear pending request
            setSchedule(payload, options);

            // Log schedule update with event count
            const eventCount = payload?.schedule?.length || 0;
            addLogMessage(`Schedule loaded (${eventCount} events)`, 'info');
            break;

        case 'log':
            // Add log message to activity log
            addLogMessage(payload, 'info');
            break;

        case 'obsConnected':
            // OBS connection established
            setOBSStatus(true, {
                version: payload.obsVersion,
                timestamp: payload.timestamp
            });
            addLogMessage(`OBS connected: ${payload.obsVersion}`, 'info');
            break;

        case 'obsDisconnected':
            // OBS connection lost
            setOBSStatus(false, {
                timestamp: payload.timestamp
            });
            addLogMessage('OBS disconnected', 'warning');
            break;

        case 'virtualCamStarted':
            // VirtualCam started - stream is now available
            setPreviewStatus('available');
            break;

        case 'virtualCamStopped':
            // VirtualCam stopped - stream no longer available
            setPreviewStatus('unavailable');
            break;

        case 'currentStatus':
            // Initial status received from server
            if (payload.obsConnected) {
                setOBSStatus(true, {
                    version: payload.obsVersion,
                    timestamp: new Date()
                });
            } else {
                setOBSStatus(false, {
                    timestamp: new Date()
                });
            }

            // Set preview status based on VirtualCam state
            if (payload.virtualCamActive) {
                setPreviewStatus('available');
            } else {
                setPreviewStatus('unavailable');
            }
            break;

        case 'previewReady':
            // Source preview HLS stream is ready
            document.dispatchEvent(new CustomEvent('preview:ready', {
                detail: { hlsUrl: payload.hlsUrl }
            }));
            break;

        case 'previewError':
            // Source preview generation failed
            document.dispatchEvent(new CustomEvent('preview:error', {
                detail: { error: payload.error }
            }));
            break;

        case 'previewStopped':
            // Source preview was automatically stopped (timeout, etc)
            document.dispatchEvent(new CustomEvent('preview:stopped', {
                detail: { reason: payload.reason }
            }));
            break;

        default:
            // Unknown action - silently ignore in production
            break;
    }
}

/**
 * Sends a structured message to the server.
 * @param {string} action - The action identifier (e.g., 'getSchedule').
 * @param {object} [payload] - The data payload to send.
 */
function sendMessage(action, payload = {}) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ action, payload }));
    }
    // Silently fail if WebSocket is not open
}

/**
 * Request schedule from server (for manual user action)
 * This will prompt the user if there are unsaved changes
 */
function getScheduleFromUser() {
    pendingScheduleRequest = { fromUser: true };
    sendMessage('getSchedule', {});
}

// Export public functions using named exports (per spec section 13.2)
export { connect, sendMessage, getScheduleFromUser };

