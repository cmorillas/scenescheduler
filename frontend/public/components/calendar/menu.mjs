// File: components/calendar/menu.mjs
// Create and manage calendar menu dropdown
// Specs: Section 4.9

import { handleMenuAction } from './menu-actions.mjs';

// ================================
// STATE
// ================================
let menuElement = null;
let calendarInstance = null;

// ================================
// MENU CREATION
// ================================

/**
 * Create menu DOM and attach handlers
 * @param {Calendar} calendar - FullCalendar instance
 */
export function createMenu(calendar) {
    calendarInstance = calendar;

    // Create menu element
    menuElement = document.createElement('div');
    menuElement.id = 'calendar-menu';
    menuElement.className = 'calendar-menu hidden';
    menuElement.innerHTML = `
        <div class="menu-section">
            <div class="menu-item" data-action="new">New Schedule</div>
            <div class="menu-item" data-action="load-local">Load from File</div>
            <div class="menu-item" data-action="save-local">Save to File</div>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-section">
            <div class="menu-item" data-action="get-server">Get from Server</div>
            <div class="menu-item" data-action="commit-server">Commit to Server</div>
        </div>
    `;

    // Attach to document body
    document.body.appendChild(menuElement);

    // Setup event listeners
    setupMenuListeners();
}

// ================================
// EVENT LISTENERS
// ================================

function setupMenuListeners() {
    // Handle menu item clicks
    menuElement.addEventListener('click', (e) => {
        const actionItem = e.target.closest('.menu-item');
        if (!actionItem) return;

        const action = actionItem.dataset.action;
        if (action && calendarInstance) {
            handleMenuAction(calendarInstance, action);
            hideMenu();
        }
    });

    // Close menu when clicking outside
    document.addEventListener('click', (e) => {
        const menuButton = document.querySelector('.fc-menu-button');
        if (!menuElement) return;

        // Don't close if clicking the menu or the button
        if (menuElement.contains(e.target) || e.target === menuButton) return;

        hideMenu();
    });
}

// ================================
// MENU VISIBILITY
// ================================

/**
 * Toggle menu visibility
 */
export function toggleMenu() {
    const btn = document.querySelector('.fc-menu-button');
    const menu = document.getElementById('calendar-menu');
    if (!btn || !menu) return;

    if (menu.classList.contains('hidden')) {
        showMenu();
    } else {
        hideMenu();
    }
}

/**
 * Show menu (position relative to button)
 */
function showMenu() {
    const btn = document.querySelector('.fc-menu-button');
    const menu = document.getElementById('calendar-menu');
    if (!btn || !menu) return;

    const rect = btn.getBoundingClientRect();

    // Position menu below button, aligned to right edge
    menu.style.top = `${rect.bottom + 6}px`;
    menu.style.right = `${window.innerWidth - rect.right}px`;
    menu.style.left = 'auto';

    menu.classList.remove('hidden');
}

/**
 * Hide menu
 */
function hideMenu() {
    const menu = document.getElementById('calendar-menu');
    if (!menu) return;
    menu.classList.add('hidden');
}
