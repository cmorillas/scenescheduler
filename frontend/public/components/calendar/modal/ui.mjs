// File: components/calendar/modal/ui.mjs
// UI interactions for modal (tabs, drag, colors)
// Specs: Section 4.11.3

import { calculateDuration, toggleRecurring, updateUriHint, autoGenerateSourceName } from './form.mjs';
import { cleanupPreview } from './preview.mjs';

// ================================
// DOM ELEMENT CACHE
// ================================
const dom = {
    modal: document.getElementById('task-modal'),
    content: document.getElementById('draggable-content'),
    header: document.getElementById('drag-handle'),
    form: document.getElementById('task-form'),
    start: document.getElementById('task-start'),
    end: document.getElementById('task-end'),
    title: document.getElementById('task-title'),
    recurChk: document.getElementById('recurring-toggle'),
    inputKind: document.getElementById('task-input-kind'),
    tabNav: document.querySelector('.tab-nav'),
    tabPanes: document.querySelectorAll('.tab-pane'),
    colorInputs: document.querySelectorAll('.custom-color-input'),
    closeBtn: document.getElementById('close-modal-btn'),
    cancelBtn: document.getElementById('cancel-task'),
    deleteBtn: document.getElementById('delete-task')
};

// ================================
// DRAG STATE
// ================================
let isDragging = false;
let offsetX = 0;
let offsetY = 0;

// ================================
// UI HANDLER INITIALIZATION
// ================================

/**
 * Initialize all UI event handlers
 * @param {Function} onSave - Save callback
 * @param {Function} onDelete - Delete callback
 * @param {Function} onClose - Close callback
 */
export function initUIHandlers(onSave, onDelete, onClose) {
    // Modal close handlers
    dom.closeBtn.addEventListener('click', onClose);
    dom.cancelBtn.addEventListener('click', onClose);
    dom.modal.addEventListener('click', e => e.target === dom.modal && onClose());
    document.addEventListener('keydown', e => e.key === 'Escape' && onClose());

    // Form submit and delete
    dom.form.addEventListener('submit', onSave);
    dom.deleteBtn.addEventListener('click', onDelete);

    // Drag handlers
    dom.header.addEventListener('mousedown', startDrag);

    // Form field handlers
    dom.start.addEventListener('change', calculateDuration);
    dom.end.addEventListener('change', calculateDuration);
    dom.title.addEventListener('input', autoGenerateSourceName);
    dom.recurChk.addEventListener('change', toggleRecurring);
    dom.inputKind.addEventListener('change', () => updateUriHint(dom.inputKind.value));

    // Tab navigation
    initTabNavigation();

    // Color picker handlers
    initColorPickers();
}

// ================================
// TAB NAVIGATION
// ================================

function initTabNavigation() {
    if (!dom.tabNav) return;

    dom.tabNav.addEventListener('click', (e) => {
        const targetBtn = e.target.closest('.tab-btn');
        if (!targetBtn) return;

        // Get currently active tab before switching
        const activeTab = document.querySelector('.tab-pane.active');
        const isLeavingPreviewTab = activeTab && activeTab.id === 'preview-tab';

        // If leaving preview tab, cleanup preview
        if (isLeavingPreviewTab) {
            cleanupPreview();
        }

        // Remove active class from all tabs and panes
        dom.tabNav.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        dom.tabPanes.forEach(pane => pane.classList.remove('active'));

        // Add active class to clicked tab and corresponding pane
        targetBtn.classList.add('active');
        const targetPane = document.querySelector(targetBtn.dataset.target);
        if (targetPane) targetPane.classList.add('active');
    });
}

// ================================
// COLOR PICKER INTERACTIONS
// ================================

function initColorPickers() {
    dom.colorInputs.forEach(wrapper => {
        const input = wrapper.querySelector('input[type="color"]');
        const clearBtn = wrapper.querySelector('.clear-color-btn');

        if (clearBtn) {
            clearBtn.addEventListener('click', () => {
                wrapper.dataset.cleared = 'true';
            });
        }

        if (input) {
            input.addEventListener('input', () => {
                delete wrapper.dataset.cleared;
            });
        }
    });
}

// ================================
// DRAG AND DROP FUNCTIONALITY
// ================================

function startDrag(e) {
    // Don't start drag if clicking on interactive elements
    if (e.target.closest('button, input, select, textarea, .slider')) return;

    isDragging = true;
    dom.header.style.cursor = 'grabbing';

    const rect = dom.content.getBoundingClientRect();
    offsetX = e.clientX - rect.left;
    offsetY = e.clientY - rect.top;

    document.addEventListener('mousemove', moveDrag);
    document.addEventListener('mouseup', endDrag);
    e.preventDefault();
}

function moveDrag(e) {
    if (!isDragging) return;

    let x = e.clientX - offsetX;
    let y = e.clientY - offsetY;

    // Keep modal within viewport bounds
    const rect = dom.content.getBoundingClientRect();
    x = Math.max(0, Math.min(window.innerWidth - rect.width, x));
    y = Math.max(0, Math.min(window.innerHeight - rect.height, y));

    dom.content.style.left = `${x}px`;
    dom.content.style.top = `${y}px`;
    dom.content.style.transform = 'none';
}

function endDrag() {
    isDragging = false;
    dom.header.style.cursor = 'grab';
    document.removeEventListener('mousemove', moveDrag);
    document.removeEventListener('mouseup', endDrag);
}

// ================================
// PUBLIC API
// ================================

/**
 * Reset drag state (useful when closing modal)
 */
export function resetDragState() {
    isDragging = false;
    dom.header.style.cursor = 'grab';
    document.removeEventListener('mousemove', moveDrag);
    document.removeEventListener('mouseup', endDrag);
}

/**
 * Activate first tab (useful when opening modal)
 */
export function activateFirstTab() {
    if (!dom.tabNav) return;

    dom.tabNav.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    dom.tabPanes.forEach(pane => pane.classList.remove('active'));

    const firstBtn = dom.tabNav.querySelector('.tab-btn');
    if (firstBtn) {
        firstBtn.classList.add('active');
        const firstPane = document.querySelector(firstBtn.dataset.target);
        if (firstPane) firstPane.classList.add('active');
    }
}
