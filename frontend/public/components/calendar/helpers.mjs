// =========================================================
// File: components/calendar/helpers.mjs
// A collection of utility functions for date, time, and string manipulation.
// =========================================================

/**
 * Pads a number with a leading zero if it's less than 10.
 * @param {number} n - The number to pad.
 * @returns {string}
 */
export function pad2(n) {
  return String(n).padStart(2, '0');
}

/**
 * Generates a short, unique ID for UI purposes.
 * @param {string} [prefix='evt-'] - The prefix for the ID.
 * @returns {string}
 */
export function genId(prefix = 'evt-') {
  return `${prefix}${Math.random().toString(36).slice(2, 9)}`;
}

// ================================
// TIME FORMATTING
// ================================

/**
 * Ensures a time string is in "HH:MM:SS" format.
 * @param {string} s - The time string (e.g., "9:5", "09:05", "09:05:30").
 * @returns {string}
 */
export function ensureHHMMSS(s = '0:0:0') {
  const parts = String(s).split(':');
  const h = parts[0] || '0';
  const m = parts[1] || '0';
  const sec = parts[2] || '0';
  return `${pad2(h)}:${pad2(m)}:${pad2(sec)}`;
}

/**
 * Converts a time string "HH:MM:SS" to total milliseconds.
 * @param {string} timeStr - The time string.
 * @returns {number}
 */
export function timeToMs(timeStr) {
  const [h, m, s] = ensureHHMMSS(timeStr).split(':').map(Number);
  return (h * 3600 + m * 60 + s) * 1000;
}

/**
 * Converts total milliseconds to a time string "HH:MM:SS".
 * @param {number} totalMs - Total milliseconds.
 * @returns {string}
 */
export function msToTime(totalMs) {
  const totalSeconds = Math.floor(totalMs / 1000);
  const h = Math.floor(totalSeconds / 3600) % 24;
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  return `${pad2(h)}:${pad2(m)}:${pad2(s)}`;
}

// ================================
// DATE FORMATTING
// ================================

/**
 * Formats a Date object or date string to "YYYY-MM-DD".
 * @param {Date|string} value - The date to format.
 * @returns {string}
 */
export function toLocalDateString(value) {
  if (!value) return '';
  const d = (value instanceof Date) ? value : new Date(value);
  if (isNaN(d.getTime())) return '';
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;
}

/**
 * Formats a Date object or date string for use in `<input type="datetime-local">`.
 * @param {Date|string} value - The date to format.
 * @returns {string}
 */
export function toLocalDateTimeString(value) {
    if (!value) return '';
    const d = (value instanceof Date) ? value : new Date(value);
    if (isNaN(d.getTime())) return '';
    return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}T${pad2(d.getHours())}:${pad2(d.getMinutes())}`;
}

/**
 * Returns today's date as "YYYY-MM-DD".
 * @returns {string}
 */
export function formatTodayYYYYMMDD() {
  return toLocalDateString(new Date());
}

/**
 * Splits a Date object into its date and time parts.
 * @param {Date|string} value - The date to split.
 * @returns {{date: string, time: string}}
 */
export function splitLocal(value) {
  if (!value) return { date: '', time: '00:00:00' };
  const d = (value instanceof Date) ? value : new Date(value);
  if (isNaN(d.getTime())) return { date: '', time: '00:00:00' };
  const date = toLocalDateString(d);
  const time = `${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`;
  return { date, time };
}

/**
 * Adds seconds to a local ISO-like date string.
 * @param {string} localIso - The date string (e.g., "2025-08-29T10:00:00").
 * @param {number} seconds - The number of seconds to add.
 * @returns {string|null}
 */
export function addSecondsLocalISO(localIso, seconds) {
  if (!localIso) return null;
  const d = new Date(localIso);
  if (isNaN(d.getTime())) return null;
  d.setSeconds(d.getSeconds() + Number(seconds || 0));
  return `${toLocalDateString(d)}T${splitLocal(d).time}`;
}

// ================================
// DAY OF WEEK CONVERSIONS
// ================================

const DOW_MAP = {
  NUM_TO_NAME: ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'],
  NAME_TO_NUM: { SUN: 0, MON: 1, TUE: 2, WED: 3, THU: 4, FRI: 5, SAT: 6 }
};

export function daysOfWeekNumsToNames(nums = []) {
  return nums.map(n => DOW_MAP.NUM_TO_NAME[n]).filter(Boolean);
}

export function weekdaysNamesToNums(names = []) {
  return names.map(n => DOW_MAP.NAME_TO_NUM[String(n).toUpperCase()]).filter(n => n !== undefined);
}

/**
 * Kept for compatibility. Returns time in "HH:MM:SS" format from a Date.
 */
export function dateToHHMMSS(date) {
    return splitLocal(date).time;
}

// ================================
// CALENDAR UI HELPERS
// ================================

/**
 * Converts a 'HH:mm:ss' duration string to milliseconds.
 * @param {string} durationString - The duration string (e.g., "01:00:00").
 * @returns {number} The duration in milliseconds.
 */
export function durationStringToMs(durationString) {
    if (typeof durationString !== 'string') return 3600000; // Default to 1 hour
    const parts = durationString.split(':').map(Number);
    const hours = parts[0] || 0;
    const minutes = parts[1] || 0;
    const seconds = parts[2] || 0;
    return (hours * 3600 + minutes * 60 + seconds) * 1000;
}

/**
 * Applies the custom left border style to an event element if it's recurring.
 * @param {EventApi} event - The FullCalendar event object.
 * @param {HTMLElement} el - The event's HTML element.
 */
export function applyRecurringEventStyles(event, el) {
    const isRecurring = event.extendedProps?.recurrence && Object.keys(event.extendedProps.recurrence).length > 0;
    if (isRecurring) {
        const color = event.borderColor || 'var(--primary)';
        el.style.border = 'none'; // First, reset any border set by FullCalendar
        el.style.borderLeft = `8px solid ${color}`; // Then, apply our specific border
    }
}

