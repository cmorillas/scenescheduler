// File: components/calendar/calendar-events.mjs
// This module centralizes all logic for creating, updating, and deleting calendar events.

import { dateToHHMMSS } from './helpers.mjs';

/**
 * Creates or updates an event on the calendar.
 * This is the single source of truth for all event modifications.
 * @param {Calendar} calendar - The FullCalendar instance.
 * @param {Object} eventData - The new data for the event.
 * @param {EventApi | null} existingEvent - The existing event to be replaced, if any.
 */
function updateEvent(calendar, eventData, existingEvent = null) {
    if (existingEvent) {
        // If we are updating, remove the old event first.
        existingEvent.remove();
    }
    // Add the new or updated event to the calendar.
    calendar.addEvent(eventData);
}

/**
 * Deletes an event from the calendar after user confirmation.
 * @param {EventApi} eventToDelete - The event object to be deleted.
 * @returns {boolean} - True if the event was deleted, false otherwise.
 */
function deleteEvent(eventToDelete) {
    if (!eventToDelete) return false;
    
    if (confirm('Are you sure you want to delete this event?')) {
        eventToDelete.remove();
        return true;
    }
    return false;
}

/**
 * Prepares the data for a recurring event update after a drag or resize action.
 * @param {EventApi} event - The event that was moved.
 * @param {Object} info - The info object from the FullCalendar callback.
 * @returns {Object} A complete event data object for the updated series.
 */
function prepareRecurringUpdate(event, info) {
    const newStartTime = dateToHHMMSS(info.event.start);
    const newEndTime = dateToHHMMSS(info.event.end);

    // Build the new master event object by copying existing properties and overriding times.
    return {
        id: event.id,
        title: event.title,
        backgroundColor: event.backgroundColor,
        borderColor: event.borderColor,
        textColor: event.textColor,
        classNames: event.classNames,
        extendedProps: {
            ...event.extendedProps,
            recurrence: {
                ...event.extendedProps.recurrence,
                startTime: newStartTime,
                endTime: newEndTime
            }
        },
        // Provide FullCalendar with the necessary top-level properties for rendering
        daysOfWeek: event.extendedProps.recurrence.daysOfWeek,
        startTime: newStartTime,
        endTime: newEndTime,
        startRecur: event.extendedProps.recurrence.startRecur,
        endRecur: event.extendedProps.recurrence.endRecur,
    };
}


export { updateEvent, deleteEvent, prepareRecurringUpdate };

