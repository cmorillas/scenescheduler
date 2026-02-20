// File: components/calendar/grid-actions.mjs
import { updateEvent, prepareRecurringUpdate } from './calendar-events.mjs';

/**
 * Handles the update of a recurring event series when an instance is moved or resized.
 * @param {Calendar} calendar - The FullCalendar instance.
 * @param {EventApi} event - The specific event instance that was acted upon.
 * @param {Object} info - The information object from the event callback.
 */
function handleRecurringSeriesUpdate(calendar, event, info) {
    const userConfirmed = confirm('This will update all events in this recurring series. Continue?');
    if (!userConfirmed) {
        info.revert();
        return;
    }

    // 1. Prepare the complete, updated event data object.
    const newEventData = prepareRecurringUpdate(event, info);

    // 2. Call the centralized update function.
    updateEvent(calendar, newEventData, event);
}

/**
 * Handles the update of a simple (non-recurring) event.
 * @param {EventApi} event - The event that was acted upon.
 */
function handleSimpleEventUpdate(event) {
    // In a real application, you would typically make an API call here.
}

/**
 * The main entry point for handling grid actions like dropping or resizing an event.
 */
export function handleGridAction(calendar, info) {
    const isRecurring = !!(info.event.extendedProps?.recurrence && Object.keys(info.event.extendedProps.recurrence).length > 0);

    if (isRecurring) {
        handleRecurringSeriesUpdate(calendar, info.event, info);
    } else {
        handleSimpleEventUpdate(info.event);
    }
}

