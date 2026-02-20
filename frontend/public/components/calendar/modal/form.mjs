// File: components/calendar/modal/form.mjs
// Form data population and extraction
// Specs: Section 4.11.2

import { toLocalDateTimeString, toLocalDateString, formatTodayYYYYMMDD, ensureHHMMSS } from '../helpers.mjs';
import { parseJsonField, getUriHint, isAllowedKind } from './validation.mjs';

// ================================
// DOM ELEMENT CACHE
// ================================
const dom = {
    form: document.getElementById('task-form'),
    title: document.getElementById('task-title'),
    description: document.getElementById('task-description'),
    textColor: document.getElementById('task-text-color'),
    backgroundColor: document.getElementById('task-background-color'),
    borderColor: document.getElementById('task-border-color'),
    tags: document.getElementById('task-tags'),
    classNames: document.getElementById('task-classnames'),
    start: document.getElementById('task-start'),
    end: document.getElementById('task-end'),
    duration: document.getElementById('task-duration'),
    inputName: document.getElementById('task-input-name'),
    inputKind: document.getElementById('task-input-kind'),
    inputUri: document.getElementById('task-input-uri'),
    uriHint: document.getElementById('uri-hint'),
    inputSettings: document.getElementById('task-input-settings'),
    recurChk: document.getElementById('recurring-toggle'),
    recurBox: document.getElementById('recurring-fields'),
    recurStart: document.getElementById('recurring-start'),
    recurEnd: document.getElementById('recurring-end'),
    enabled: document.getElementById('task-enabled'),
    preload: document.getElementById('task-preload'),
    onEnd: document.getElementById('task-onend'),
    transform: document.getElementById('task-transform'),
    weekdayCheckboxes: document.querySelectorAll('.weekdays-selector input[type="checkbox"]'),
    colorInputs: document.querySelectorAll('.custom-color-input'),
    tabNav: document.querySelector('.tab-nav'),
    tabPanes: document.querySelectorAll('.tab-pane'),
    filePreview: document.getElementById('file-preview')
};

// ================================
// FORM POPULATION
// ================================

/**
 * Populate form fields from event object
 * @param {Object} event - FullCalendar event object
 */
export function populateForm(event) {
    const ext = event.extendedProps || {};
    const isRecurring = !!(ext.recurrence && Object.keys(ext.recurrence).length > 0);

    // General Tab
    dom.title.value = event.title || '';
    dom.description.value = ext.description || '';
    dom.tags.value = (ext.tags || []).join(' ');
    dom.classNames.value = (event.classNames || []).filter(c => c !== 'recurring-event').join(' ');

    setColorInputValue(dom.textColor, event.textColor || '#f8fafc');
    setColorInputValue(dom.backgroundColor, event.backgroundColor || '#3b82f6');
    setColorInputValue(dom.borderColor, event.borderColor || '#ffffff');

    // Source Tab
    dom.inputName.value = ext.inputName || event.title || '';
    dom.inputKind.value = isAllowedKind(ext.inputKind) ? ext.inputKind : 'browser_source';
    updateUriHint(dom.inputKind.value);
    dom.inputUri.value = ext.inputUri || '';
    dom.inputSettings.value = JSON.stringify(ext.inputSettings || {}, null, 2);
    dom.transform.value = Object.keys(ext.transform || {}).length ? JSON.stringify(ext.transform, null, 2) : '';

    // Behavior Tab
    dom.enabled.checked = ext.enabled ?? true;
    dom.preload.value = Number(ext.automation?.preloadSeconds ?? 0);
    dom.onEnd.value = ext.automation?.onEndAction || 'hide';

    // Timing & Recurrence Tab
    dom.recurChk.checked = isRecurring;
    toggleRecurring(); // Show/hide recurring fields based on checkbox state

    if (isRecurring) {
        const recData = ext.recurrence || {};
        const baseDate = event.start ? toLocalDateString(event.start) : formatTodayYYYYMMDD();
        setDateTimeLocal(dom.start, `${baseDate}T${ensureHHMMSS(recData.startTime)}`);
        setDateTimeLocal(dom.end, `${baseDate}T${ensureHHMMSS(recData.endTime)}`);
        dom.recurStart.value = recData.startRecur || '';
        dom.recurEnd.value = recData.endRecur || '';
        dom.weekdayCheckboxes.forEach(cb => {
            cb.checked = (recData.daysOfWeek || []).includes(parseInt(cb.value, 10));
        });
    } else {
        if (event.start) setDateTimeLocal(dom.start, event.start);
        if (event.end) setDateTimeLocal(dom.end, event.end);
    }

    calculateDuration();
}

