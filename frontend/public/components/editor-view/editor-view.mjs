// File: components/editor-view/editor-view.mjs
// Responsibility: Coordinate all Editor tab components

import { initEditorCalendar } from '../calendar/calendar-editor.mjs';

/**
 * Initialize Editor View
 * Specs: Section 4.5 (modified - actions moved to calendar menu)
 */
export async function initEditorView() {

    // Get containers
    const editorCalendarContainer = document.getElementById('editor-calendar');

    if (editorCalendarContainer) {
        // Initialize editor calendar (editable)
        // The calendar will have a menu button (•••) that includes:
        // - New Schedule
        // - Load from File
        // - Save to File
        // - Get from Server
        // - Commit to Server
        // - Revert Changes
        initEditorCalendar(editorCalendarContainer);
    }

}
