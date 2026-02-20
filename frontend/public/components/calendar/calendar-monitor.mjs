// File: components/calendar/calendar-monitor.mjs
// Read-only calendar for Monitor view
// Specs: Section 4.4

// FullCalendar is loaded globally from vendor/fullcalendar-6.1.19/dist/index.global.min.js
// Available as: FullCalendar (global variable)

import { updateZoom } from './zoom.mjs';
import { importSchedule } from './schedule-adapter.mjs';
import { subscribe, getState } from '../../shared/app-state.mjs';
import { openTaskModal } from './modal.mjs';
import { applyRecurringEventStyles } from './helpers.mjs';

// ================================
// INITIALIZATION
// ================================

/**
 * Initialize Monitor Calendar (read-only)
 * @param {HTMLElement} container - DOM element to render calendar into
 * @returns {Calendar} FullCalendar instance
 */
export function initMonitorCalendar(container) {

    // Check FullCalendar is available
    if (typeof FullCalendar === 'undefined') {
        console.error('FullCalendar is not loaded! Make sure index.html includes the FullCalendar script.');
        return null;
    }

    // Clear placeholder content and create calendar instance element
    container.innerHTML = '';

    // Monitor-specific configuration
    const calendar = new FullCalendar.Calendar(container, {
        initialView: 'timeGridWeek',
        headerToolbar: {
            left: 'prev,next today zoomIn,zoomOut',
            center: 'title',
            right: 'timeGridWeek,timeGridDay'
        },
        customButtons: {
            zoomIn: { text: '+', click: () => updateZoom(calendar, 'in') },
            zoomOut: { text: '-', click: () => updateZoom(calendar, 'out') }
        },
        slotDuration: '01:00:00',
        slotLabelInterval: '01:00:00',
        slotLabelFormat: {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        },
        slotMinTime: '00:00:00',
        slotMaxTime: '24:00:00',
        height: '100%',
        expandRows: true,
        allDaySlot: false,
        nowIndicator: true,
        editable: false,        // READ-ONLY
        selectable: false,      // READ-ONLY
        eventOverlap: false,
        eventTimeFormat: {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        },
        scrollTime: '14:00:00',
        events: [],

        // Event rendering
        eventDidMount: function(info) {
            const now = new Date();
            const event = info.event;

            // Apply recurring event styles
            applyRecurringEventStyles(event, info.el);

            // Highlight current program
            if (event.start && event.end) {
                if (now >= event.start && now <= event.end) {
                    info.el.classList.add('current-program');
                }
            }
        },

        // Event click - open read-only modal
        eventClick: function(info) {
            openTaskModal(calendar, { event: info.event, readOnly: true });
        }
    });


    try {
        calendar.render();
    } catch (error) {
        console.error('Monitor: Error rendering calendar:', error);
    }

    // Subscribe to schedule updates from WebSocket
    subscribe((path, state) => {
        if (path === 'schedule' && state.schedule) {
            importSchedule(calendar, state.schedule);
        }
    });

    // Load initial schedule if available
    const currentState = getState();
    if (currentState.schedule) {
        importSchedule(calendar, currentState.schedule);
    }

    // Note: Schedule updates are now handled via app-state subscriptions above

    return calendar;
}
