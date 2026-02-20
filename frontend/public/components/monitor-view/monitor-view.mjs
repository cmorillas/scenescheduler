// File: components/monitor-view/monitor-view.mjs
// Responsibility: Coordinate all Monitor tab components

import { initInfoWindow } from '../info-window/info-window.mjs';
import { initLivePreview, cleanupLivePreview } from '../live-preview/live-preview.mjs';
import { initMonitorCalendar } from '../calendar/calendar-monitor.mjs';

/**
 * Initialize Monitor View
 * Specs: Section 4.4
 */
export async function initMonitorView() {

    // Get containers
    const livePreviewContainer = document.getElementById('live-preview-container');
    const infoWindowContainer = document.getElementById('info-window-container');
    const monitorCalendarContainer = document.getElementById('monitor-calendar');

    // Initialize sub-components
    if (livePreviewContainer) {
        await initLivePreview(livePreviewContainer);
    }

    if (infoWindowContainer) {
        await initInfoWindow(infoWindowContainer);
    }

    if (monitorCalendarContainer) {
        initMonitorCalendar(monitorCalendarContainer);
    }

    // Setup cleanup when leaving Monitor view
    const handleViewChange = (e) => {
        if (e.detail.view !== 'monitor') {
            // User is leaving Monitor view - cleanup resources
            cleanupLivePreview();
            document.removeEventListener('view:changed', handleViewChange);
        }
    };

    document.addEventListener('view:changed', handleViewChange);
}
