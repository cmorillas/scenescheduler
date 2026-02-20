// Scene Scheduler - OBS Automation Software
// Copyright (c) 2025 Scene Scheduler, S.L. - All Rights Reserved
// This is proprietary software. Unauthorized copying or distribution is prohibited.

// File: main.mjs
// Application Entry Point

import { connect as connectWebSocket } from './services/websocket.mjs';
import { initViewSwitcher } from './components/view-switcher/view-switcher.mjs';
import { initMonitorView } from './components/monitor-view/monitor-view.mjs';
import { initEditorView } from './components/editor-view/editor-view.mjs';
import { initUIUpdater } from './shared/ui-updater.mjs';

/**
 * Main application initialization
 * Follows the startup sequence from specs section 6.1
 */
async function main() {
    // 1. Verify all required containers exist
    const requiredElements = [
        '#monitor-view',
        '#editor-view',
        '#view-dropdown'
    ];

    for (const selector of requiredElements) {
        if (!document.querySelector(selector)) {
            console.error(`Required element not found: ${selector}`);
            return;
        }
    }

    // 2. Initialize UI Updater (state → UI)
    initUIUpdater();

    // 3. Initialize View Switcher (tab navigation)
    initViewSwitcher();

    // 4. Initialize Monitor View (read-only observation)
    await initMonitorView();

    // 5. Initialize Editor View (editable workspace)
    await initEditorView();

    // 6. Connect WebSocket (will auto-fetch schedule)
    connectWebSocket();
}

// Start the application when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', main);
} else {
    main();
}