/**
 * Reset form to initial state
 */
export function resetForm() {
    dom.form.reset();
    dom.duration.value = '';

    if (dom.filePreview) {
        dom.filePreview.src = '';
        dom.filePreview.poster = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='450' height='253' viewBox='0 0 450 253'%3E%3Crect width='100%' height='100%' fill='%23000'%3E%3C/rect%3E%3Ctext x='50%' y='50%' font-family='sans-serif' font-size='16' fill='%23FFF' text-anchor='middle' dominant-baseline='middle'%3EPreview%3C/text%3E%3C/svg%3E";
    }

    dom.colorInputs.forEach(wrapper => delete wrapper.dataset.cleared);
    if (dom.title && dom.title.dataset) {
        delete dom.title.dataset.prevSlug;
    }

    // Reset recurring fields visibility
    dom.recurChk.checked = false;
    dom.recurBox.style.display = 'none';
    dom.recurBox.classList.remove('show');

    if (dom.tabNav) {
        dom.tabNav.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        dom.tabPanes.forEach(pane => pane.classList.remove('active'));
        const firstBtn = dom.tabNav.querySelector('.tab-btn');
        if (firstBtn) {
            firstBtn.classList.add('active');
            const firstPane = document.querySelector(firstBtn.dataset.target);
            if (firstPane) firstPane.classList.add('active');
        }
    }
}

/**
 * Set default values for new task
 * @param {string} startStr - Start datetime string
 * @param {string} endStr - End datetime string
 */
export function setNewTaskDefaults(startStr = null, endStr = null) {
    let startTime, endTime;
    if (startStr) {
        startTime = new Date(startStr);
        endTime = endStr ? new Date(endStr) : new Date(startTime.getTime() + 60 * 60 * 1000);
    } else {
        const now = new Date();
        startTime = new Date(now);
        startTime.setMinutes(Math.ceil(startTime.getMinutes() / 15) * 15, 0, 0);
        endTime = new Date(startTime.getTime() + 60 * 60 * 1000);
    }
    setDateTimeLocal(dom.start, startTime);
    setDateTimeLocal(dom.end, endTime);
    calculateDuration();

    // Explicitly set default colors to ensure consistency
    setColorInputValue(dom.textColor, '#f8fafc');
    setColorInputValue(dom.backgroundColor, '#3b82f6');
    setColorInputValue(dom.borderColor, '#22c55e');

    dom.inputKind.value = 'browser_source';
    updateUriHint(dom.inputKind.value);
    dom.enabled.checked = true;
    dom.preload.value = 0;
    dom.onEnd.value = 'hide';

    // Initialize recurring toggle state
    toggleRecurring();
}

// ================================
// FORM EXTRACTION
// ================================

/**
 * Extract event data from form fields
 * @returns {Object|null} Event data object or null if validation fails
 */
