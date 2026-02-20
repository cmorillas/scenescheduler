// File: components/calendar/modal.mjs
// Modal coordinator - public API and lifecycle management
// Specs: Section 4.11.1

import { updateEvent, deleteEvent } from './calendar-events.mjs';
import { populateForm, resetForm, setNewTaskDefaults, extractFormData } from './modal/form.mjs';
import { validateForm } from './modal/validation.mjs';
import { initUIHandlers, resetDragState } from './modal/ui.mjs';
import { initPreview, updatePreviewInfo, cleanupPreview } from './modal/preview.mjs';

// ================================
// DOM ELEMENT CACHE
// ================================
const dom = {
    modal: document.getElementById('task-modal'),
    modalTitle: document.getElementById('modal-title-text')
};

// ================================
// STATE
// ================================
let activeEvent = null;
let localCalendarInstance = null;

// ================================
// PUBLIC API
// ================================

/**
 * Open modal for creating or editing an event
 * @param {Calendar} calendar - FullCalendar instance
 * @param {Object} options - Options object
 * @param {Object} options.event - Existing event to edit (optional)
 * @param {string} options.start - Start datetime for new event (optional)
 * @param {string} options.end - End datetime for new event (optional)
 * @param {boolean} options.readOnly - Open in read-only mode (optional)
 */
export function openTaskModal(calendar, options = {}) {
    const { event, start, end, readOnly = false } = options;
    localCalendarInstance = calendar;
    activeEvent = event || null;

    resetForm();

    if (activeEvent) {
        dom.modalTitle.textContent = readOnly ? 'Event Information' : 'Edit Task';
        populateForm(activeEvent);
    } else {
        dom.modalTitle.textContent = 'New Task';
        setNewTaskDefaults(start, end);
    }

    // Initialize preview module
    initPreview();

    // Update preview info with source data
    if (activeEvent) {
        const ext = activeEvent.extendedProps || {};
        updatePreviewInfo({
            inputKind: ext.inputKind || 'browser_source',
            uri: ext.inputUri || '',
            inputSettings: ext.inputSettings || {}
        });
    }

    // Apply readonly mode if specified
    if (readOnly) {
        applyReadOnlyMode();
    }

    dom.modal.style.display = 'flex';
}

/**
 * Close modal and reset state
 */
function closeModal() {
    // Cleanup preview (will stop if playing)
    cleanupPreview();

    dom.modal.style.display = 'none';
    resetDragState();
    removeReadOnlyMode(); // Clean up readonly mode if it was applied
    activeEvent = null;
    localCalendarInstance = null;
}

// ================================
// SAVE & DELETE LOGIC
// ================================

/**
 * Save task (create or update)
 * @param {Event} e - Form submit event
 */
function saveTask(e) {
    e?.preventDefault?.();
    if (!localCalendarInstance) return;

    // Extract form data
    const eventData = extractFormData();
    if (!eventData) return; // Validation failed in extraction

    // Validate form data
    const validatedData = validateForm(eventData);
    if (!validatedData) return; // Validation failed

    // Event data validated and ready to save

    // Use the centralized update function
    updateEvent(localCalendarInstance, validatedData, activeEvent);

    closeModal();
}

/**
 * Delete task
 */
function deleteTask() {
    // Call the centralized delete function
    if (deleteEvent(activeEvent)) {
        closeModal();
    }
}

// ================================
// READ-ONLY MODE
// ================================

/**
 * Apply read-only mode to the modal
 */
function applyReadOnlyMode() {
    // Disable all inputs, selects, and textareas
    const inputs = dom.modal.querySelectorAll('input, select, textarea');
    inputs.forEach(input => {
        input.disabled = true;
    });

    // Hide only Save and Delete buttons (keep tabs and close buttons visible)
    const saveBtn = document.getElementById('save-task');
    const deleteBtn = document.getElementById('delete-task');
    if (saveBtn) saveBtn.style.display = 'none';
    if (deleteBtn) deleteBtn.style.display = 'none';

    // Change cancel button to "Close"
    const cancelBtn = document.getElementById('cancel-task');
    if (cancelBtn) cancelBtn.textContent = 'Close';

    // Disable color picker buttons but keep them visible
    const colorButtons = dom.modal.querySelectorAll('.color-picker button');
    colorButtons.forEach(btn => {
        btn.disabled = true;
    });

    // Add readonly class to modal for styling
    dom.modal.classList.add('modal-readonly');
}

/**
 * Remove read-only mode from the modal
 */
function removeReadOnlyMode() {
    // Re-enable all inputs
    const inputs = dom.modal.querySelectorAll('input, select, textarea');
    inputs.forEach(input => {
        input.disabled = false;
    });

    // Re-enable and show Save and Delete buttons
    const saveBtn = document.getElementById('save-task');
    const deleteBtn = document.getElementById('delete-task');
    if (saveBtn) saveBtn.style.display = '';
    if (deleteBtn) deleteBtn.style.display = '';

    // Restore cancel button text
    const cancelBtn = document.getElementById('cancel-task');
    if (cancelBtn) cancelBtn.textContent = 'Cancel';

    // Re-enable color picker buttons
    const colorButtons = dom.modal.querySelectorAll('.color-picker button');
    colorButtons.forEach(btn => {
        btn.disabled = false;
    });

    // Remove readonly class
    dom.modal.classList.remove('modal-readonly');
}

// ================================
// INITIALIZATION
// ================================

/**
 * Initialize modal event listeners
 */
function initModal() {
    initUIHandlers(saveTask, deleteTask, closeModal);
}

// Initialize on script load
initModal();
