// File: frontend/components/calendar/schedule-adapter.mjs
//
// This module acts as an adapter between FullCalendar's event objects and the
// canonical Schedule v1.0 JSON format. It is responsible for all data
// transformations required for importing from and exporting to the backend.
//
// The entire schedule is wrapped in an object containing metadata:
// {
//   "version": "1.0",
//   "scheduleName": "Schedule",
//   "schedule": [
//     {
//       "id": "string",
//       "title": "string",
//       "enabled": boolean,
//       "general": {
//         "description": "string",
//         "tags": ["string"],
//         "classNames": ["string"],
//         "textColor": "string",
//         "backgroundColor": "string",
//         "borderColor": "string"
//       },
//       "source": {
//         "name": "string", // This is the technical source name
//         "inputKind": "string",
//         "uri": "string",
//         "inputSettings": {},
//         "transform": {}
//       },
//       "timing": {
//         "start": "YYYY-MM-DDTHH:MM:SSZ",
//         "end": "YYYY-MM-DDTHH:MM:SSZ",
//         "isRecurring": boolean,
//         "recurrence": {
//           "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"],
//           "startRecur": "YYYY-MM-DD",
//           "endRecur": "YYYY-MM-DD"
//         }
//       },
//       "behavior": {
//         "onEndAction": "string",
//         "preloadSeconds": number
//       }
//     }
//   ]
// }
//
// Order of sections:
// 1) Public API
// 2) Mappers (Event -> Schedule Item, Schedule Item -> Event)
// 3) Local Helpers

import {
  genId,
  ensureHHMMSS,
  daysOfWeekNumsToNames,
  weekdaysNamesToNums
} from './helpers.mjs';

// =============================
// PUBLIC API
// =============================

/**
 * Build a Schedule 1.0 object from current FullCalendar events.
 */
export function exportSchedule(
  calendar,
  { scheduleName = 'Schedule', version = '1.0' } = {}
) {
  const singles = [];
  const seriesMap = new Map();

  for (const ev of calendar.getEvents()) {
    const item = eventToScheduleItem(ev);
    if (!item) continue;

    if (item.timing.isRecurring) {
      const key = buildSeriesKey(item);
      if (!seriesMap.has(key)) seriesMap.set(key, item);
    } else {
      singles.push(item);
    }
  }

  const schedule = [...singles, ...seriesMap.values()];
  return { version, scheduleName, schedule };
}

/**
 * Replace all FullCalendar events with items from a Schedule 1.0 JSON.
 */
export function importSchedule(calendar, scheduleJson) {
  if (!scheduleJson || !Array.isArray(scheduleJson.schedule)) return;
  const inputs = scheduleJson.schedule.map(scheduleItemToEvent);
  calendar.removeAllEvents();
  calendar.addEventSource(inputs);
}

// =============================
// MAPPERS
// =============================

/**
 * FullCalendar EventApi -> Schedule 1.0 item.
 */
function eventToScheduleItem(ev) {
  const xp = ev.extendedProps || {};
  const isRecurring = !!(xp.recurrence && Object.keys(xp.recurrence).length > 0);

  const base = {
    id: ev.id || genId(),
    title: ev.title || '',
    enabled: Boolean(xp.enabled ?? true),
    general: {
        description: (xp.description ?? "").toString(),
        tags: Array.isArray(xp.tags) ? xp.tags : [],
        classNames: (ev.classNames || []).filter(c => c !== 'recurring-event'),
        textColor: ev.textColor || "",
        backgroundColor: ev.backgroundColor || "",
        borderColor: ev.borderColor || ""
    },
    source: {
        name: xp.inputName || ev.title || '',
        inputKind: xp.inputKind || 'browser_source',
        uri: (xp.inputUri ?? "").toString(),
        inputSettings: xp.inputSettings || {},
        transform: (xp.transform && typeof xp.transform === 'object') ? xp.transform : {}
    },
    behavior: {
        onEndAction: xp.automation?.onEndAction ?? 'hide',
        preloadSeconds: Number(xp.automation?.preloadSeconds ?? 0),
    }
  };
  
  const timing = {
      // For non-recurring events, convert dates to UTC ISO string with 'Z'
      start: ev.start ? ev.start.toISOString() : null,
      end: ev.end ? ev.end.toISOString() : null,
      isRecurring: isRecurring,
      recurrence: {
          daysOfWeek: [],
          startRecur: "",
          endRecur: ""
      }
  };

  if (isRecurring) {
    const recData = xp.recurrence;
    
    // For recurring events, the date is a template. We build the string and append 'Z'
    // to conform to the backend's strict UTC format requirement.
    const baseDate = recData.startRecur || '1970-01-01'; // Use a fixed date for consistency
    timing.start = `${baseDate}T${ensureHHMMSS(recData.startTime)}Z`;
    timing.end = `${baseDate}T${ensureHHMMSS(recData.endTime)}Z`;

    timing.recurrence = {
      daysOfWeek: daysOfWeekNumsToNames(recData.daysOfWeek || []),
      startRecur: recData.startRecur || "",
      endRecur: recData.endRecur || "",
    };
  }

  return { ...base, timing };
}