export function extractFormData() {
    const settingsObj = parseJsonField(dom.inputSettings.value, 'Settings (JSON)');
    const transformObj = parseJsonField(dom.transform.value, 'Transform (JSON)');
    if (settingsObj === null || transformObj === null) return null;

    const eventData = {
        title: dom.title.value,
        classNames: parseTags(dom.classNames.value),
        textColor: getColorInputValue(dom.textColor),
        backgroundColor: getColorInputValue(dom.backgroundColor),
        borderColor: getColorInputValue(dom.borderColor),
        extendedProps: {
            description: dom.description.value,
            tags: parseTags(dom.tags.value),
            enabled: dom.enabled.checked,
            automation: {
                onEndAction: dom.onEnd.value,
                preloadSeconds: Number(dom.preload.value || 0)
            },
            inputName: dom.inputName.value,
            inputKind: dom.inputKind.value,
            inputUri: dom.inputUri.value,
            inputSettings: settingsObj,
            transform: transformObj,
            recurrence: {}
        }
    };

    if (dom.recurChk.checked) {
        const recurrenceData = {
            daysOfWeek: Array.from(dom.weekdayCheckboxes).filter(cb => cb.checked).map(cb => parseInt(cb.value, 10)),
            startTime: ensureHHMMSS(dom.start.value.split('T')[1]),
            endTime: ensureHHMMSS(dom.end.value.split('T')[1]),
            startRecur: dom.recurStart.value || null,
            endRecur: dom.recurEnd.value || null
        };
        eventData.extendedProps.recurrence = recurrenceData;

        Object.assign(eventData, {
            daysOfWeek: recurrenceData.daysOfWeek,
            startTime: recurrenceData.startTime,
            endTime: recurrenceData.endTime,
            startRecur: recurrenceData.startRecur,
            endRecur: recurrenceData.endRecur
        });
        eventData.classNames.push('recurring-event');
    } else {
        Object.assign(eventData, {
            start: new Date(dom.start.value),
            end: new Date(dom.end.value)
        });
    }

    return eventData;
}

// ================================
// HELPER FUNCTIONS
// ================================

/**
 * Calculate and update duration display
 */
export function calculateDuration() {
    if (dom.start.value && dom.end.value) {
        const diff = new Date(dom.end.value) - new Date(dom.start.value);
        if (diff >= 0) {
            const s = Math.floor(diff / 1000);
            const h = Math.floor(s / 3600);
            const m = Math.floor((s % 3600) / 60);
            const pad = n => String(n).padStart(2, '0');
            dom.duration.value = `${pad(h)}:${pad(m)}:${pad(s % 60)}`;
        } else {
            dom.duration.value = 'Invalid';
        }
    }
}

/**
 * Toggle recurring fields visibility
 */
export function toggleRecurring() {
    const isRecur = dom.recurChk.checked;

    if (isRecur) {
        dom.recurBox.classList.add('show');
        dom.recurBox.style.display = 'flex';
    } else {
        dom.recurBox.classList.remove('show');
        dom.recurBox.style.display = 'none';
    }
}

/**
 * Update URI hint based on input kind
 * @param {string} kind - Input kind
 */
export function updateUriHint(kind) {
    const hint = getUriHint(kind);
    dom.uriHint.textContent = hint;
    dom.inputUri.placeholder = hint;
}

/**
 * Auto-generate source name from title
 */
export function autoGenerateSourceName() {
    if (!dom.title || !dom.inputName) return;
    const currentTitle = dom.title.value;
    const currentSourceName = dom.inputName.value;
    const slug = currentTitle.toLowerCase().trim().replace(/\s+/g, '_').replace(/[^\w-]+/g, '');
    const prevSlug = dom.title.dataset.prevSlug || '';
    if (currentSourceName === '' || currentSourceName === prevSlug) {
        dom.inputName.value = slug;
    }
    dom.title.dataset.prevSlug = slug;
}

// ================================
// PRIVATE HELPERS
// ================================

function parseTags(s) {
    return Array.from(new Set(String(s || '').split(/\s+/).filter(Boolean).map(t => t.trim())));
}

function setDateTimeLocal(inputEl, value) {
    if (!inputEl) return;
    inputEl.value = toLocalDateTimeString(value);
}

function getColorInputValue(inputEl) {
    const wrapper = inputEl.parentElement;
    return wrapper.dataset.cleared === 'true' ? '' : inputEl.value;
}

function setColorInputValue(inputEl, color) {
    const wrapper = inputEl.parentElement;
    if (color && color !== 'transparent') {
        inputEl.value = color;
        delete wrapper.dataset.cleared;
    } else {
        wrapper.dataset.cleared = 'true';
    }
}
