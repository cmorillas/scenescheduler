// File: components/calendar/menu-actions.mjs

import { exportSchedule, importSchedule } from './schedule-adapter.mjs';
import { sendMessage, getScheduleFromUser } from '../../services/websocket.mjs';
import { addLogMessage } from '../../shared/ui-updater.mjs';

// =============================
// EXPORTED FUNCTION
// =============================
/**
 * Single entry point used by the calendar menu.
 */
export function handleMenuAction(calendar, action) {
  switch (action) {
    case 'new':
      if (confirm('Are you sure? All current events will be removed.')) {
        calendar.removeAllEvents();
      }
      break;

    case 'load-local':
      loadScheduleFromFile(calendar);
      break;

    case 'save-local':
      saveScheduleToFile(calendar);
      break;

    case 'get-server':
      // Request schedule from server (user action)
      // Will prompt if there are unsaved changes
      getScheduleFromUser();
      break;

    case 'commit-server':
      // 1. Get the current schedule from the calendar in our defined JSON format.
      const schedule = exportSchedule(calendar);

      // 2. Send the entire schedule object to the server.
      sendMessage('commitSchedule', schedule);

      // Log the commit action
      const eventCount = schedule?.schedule?.length || 0;
      addLogMessage(`Committed ${eventCount} events to server`, 'info');
      break;
  }
}

// =============================
// PRIVATE FUNCTIONS
// =============================
function saveScheduleToFile(calendar) {
  const schedule = exportSchedule(calendar);

  const blob = new Blob([JSON.stringify(schedule, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);

  const a = document.createElement('a');
  a.href = url;
  a.download = `schedule-${new Date().toISOString().slice(0, 10)}.json`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function loadScheduleFromFile(calendar) {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = '.json,application/json';

  input.onchange = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (ev) => {
      try {
        const json = JSON.parse(ev.target.result);
        if (!json?.schedule || !Array.isArray(json.schedule)) {
          throw new Error('Invalid schedule format: missing "schedule" array.');
        }
        if (confirm('Are you sure? All current events will be removed.')) {
          importSchedule(calendar, json);
        }
      } catch (err) {
        alert('Error parsing the JSON file: ' + err.message);
        console.error('JSON Parse Error:', err);
      }
    };
    reader.readAsText(file);
  };

  input.click();
}

