// File: shared/app-state.mjs
// Global Application State
// Specs: Section 5.1 - Simple reactive state management

/**
 * Global application state
 * All state is stored here and mutations trigger UI updates
 */
const AppState = {
    // WebSocket connection state (Frontend ↔ Backend)
    websocket: {
        connected: false,
        status: 'disconnected', // 'disconnected' | 'connecting' | 'connected' | 'error'
        statusText: 'Disconnected'
    },

    // OBS connection state (Backend ↔ OBS)
    obs: {
        connected: false,
        version: null,
        status: 'disconnected', // 'disconnected' | 'connected' | 'unknown'
        statusText: 'Disconnected',
        unknown: false // true when WebSocket is disconnected
    },

    // Live Preview state (video stream availability)
    preview: {
        available: false,
        status: 'unavailable', // 'unavailable' | 'available' | 'connecting' | 'connected' | 'unknown'
        statusText: 'No Stream'
    },

    // Editor state
    editor: {
        isDirty: false,
        changeCount: 0,
        isSyncing: false,
        lastSyncTime: null,
        status: 'clean', // 'clean' | 'dirty' | 'syncing' | 'error'
        statusText: 'Synced'
    },

    // Schedule data (Schedule 1.0 format)
    schedule: null,

    // Editor working copy (modified schedule)
    workingSchedule: null,

    // Current view
    currentView: 'monitor', // 'monitor' | 'editor'

    // Current program (for Monitor view highlighting)
    currentProgram: null
};

/**
 * State change listeners
 * Functions that should be called when state changes
 */
const listeners = [];

/**
 * Register a listener function to be called on state changes
 * @param {Function} callback - Function to call when state changes
 * @returns {Function} Unsubscribe function
 */
export function subscribe(callback) {
    listeners.push(callback);

    // Return unsubscribe function
    return () => {
        const index = listeners.indexOf(callback);
        if (index > -1) {
            listeners.splice(index, 1);
        }
    };
}

/**
 * Notify all listeners that state has changed
 * @param {string} path - Dot-notation path of what changed (e.g., 'editor.isDirty')
 */
function notifyListeners(path) {
    listeners.forEach(listener => {
        try {
            listener(path, AppState);
        } catch (error) {
            console.error('Error in state listener:', error);
        }
    });
}

/**
 * Update WebSocket connection state
 * @param {boolean} connected - Connection status
 * @param {string} status - Status identifier
 * @param {string} statusText - Display text
 */
export function setWebSocketStatus(connected, status, statusText) {
    AppState.websocket.connected = connected;
    AppState.websocket.status = status;
    AppState.websocket.statusText = statusText;
    notifyListeners('websocket');
}

/**
 * Update editor dirty state
 * @param {boolean} isDirty - Whether there are unsaved changes
 * @param {number} changeCount - Number of changes
 */
export function setEditorDirty(isDirty, changeCount = 0) {
    AppState.editor.isDirty = isDirty;
    AppState.editor.changeCount = changeCount;

    // Update status based on dirty state
    if (isDirty) {
        AppState.editor.status = 'dirty';
        AppState.editor.statusText = changeCount > 0
            ? `${changeCount} unsaved change${changeCount !== 1 ? 's' : ''}`
            : 'Unsaved changes';
    } else {
        AppState.editor.status = 'clean';
        AppState.editor.statusText = 'Synced';
    }

    notifyListeners('editor');
}

/**
 * Update editor syncing state
 * @param {boolean} isSyncing - Whether currently syncing
 */
export function setEditorSyncing(isSyncing) {
    AppState.editor.isSyncing = isSyncing;

    if (isSyncing) {
        AppState.editor.status = 'syncing';
        AppState.editor.statusText = 'Syncing...';
    } else {
        // Restore previous status
        if (AppState.editor.isDirty) {
            AppState.editor.status = 'dirty';
            AppState.editor.statusText = AppState.editor.changeCount > 0
                ? `${AppState.editor.changeCount} unsaved change${AppState.editor.changeCount !== 1 ? 's' : ''}`
                : 'Unsaved changes';
        } else {
            AppState.editor.status = 'clean';
            AppState.editor.statusText = 'Synced';
            AppState.editor.lastSyncTime = new Date();
        }
    }

    notifyListeners('editor');
}

