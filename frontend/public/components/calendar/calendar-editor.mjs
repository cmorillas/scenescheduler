// File: components/calendar/calendar-editor.mjs
// Editable calendar for Editor view
// Specs: Section 4.5

// FullCalendar is loaded globally from vendor/fullcalendar-6.1.19/dist/index.global.min.js
// Available as: FullCalendar (global variable)

import { openTaskModal } from './modal.mjs';
import { updateZoom } from './zoom.mjs';
import { handleGridAction } from './grid-actions.mjs';
import { createMenu, toggleMenu } from './menu.mjs';
import { importSchedule, exportSchedule } from './schedule-adapter.mjs';
import { subscribe, getState, setWorkingSchedule, setEditorDirty, setSchedule } from '../../shared/app-state.mjs';
import { applyRecurringEventStyles, durationStringToMs } from './helpers.mjs';

// ================================
// INITIALIZATION
// ================================

/**
 * Initialize Editor Calendar (editable)
 * @param {HTMLElement} container - DOM element to render calendar into
 * @returns {Calendar} FullCalendar instance
 */
export function initEditorCalendar(container) {

    // Check FullCalendar is available
    if (typeof FullCalendar === 'undefined') {
        console.error('FullCalendar is not loaded! Make sure index.html includes the FullCalendar script.');
        return null;
    }

    // Clear placeholder content and create calendar instance element
    container.innerHTML = '';

    // Editor-specific configuration
    const calendar = new FullCalendar.Calendar(container, {
        initialView: 'timeGridWeek',
        headerToolbar: {
            left: 'prev,next today zoomIn,zoomOut',
            center: 'title',
            right: 'timeGridWeek,timeGridDay menu'
        },
        customButtons: {
            zoomIn: { text: '+', click: () => updateZoom(calendar, 'in') },
            zoomOut: { text: '-', click: () => updateZoom(calendar, 'out') },
            menu: { text: '...', click: toggleMenu }
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
        editable: true,         // EDITABLE
        selectable: true,       // EDITABLE
        eventOverlap: false,
        eventTimeFormat: {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        },
        scrollTime: '14:00:00',
        events: [],

        // ================================
        // EVENT HANDLERS
        // ================================
        eventDidMount: function(info) {
            // Double-click to edit
            info.el.addEventListener('dblclick', () => {
                openTaskModal(calendar, { event: info.event });
            });

            // Apply recurring event styles
            applyRecurringEventStyles(info.event, info.el);
        },

        eventDrop: (info) => {
            handleGridAction(calendar, info);
            markDirty(calendar);
        },

        eventResize: (info) => {
            handleGridAction(calendar, info);
            markDirty(calendar);
        },

        eventAdd: (info) => {
            // Mark as dirty when event is added (from modal or other sources)
            markDirty(calendar);
        },

        eventChange: (info) => {
            // Mark as dirty when event is modified
            markDirty(calendar);
        },

        eventRemove: (info) => {
            // Mark as dirty when event is deleted
            markDirty(calendar);
        },

        dateClick: (info) => {
            // Get the current slot duration (e.g., "01:00:00" or "00:30:00")
            const slotDurationStr = calendar.getOption('slotDuration');

            // Convert the duration string to milliseconds
            const durationMs = durationStringToMs(slotDurationStr);

            // Create a start date object from the click info
            const startDate = new Date(info.dateStr);

            // Calculate the end time by adding the dynamic slot duration
            const endDate = new Date(startDate.getTime() + durationMs);

            // Subtract one second as requested
            endDate.setSeconds(endDate.getSeconds() - 1);

            // Open the modal with the calculated start and end times
            openTaskModal(calendar, { start: info.dateStr, end: endDate.toISOString() });
        },

        select: (info) => {
            // Create a Date object from the end string
            const endDate = new Date(info.endStr);

            // Subtract one second
            endDate.setSeconds(endDate.getSeconds() - 1);

            // Open the modal with the modified end time converted back to an ISO string
            openTaskModal(calendar, { start: info.startStr, end: endDate.toISOString() });
        },
    });


    try {
        calendar.render();
    } catch (error) {
        console.error('Editor: Error rendering calendar:', error);
    }

    createMenu(calendar);

    // Subscribe to working schedule updates
    subscribe((path, state) => {
        if (path === 'workingSchedule' && state.workingSchedule) {
            importSchedule(calendar, state.workingSchedule);
        }
    });

    // Load initial working schedule or main schedule if available
    const currentState = getState();
    const scheduleToLoad = currentState.workingSchedule || currentState.schedule;
    if (scheduleToLoad) {
        importSchedule(calendar, scheduleToLoad);

        // Set working schedule if not already set
        if (!currentState.workingSchedule) {
            setWorkingSchedule(scheduleToLoad);
        }
    }

    // Note: Auto-load from server is now handled by app-state.mjs setSchedule()
    // Manual "Get from Server" action still prompts if isDirty via menu-actions.mjs

    // Listen for confirmation requests when loading schedule with unsaved changes
    document.addEventListener('schedule:confirmLoad', (e) => {
        const { schedule } = e.detail;
        if (confirm('Load schedule from server? This will replace all current unsaved events.')) {
            // User confirmed - force load the schedule
            setSchedule(schedule, { force: true });
        }
    });

    // Re-render calendar when view becomes visible
    document.addEventListener('view:changed', (e) => {
        if (e.detail.view === 'editor') {
            setTimeout(() => {
                calendar.updateSize();
            }, 100);
        }
    });

    return calendar;
}

/**
 * Mark the editor as dirty (has unsaved changes)
 * @param {Calendar} calendar - FullCalendar instance
 */
function markDirty(calendar) {
    const eventCount = calendar.getEvents().length;

    // Export current state to working schedule
    const workingSchedule = exportSchedule(calendar);
    setWorkingSchedule(workingSchedule);

    // Mark editor as dirty
    setEditorDirty(true, eventCount);
}
