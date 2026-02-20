// File: components/calendar/modal/validation.mjs
// Validation logic for modal form
// Specs: Section 4.11.4

// ================================
// CONSTANTS
// ================================
const ALLOWED_KINDS = new Set(['browser_source', 'image_source', 'media_source', 'vlc_source', 'ffmpeg_source']);

const URI_HINTS = {
    browser_source: 'HTTP(S) URL or local file path.',
    image_source: 'Image file path or HTTP(S) URL.',
    vlc_source: 'File/URL or playlist.',
    ffmpeg_source: 'FFmpeg-compatible file/URL (e.g., rtmp://).'
};

// ================================
// VALIDATION FUNCTIONS
// ================================

/**
 * Parse and validate a JSON field
 * @param {string} raw - Raw JSON string
 * @param {string} label - Field label for error messages
 * @returns {Object|null} Parsed object or null if invalid
 */
export function parseJsonField(raw, label) {
    const text = (raw || '').trim();
    if (text === '' || text === '{}') return {};

    try {
        const obj = JSON.parse(text);
        if (obj && typeof obj === 'object') return obj;
        alert(`${label} must be a valid JSON object.`);
        return null;
    } catch (e) {
        alert(`${label} contains invalid JSON.\n\n${e.message}`);
        return null;
    }
}

/**
 * Validate form data
 * @param {Object} formData - Form data to validate
 * @returns {Object|null} Validated data or null if invalid
 */
export function validateForm(formData) {
    // Validate title
    if (!formData.title || !formData.title.trim()) {
        alert('Title is required.');
        return null;
    }

    // Validate input kind
    const kind = String(formData.extendedProps.inputKind || '').trim();
    if (!ALLOWED_KINDS.has(kind)) {
        alert('Input Kind is invalid.');
        return null;
    }

    // Validate URI
    const uri = String(formData.extendedProps.inputUri || '').trim();
    if (!uri) {
        alert(`URI is required for input kind "${kind}".\n\n${URI_HINTS[kind]}`);
        return null;
    }

    // Validate input name
    if (!formData.extendedProps.inputName || !formData.extendedProps.inputName.trim()) {
        alert('Input Name is required.');
        return null;
    }

    // Validate times
    if (!formData.extendedProps.recurrence || Object.keys(formData.extendedProps.recurrence).length === 0) {
        // For non-recurring events, validate start and end dates
        if (!formData.start || !formData.end) {
            alert('Start and End times are required.');
            return null;
        }

        const start = new Date(formData.start);
        const end = new Date(formData.end);

        if (isNaN(start.getTime()) || isNaN(end.getTime())) {
            alert('Invalid date format.');
            return null;
        }

        if (start >= end) {
            alert('Start time must be before End time.');
            return null;
        }
    } else {
        // For recurring events, validate days of week
        if (!formData.extendedProps.recurrence.daysOfWeek ||
            formData.extendedProps.recurrence.daysOfWeek.length === 0) {
            alert('Please select at least one day for recurring events.');
            return null;
        }
    }

    return formData;
}

/**
 * Get URI hint for input kind
 * @param {string} kind - Input kind
 * @returns {string} URI hint text
 */
export function getUriHint(kind) {
    return URI_HINTS[kind] || 'Provide a valid file/URL.';
}

/**
 * Check if input kind is allowed
 * @param {string} kind - Input kind to check
 * @returns {boolean}
 */
export function isAllowedKind(kind) {
    return ALLOWED_KINDS.has(kind);
}

/**
 * Get all allowed input kinds
 * @returns {Array<string>}
 */
export function getAllowedKinds() {
    return Array.from(ALLOWED_KINDS);
}