/**
 * Update editor error state
 * @param {string} errorMessage - Error message to display
 */
export function setEditorError(errorMessage) {
    AppState.editor.status = 'error';
    AppState.editor.statusText = errorMessage || 'Error syncing';
    notifyListeners('editor');
}

/**
 * Set the main schedule (from server)
 * @param {Object} schedule - Schedule 1.0 object
 * @param {Object} options - Options for loading
 * @param {boolean} options.force - Force update even if dirty
 * @param {boolean} options.fromUser - Manual user action (should prompt if dirty)
 */
export function setSchedule(schedule, options = {}) {
    AppState.schedule = schedule;

    const shouldUpdate = options.force || !AppState.workingSchedule || !AppState.editor.isDirty;

    if (shouldUpdate) {
        AppState.workingSchedule = JSON.parse(JSON.stringify(schedule));
        notifyListeners('workingSchedule');

        // Reset dirty state when loading from server
        if (AppState.editor.isDirty) {
            AppState.editor.isDirty = false;
            AppState.editor.changeCount = 0;
            AppState.editor.status = 'clean';
            AppState.editor.statusText = 'Synced';
            notifyListeners('editor');
        }
    } else if (options.fromUser && AppState.editor.isDirty) {
        // Manual "Get from Server" with unsaved changes - dispatch event to ask user
        document.dispatchEvent(new CustomEvent('schedule:confirmLoad', {
            detail: { schedule }
        }));
    }

    notifyListeners('schedule');
}

/**
 * Update the working schedule (editor modifications)
 * @param {Object} schedule - Modified schedule
 */
export function setWorkingSchedule(schedule) {
    AppState.workingSchedule = schedule;

    // Check if different from main schedule
    const isDifferent = JSON.stringify(AppState.schedule) !== JSON.stringify(schedule);

    if (isDifferent && !AppState.editor.isDirty) {
        setEditorDirty(true, 1);
    } else if (!isDifferent && AppState.editor.isDirty) {
        setEditorDirty(false, 0);
    }

    notifyListeners('workingSchedule');
}

/**
 * Set current view
 * @param {string} view - View name ('monitor' | 'editor')
 */
export function setCurrentView(view) {
    AppState.currentView = view;
    notifyListeners('currentView');
}

/**
 * Set current program (for Monitor highlighting)
 * @param {Object} program - Current program object
 */
export function setCurrentProgram(program) {
    AppState.currentProgram = program;
    notifyListeners('currentProgram');
}

/**
 * Update OBS connection status
 * @param {boolean} connected - Connection status
 * @param {Object} details - Additional details (version, etc.)
 */
export function setOBSStatus(connected, details = {}) {
    AppState.obs.connected = connected;
    AppState.obs.version = details.version || AppState.obs.version;
    AppState.obs.status = connected ? 'connected' : 'disconnected';
    AppState.obs.statusText = connected
        ? (details.version ? `Connected (${details.version})` : 'Connected')
        : 'Disconnected';
    notifyListeners('obs');
}

/**
 * Update Live Preview stream status
 * @param {string} status - Status ('unavailable' | 'connecting' | 'connected')
 * @param {Object} details - Additional details
 */
export function setPreviewStatus(status, details = {}) {
    AppState.preview.status = status;
    AppState.preview.available = (status === 'available' || status === 'connected');

    switch (status) {
        case 'available':
            AppState.preview.statusText = 'Stream Ready';
            break;
        case 'connecting':
            AppState.preview.statusText = 'Connecting...';
            break;
        case 'connected':
            AppState.preview.statusText = 'Stream Active';
            break;
        case 'unavailable':
        default:
            AppState.preview.statusText = 'No Stream';
            break;
    }

    notifyListeners('preview');
}

/**
 * Get current state (read-only)
 * DO NOT mutate the returned object directly!
 * Use the setter functions instead
 */
export function getState() {
    return AppState;
}

// State logging disabled for production

export default {
    subscribe,
    setWebSocketStatus,
    setOBSStatus,
    setPreviewStatus,
    setEditorDirty,
    setEditorSyncing,
    setEditorError,
    setSchedule,
    setWorkingSchedule,
    setCurrentView,
    setCurrentProgram,
    getState
};
