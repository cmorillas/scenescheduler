// File: shared/ui-updater.mjs
// UI Update Functions
// Specs: Section 5.2 - Manual UI updates based on state changes

import { subscribe, getState } from './app-state.mjs';

/**
 * Initialize UI updater
 * Subscribes to state changes and updates DOM accordingly
 */
export function initUIUpdater() {
    // Subscribe to all state changes
    subscribe((path, state) => {
        if (path === 'websocket') {
            updateServerStatus(state.websocket);
        } else if (path === 'obs') {
            updateOBSStatus(state.obs);
        } else if (path === 'preview') {
            updatePreviewStatus(state.preview);
        } else if (path === 'editor') {
            updateEditorStatus(state.editor);
        } else if (path === 'currentView') {
            // View switching is handled by view-switcher.mjs
        }
    });

    // Initial update
    const state = getState();
    updateServerStatus(state.websocket);
    updateOBSStatus(state.obs);
    updatePreviewStatus(state.preview);
    updateEditorStatus(state.editor);
}

/**
 * Update Server status indicator (Frontend ↔ Backend WebSocket)
 * @param {Object} websocketState - WebSocket state object
 */
function updateServerStatus(websocketState) {
    const statusElement = document.getElementById('server-status');
    const statusLight = statusElement?.querySelector('.status-light');

    if (!statusElement || !statusLight) {
        return;
    }

    // Update status light color
    statusLight.classList.remove('green', 'red', 'orange');

    switch (websocketState.status) {
        case 'connected':
            statusLight.classList.add('green');
            statusElement.title = 'Server: Connected';
            break;
        case 'connecting':
            statusLight.classList.add('orange');
            statusElement.title = 'Server: Connecting...';
            break;
        case 'disconnected':
        case 'error':
            statusLight.classList.add('red');
            statusElement.title = 'Server: Disconnected';
            break;
    }
}

/**
 * Update OBS status indicator (Backend ↔ OBS)
 * @param {Object} obsState - OBS state object
 */
function updateOBSStatus(obsState) {
    const statusElement = document.getElementById('obs-status');
    const statusLight = statusElement?.querySelector('.status-light');

    if (!statusElement || !statusLight) {
        return;
    }

    // Update status light color
    statusLight.classList.remove('green', 'red');

    if (obsState.connected) {
        statusLight.classList.add('green');
        const tooltip = obsState.version
            ? `OBS: Connected (${obsState.version})`
            : 'OBS: Connected';
        statusElement.title = tooltip;
    } else {
        statusLight.classList.add('red');
        statusElement.title = 'OBS: Disconnected';
    }
}

/**
 * Update Preview status indicator and control video play button
 * @param {Object} previewState - Preview state object
 */
function updatePreviewStatus(previewState) {
    const statusElement = document.getElementById('preview-status');
    const statusLight = statusElement?.querySelector('.status-light');
    const videoElement = document.getElementById('videoElement');

    if (!statusElement || !statusLight) {
        return;
    }

    // Update status light color
    statusLight.classList.remove('green', 'red', 'orange');

    switch (previewState.status) {
        case 'available':
            // Stream available from VirtualCam, ready to connect
            statusLight.classList.add('green');
            statusElement.title = 'Preview: Stream Ready (click play)';
            if (videoElement) {
                videoElement.removeAttribute('disabled');
                videoElement.style.pointerEvents = 'auto';
                videoElement.style.opacity = '1';
            }
            break;
        case 'connected':
            // Stream connected and playing
            statusLight.classList.add('green');
            statusElement.title = 'Preview: Stream Active';
            if (videoElement) {
                videoElement.removeAttribute('disabled');
                videoElement.style.pointerEvents = 'auto';
                videoElement.style.opacity = '1';
            }
            break;
        case 'connecting':
            statusLight.classList.add('orange');
            statusElement.title = 'Preview: Connecting...';
            if (videoElement) {
                videoElement.setAttribute('disabled', 'true');
                videoElement.style.pointerEvents = 'none';
                videoElement.style.opacity = '0.5';
            }
            break;
        case 'unavailable':
        default:
            statusLight.classList.add('red');
            statusElement.title = 'Preview: No Stream';
            if (videoElement) {
                videoElement.setAttribute('disabled', 'true');
                videoElement.style.pointerEvents = 'none';
                videoElement.style.opacity = '0.5';
            }
            break;
    }
}

/**
 * Update Editor status bar (sync state)
 * @param {Object} editorState - Editor state object
 */
function updateEditorStatus(editorState) {
    const statusElement = document.getElementById('editor-status');
    const statusCircle = statusElement?.querySelector('.status-circle');
    const statusText = document.getElementById('editor-status-text');

    if (!statusElement || !statusText) {
        return;
    }

    // Remove all status classes
    statusElement.classList.remove('clean', 'dirty', 'syncing', 'error');

    // Add current status class
    statusElement.classList.add(editorState.status);

    // Update status text
    statusText.textContent = editorState.statusText;
}

/**
 * Update Info Window (Activity Log)
 * @param {string} message - Log message
 * @param {string} type - Message type ('info', 'warning', 'error')
 */
export function addLogMessage(message, type = 'info') {
    const logOutput = document.getElementById('log-output');

    if (!logOutput) {
        return;
    }

    const timestamp = new Date().toLocaleTimeString();
    const logEntry = document.createElement('div');
    logEntry.className = `log-entry log-${type}`;
    logEntry.textContent = `[${timestamp}] ${message}`;

    logOutput.appendChild(logEntry);

    // Auto-scroll to bottom
    logOutput.scrollTop = logOutput.scrollHeight;

    // Limit log entries (keep last 100)
    while (logOutput.children.length > 100) {
        logOutput.removeChild(logOutput.firstChild);
    }
}

/**
 * Clear activity log
 */
export function clearLog() {
    const logOutput = document.getElementById('log-output');

    if (logOutput) {
        logOutput.innerHTML = '';
    }
}

export default {
    initUIUpdater,
    addLogMessage,
    clearLog
};
