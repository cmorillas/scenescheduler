// File: components/calendar/zoom.mjs
// This module provides functionality for zooming in and out of the calendar view.

import { timeToMs, msToTime } from './helpers.mjs';

// An array of predefined zoom levels, from coarsest to finest.
const ZOOM_LEVELS = ['01:00:00', '00:30:00', '00:15:00', '00:05:00', '00:01:00', '00:00:30'];

/**
 * Calculates the time at the vertical center of the visible calendar area.
 * This is used to maintain the view's position when zooming.
 * @param {Calendar} calendar - The FullCalendar instance.
 * @returns {string} The time at the center in "HH:MM:SS" format.
 */
function getCenterTime(calendar) {
	// Try multiple selectors to find the scroller element
	let scrollerEl = calendar.el.querySelector('.fc-scroller.fc-scroller-liquid-absolute');
	if (!scrollerEl) {
		scrollerEl = calendar.el.querySelector('.fc-scroller');
	}
	if (!scrollerEl) {
		scrollerEl = calendar.el.querySelector('.fc-timegrid-body')?.parentElement;
	}
	if (!scrollerEl) return '00:00:00';

	const scrollTop = scrollerEl.scrollTop;
	const clientHeight = scrollerEl.clientHeight;
	const slotHeight = calendar.el.querySelector('.fc-timegrid-slot')?.offsetHeight || 30;
	const slotDuration = calendar.getOption('slotDuration') || '00:30:00';
	const slotMinTime = calendar.getOption('slotMinTime') || '00:00:00';

	const centerPixelOffset = scrollTop + (clientHeight / 2);
	const msPerPixel = timeToMs(slotDuration) / slotHeight;
	const totalMs = timeToMs(slotMinTime) + (centerPixelOffset * msPerPixel);

	return msToTime(totalMs);
}

/**
 * Calculates the new scrollTime needed to keep a target time centered in the view after a zoom change.
 * @param {Calendar} calendar - The FullCalendar instance.
 * @param {string} targetTime - The time to keep centered, in "HH:MM:SS" format.
 * @returns {string} The new scrollTime value.
 */
function getScrollTimeForCentering(calendar, targetTime) {
	if (!calendar || !targetTime) return targetTime;

	const slotDurationMs = timeToMs(calendar.getOption('slotDuration') || '00:30:00');

	// Try multiple selectors to find the scroller element
	let scrollerEl = calendar.el.querySelector('.fc-scroller.fc-scroller-liquid-absolute');
	if (!scrollerEl) {
		scrollerEl = calendar.el.querySelector('.fc-scroller');
	}
	if (!scrollerEl) {
		scrollerEl = calendar.el.querySelector('.fc-timegrid-body')?.parentElement;
	}
	if (!scrollerEl) return targetTime;

	const slotHeight = calendar.el.querySelector('.fc-timegrid-slot')?.offsetHeight || 30;
	const visibleHeight = scrollerEl.clientHeight;
	const visibleDurationMs = (visibleHeight / slotHeight) * slotDurationMs;
	const halfVisibleDurationMs = visibleDurationMs / 2;

	const targetTimeMs = timeToMs(targetTime);
	let newScrollTimeMs = targetTimeMs - halfVisibleDurationMs;
	if (newScrollTimeMs < 0) newScrollTimeMs = 0;

	return msToTime(newScrollTimeMs);
}

/**
 * Updates the calendar's zoom level.
 * @param {Calendar} calendar - The FullCalendar instance.
 * @param {'in' | 'out'} direction - The direction to zoom.
 */
export function updateZoom(calendar, direction) {
	const currentSlotDuration = calendar.getOption('slotDuration');
	const currentIndex = ZOOM_LEVELS.indexOf(currentSlotDuration);
	
	if (currentIndex === -1) return;
	if (direction === 'in' && currentIndex >= ZOOM_LEVELS.length - 1) return;
	if (direction === 'out' && currentIndex <= 0) return;

	const centerTime = getCenterTime(calendar);
	const newIndex = (direction === 'in') ? currentIndex + 1 : currentIndex - 1;
	const newDuration = ZOOM_LEVELS[newIndex];

	calendar.setOption('slotDuration', newDuration);
	calendar.setOption('slotLabelInterval', newDuration);

	// We use a short timeout to allow the DOM to update before we calculate the new scroll position.
	setTimeout(() => {
		const scrollToTime = getScrollTimeForCentering(calendar, centerTime);
		calendar.scrollToTime(scrollToTime);
	}, 50);
}

