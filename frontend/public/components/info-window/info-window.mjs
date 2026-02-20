// File: components/info-window/info-window.mjs
// This module controls the Info Window panel, which displays WebSocket status and server messages.
// It listens for custom DOM events dispatched by the WebSocket service.

/**
 * Initializes the info window component by setting up event listeners.
 * @param {HTMLElement} container - The container element for the info window.
 */
export function initInfoWindow(container) {
    const logEl = container.querySelector('#log-output');
    const statusLightEl = container.querySelector('#info-status-light');
    const statusLabelEl = container.querySelector('#info-status-text');

    if (!logEl || !statusLightEl || !statusLabelEl) {
        // Missing required elements - silently fail
        return;
    }

    // Listen for WebSocket status changes and update the UI accordingly.
    document.addEventListener('ws:statusChange', (e) => {
        const { text, color } = e.detail;
        statusLightEl.className = `status-light ${color}`;
        statusLabelEl.textContent = `WebSocket ${text}`;
    });

    // Listen for incoming WebSocket messages to log them (excluding currentSchedule)
    document.addEventListener('ws:message', (e) => {
        const msg = e.detail;
        const action = msg.action || 'N/A';

        // Skip logging currentSchedule messages (schedule updates are frequent and noisy)
        if (action === 'currentSchedule') {
            return;
        }

        // Format the message for display in the log
        const payloadStr = JSON.stringify(msg.payload || {});
        const line = `[IN] Action: ${action} | Payload: ${payloadStr}\n`;

        logEl.textContent += line;
        logEl.scrollTop = logEl.scrollHeight; // Auto-scroll to the bottom
    });
}