/**
 * Schedule 1.0 item -> FullCalendar EventInput.
 */
function scheduleItemToEvent(item) {
  const general = item.general || {};
  const source = item.source || {};
  const timing = item.timing || {};
  const behavior = item.behavior || {};
  const isRecurring = timing.isRecurring;

  const extendedProps = {
    description: (general.description ?? "").toString(),
    enabled: Boolean(item.enabled ?? true),
    tags: Array.isArray(general.tags) ? general.tags : [],
    automation: {
      onEndAction: behavior.onEndAction ?? 'hide',
      preloadSeconds: Number(behavior.preloadSeconds ?? 0)
    },
    inputName: source.name || item.title || '',
    inputKind: source.inputKind || 'browser_source',
    inputUri: (source.uri ?? "").toString(),
    inputSettings: source.inputSettings || {},
    transform: (source.transform && typeof source.transform === 'object') ? source.transform : {},
    recurrence: {}
  };
  
  const classNames = Array.isArray(general.classNames) ? general.classNames : [];
  
  const baseEvent = {
      id: item.id || genId(),
      title: item.title || '',
      textColor: general.textColor || undefined,
      backgroundColor: general.backgroundColor || undefined,
      borderColor: general.borderColor || undefined,
      extendedProps,
      classNames
  };

  if (isRecurring) {
    const rec = timing.recurrence || {};
    const daysOfWeekNumbers = weekdaysNamesToNums(rec.daysOfWeek || []);
    
    // Remove 'Z' for FullCalendar if present, as it handles timezones internally
    const startTime = ensureHHMMSS((timing.start || '').split('T')[1]?.replace('Z', ''));
    const endTime = ensureHHMMSS((timing.end || '').split('T')[1]?.replace('Z', ''));
    
    // Store our recurrence rules as the single source of truth
    extendedProps.recurrence = {
        daysOfWeek: daysOfWeekNumbers,
        startRecur: rec.startRecur,
        endRecur: rec.endRecur,
        startTime: startTime,
        endTime: endTime
    };

    return {
      ...baseEvent,
      daysOfWeek: daysOfWeekNumbers,
      startTime: startTime,
      endTime: endTime,
      startRecur: rec.startRecur,
      endRecur: rec.endRecur,
      classNames: uniq([...classNames, 'recurring-event'])
    };
  }

  // Non-recurring
  return {
    ...baseEvent,
    start: timing.start || null,
    end: timing.end || null,
    classNames: uniq(classNames)
  };
}


// =============================
// LOCAL HELPERS
// =============================

/**
 * Creates a stable JSON string key to identify a unique recurring series.
 */
function buildSeriesKey(item) {
    const { general, source, behavior, title, enabled, timing } = item;
    
    // The key is built using the canonical timing data from the master event,
    // which is now consistent for all instances.
    const keyData = {
        title,
        enabled,
        general,
        source,
        behavior,
        timing
    };
    return JSON.stringify(keyData);
}

/**
 * Returns a new array with unique values.
 */
function uniq(arr) {
  return Array.from(new Set((arr || []).filter(Boolean)));
}

