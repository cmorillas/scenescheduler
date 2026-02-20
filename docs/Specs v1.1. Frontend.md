# **Scene Scheduler for OBS \- Frontend Implementation Guide**

**Status:** Prescriptive  
 **Version:** 1.1  
 **Date:** 2025-10-11

---

## **IMPORTANT NOTE**

This is a prescriptive specifications document. It defines how the frontend is built. It serves as:

* Complete architectural guide for implementation  
* Reference for module design and responsibilities  
* Design contract for the visual schedule editor and monitor

---

## **1\. Project Overview**

### **1.1. System Context**

This frontend is a visual editor and real-time monitor for an OBS automation backend.

**The Backend Context:**

* A Go backend server continuously controls OBS Studio via obs-websocket  
* The backend reads `schedule.json` (a JSON file) as its single source of truth  
* Every second, the backend evaluates which program should be active and switches OBS scenes accordingly  
* The backend includes an embedded web server that serves this frontend application  
* The backend provides WebSocket APIs for communication and monitoring

**Frontend's Dual Role:**

1. **EDITOR** \- Schedule Management (90% of usage)

   * Visual calendar-based interface for editing schedule.json  
   * Create, modify, delete scheduled programs  
   * Load/save schedules from/to server or local files  
   * Working copy that can diverge from server (sandbox mode)  
   * Explicit commit workflow to apply changes  
2. **MONITOR** \- Real-Time Observation (10% of usage)

   * Live Preview: Display OBS output via WebRTC stream (what's currently broadcasting)  
   * Activity Log: Real-time stream of backend operations (connections, switches, errors)  
   * Read-only Calendar: Shows the actual schedule running on the server  
   * Connection Status: Visual indicator of backend WebSocket connection  
   * Current program highlight: See which event is broadcasting NOW

**Key Architectural Principle:**

The frontend operates in **two distinct modes**, each with its own calendar instance:

* **Monitor View:** Read-only observation of reality (server schedule \+ live video \+ activity log)  
* **Editor View:** Editable workspace for experimenting with changes (working copy \+ commit workflow)

This separation ensures users cannot accidentally modify the live broadcast schedule while monitoring, and provides a safe sandbox for testing changes.

---

### **1.2. Core Characteristics**

* **Auxiliary tool:** Backend functions completely independently; frontend is convenience only  
* **Dual state:** Maintains separate server state (Monitor) and working copy (Editor)  
* **Explicit synchronization:** User manually commits from Editor to server  
* **Isolated editing:** Changes in Editor don't affect Monitor or live broadcast until committed  
* **Maximum simplicity:** Native browser features over custom implementations  
* **No authentication:** Inherited from backend WebServer configuration

---

### **1.3. Non-Goals**

This frontend does NOT:

* ❌ Control OBS directly (backend does this)  
* ❌ Command backend to switch programs (backend follows schedule autonomously)  
* ❌ Require authentication (inherited from backend)  
* ❌ Support offline editing  
* ❌ Provide real-time collaboration  
* ❌ Store schedules locally (server is source of truth)

The frontend is purely:

* A visual editor for schedule.json  
* A window into backend operations  
* Not a remote control for OBS

---

## **2\. Architecture Overview**

### **2.1. System Context Diagram**

┌──────────────────────────────────────────────┐  
│  Backend (Autonomous)                        │  
│  ├─ schedule.json (source of truth)          │  
│  ├─ Scheduler (reads schedule every second)  │  
│  ├─ OBSClient (controls OBS)                 │  
│  └─ WebServer (serves frontend \+ WebSocket)  │  
└──────────────────────────────────────────────┘  
                    ↕ HTTP/WebSocket  
┌──────────────────────────────────────────────┐  
│  Frontend (Editor \+ Monitor)                 │  
│  ├─ Monitor View (read-only observation)     │  
│  │  ├─ Monitor Calendar (server schedule)    │  
│  │  ├─ Live Preview (video feed)             │  
│  │  └─ Activity Log (backend events)         │  
│  │                                            │  
│  └─ Editor View (editable workspace)         │  
│     ├─ Editor Calendar (working copy)        │  
│     ├─ Status Bar (sync state)               │  
│     └─ Action Buttons (commit/revert)        │  
└──────────────────────────────────────────────┘

---

### **2.2. Dual View Architecture**

**Monitor View (Read-only):**

* Purpose: Observe what is actually happening  
* Calendar: Always synced with server's schedule.json  
* Interactions: None (read-only)  
* Auto-updates: Yes (when server schedule changes)  
* Use case: "What is broadcasting right now?"

**Editor View (Editable):**

* Purpose: Experiment with schedule changes  
* Calendar: Working copy that can diverge from server  
* Interactions: Full CRUD operations  
* Auto-updates: No (isolated from server changes)  
* Use case: "Let me try changing the schedule"

**State Model:**

Backend: schedule.json (single source of truth)  
   ↓  
Frontend maintains TWO copies:  
   ├─ Server State → displayed in Monitor View  
   │                 (always current, auto-syncs)  
   └─ Working Copy → displayed in Editor View  
                     (can diverge, manual sync)

**Synchronization:**

* Monitor View: Auto-syncs when server sends updates  
* Editor View: User explicitly chooses when to:  
  * Load from server (discard local changes)  
  * Commit to server (apply local changes)  
  * Revert (discard changes, reload from server)  
  * Load from local file (replace working copy with file)  
  * Save to local file (export working copy to JSON)

---

### **2.3. Communication Architecture**

**Pattern:** DOM CustomEvents for loose coupling

**WebSocket Service:**

* Manages connection lifecycle  
* Dispatches all messages as CustomEvent on document  
* No knowledge of consumers

**Components:**

* Subscribe to events via addEventListener  
* Fully decoupled from each other  
* Can be added/removed independently

**Event Flow:**

WebSocket receives message  
  → Dispatches 'ws:message' CustomEvent  
    → Monitor View reacts (updates calendar)  
    → Editor View reacts (updates status if relevant)  
    → Info Window logs message  
    → Other components can listen (future extensibility)

---

## **3\. Module Structure**

### **3.1. Directory Layout**

frontend/  
├── index.html  
├── main.mjs                         \# Entry point  
├── main.css                         \# Global layout \+ CSS variables  
│  
├── services/
│   └── websocket.mjs                \# WebSocket lifecycle
│
├── shared/  
│   ├── utils.mjs                    \# Pure utilities
│   ├── app-state.mjs                \# Global state object (flags only)
│   └── ui-updater.mjs               \# Manual UI update function
│  
└── components/  
    ├── view-switcher/  
    │   ├── view-switcher.mjs        \# Tab navigation  
    │   └── view-switcher.css  
    │  
    ├── monitor-view/  
    │   ├── monitor-view.mjs         \# Monitor tab coordinator  
    │   └── monitor-view.css  
    │  
    ├── editor-view/  
    │   ├── editor-view.mjs          \# Editor tab coordinator  
    │   └── editor-view.css  
    │  
    ├── info-window/  
    │   ├── info-window.mjs          \# Activity log  
    │   └── info-window.css  
    │  
    ├── live-preview/  
    │   ├── live-preview.mjs         \# WHEP client  
    │   └── live-preview.css  
    │  
    └── calendar/  
        ├── calendar-shared.mjs      \# Common FullCalendar config  
        ├── calendar-monitor.mjs     \# Read-only calendar instance  
        ├── calendar-editor.mjs      \# Editable calendar instance  
        ├── calendar.css  
        ├── schedule-adapter.mjs     \# Format conversion  
        ├── calendar-events.mjs      \# CRUD operations  
        ├── grid-actions.mjs         \# Drag/resize handlers  
        ├── menu-actions.mjs         \# Load/save/commit actions  
        ├── menu.mjs                 \# Menu creation  
        ├── status-bar.mjs           \# Sync state indicator  
        ├── helpers.mjs              \# Pure utility functions  
        ├── zoom.mjs                 \# View zoom control  
        │  
        └── modal/  
            ├── modal.mjs            \# Coordinator  
            ├── modal.css  
            ├── form.mjs             \# Form population/extraction  
            ├── ui.mjs               \# UI interactions (tabs, drag, colors)  
            └── validation.mjs       \# Validation logic

---

### **3.2. File Naming Convention**

**Rule:** Short names for standard concepts, descriptive for domain-specific

**Standard concepts (short):**

* `state.mjs` \- Application state  
* `helpers.mjs` \- Utility functions  
* `validation.mjs` \- Input validation

**Domain-specific (descriptive):**

* `schedule-adapter.mjs` \- Schedule format conversion  
* `view-switcher.mjs` \- Tab navigation  
* `status-bar.mjs` \- Sync status indicator

**Rationale:** Directory provides context. `calendar/helpers.mjs` is self-explanatory. Domain concepts need clarity for discoverability.

---

## **4\. Core Modules**

### **4.1. Application State & UI Updates**

**Responsibility:** Centralized management for simple, shared state flags and a manual coordinator for UI updates. This pattern prioritizes simplicity and clarity over reactivity.

---

#### **`shared/app-state.mjs` (El Estado)**

**Responsibility:** Hold shared, simple state variables. It's a plain JavaScript object with no logic.

**Exported Object:**

```javascript
export const AppState = {
    // Connection State
    isConnected: false,

    // Editor Sync State
    isDirty: false,
    changeCount: 0,

    // Monitor State
    currentProgramId: null,

    // UI State
    activeView: 'monitor'
};
```
**Note:** Complex data like schedule arrays are managed by their respective calendar components, not in this global object.

---

#### **`shared/ui-updater.mjs` (El Actualizador)**

**Responsibility:** Provide a single function that reads the current state from `AppState` and updates all relevant parts of the DOM.

**Exports:**
* `updateUI()` - Reads the entire `AppState` and synchronizes the UI.

**Pattern: Manual Update Flow**

The flow for state changes is always the same, explicit 3-step process:
1.  **Modify State:** Another module directly changes a property (e.g., `AppState.isDirty = true;`).
2.  **Trigger Update:** The same module then immediately calls `updateUI()`.
3.  **UI Reacts:** The `updateUI()` function reads the new state and updates all necessary DOM elements (buttons, status bars, indicators, etc.).

---

### **4.2. WebSocket Service**

**File:** `services/websocket.mjs`  
 **Responsibility:** WebSocket connection lifecycle and message dispatching

**Exports:**

* `connect()` \- Establish connection  
* `sendMessage(action, payload)` \- Send to server

**Connection Management:**

* URL: `/ws` (relative to current origin, auto-detects http/https → ws/wss)  
* Auto-reconnect: Exponential backoff starting at 3 seconds  
* Reconnects indefinitely until connection restored

**Message Dispatching:**

All received messages dispatched as:

document.dispatchEvent(new CustomEvent('ws:message', {  
    detail: { action: string, payload: object }  
}));

Connection state changes dispatched as:

document.dispatchEvent(new CustomEvent('ws:statusChange', {  
    detail: { text: string, color: string }  
}));

**Protocol:**

Outgoing:

{ "action": "getSchedule" | "commitSchedule", "payload": {...} }

Incoming:

{ "action": "currentSchedule" | "scheduleChanged" | "commitSuccess" | "commitError", "payload": {...} } obs.program.changed

---

### **4.3. View Switcher**

**File:** `components/view-switcher/view-switcher.mjs`  
 **Responsibility:** Handle tab switching between Monitor and Editor views

**Exports:**

* `initViewSwitcher()` \- Initialize tab navigation

**Behavior:**

* Switches active view  
* Shows/hides appropriate DOM sections  
* Updates tab visual state  
* Persists last active view in localStorage  
* Shows dirty indicator on Editor tab when isDirty

**Events Dispatched:**

document.dispatchEvent(new CustomEvent('view:changed', {  
    detail: { view: 'monitor' | 'editor' }  
}));

---

### **4.4. Monitor View**

**File:** `components/monitor-view/monitor-view.mjs`  
 **Responsibility:** Coordinate all Monitor tab components

**Exports:**

* `initMonitorView()` \- Initialize monitor components

**Sub-components:**

1. **Monitor Calendar** \- Read-only calendar showing server schedule  
2. **Live Preview** \- Video feed of OBS output  
3. **Activity Log** \- Real-time backend events

**Behavior:**

* Initializes read-only calendar  
* Subscribes to `appState.serverSchedule` changes  
* Auto-refreshes calendar when server schedule updates  
* Highlights current program in calendar  
* Does NOT allow any editing

**Current Program Highlight:**

* Event currently broadcasting has special styling  
* Visual indicator: "▶ LIVE" badge  
* Pulsing glow animation  
* Updates based on `obs.program.changed` messages

---

### **4.5. Editor View**

**File:** `components/editor-view/editor-view.mjs`  
 **Responsibility:** Coordinate all Editor tab components

**Exports:**

* `initEditorView()` \- Initialize editor components

**Sub-components:**

1. **Editor Calendar** \- Editable calendar with working copy  
2. **Status Bar** \- Shows sync state (clean/dirty/syncing/error)  
3. **Action Buttons** \- Revert, Sync from Server, Commit  
4. **Calendar Menu** \- New, Load from File, Save to File, Get from Server, Commit to Server

**Behavior:**

* Initializes editable calendar  
* The central `updateUI()` function is responsible for updating the status bar and action buttons based on `AppState.isDirty`.
* Enables/disables action buttons based on isDirty  
* Warns before closing tab if isDirty  
* Provides menu for file operations and server sync

**Action Buttons:**

* **Revert Changes:** Discard all changes, reload from server (requires confirmation)  
* **Sync from Server:** Pull latest server schedule (overwrites local changes with confirmation)  
* **Commit to Server:** Send working copy to server (applies changes to live schedule)

---

### **4.6. Calendar Components**

#### **4.6.1. Shared Configuration**

**File:** `components/calendar/calendar-shared.mjs`  
 **Responsibility:** Common FullCalendar configuration

**Exports:**

* `getSharedConfig()` \- Returns base FullCalendar config object

**Configuration Includes:**

* View settings (timeGridWeek, slotDuration, etc.)  
* Time format (24-hour, HH:MM:SS)  
* Visual styling  
* Event rendering helpers  
* Zoom buttons (+/-)  
* Menu button (...) \- only in Editor calendar  
* Everything common to both calendar instances

---

#### **4.6.2. Monitor Calendar**

**File:** `components/calendar/calendar-monitor.mjs`  
 **Responsibility:** Read-only calendar for Monitor view

**Exports:**

* `initMonitorCalendar(container)` \- Returns FullCalendar instance

**Configuration:**

* Inherits from shared config  
* `editable: false` \- Cannot drag/resize  
* `selectable: false` \- Cannot select time ranges  
* `eventClick: null` \- No modal on click  
* `dateClick: null` \- No event creation  
* No menu button  
* Custom rendering for current program highlight

**Special Features:**

* Auto-refreshes when `appState.serverSchedule` changes  
* Highlights currently broadcasting event  
* Shows "▶ LIVE" badge on current program  
* Completely passive (no user interactions)

---

#### **4.6.3. Editor Calendar**

**File:** `components/calendar/calendar-editor.mjs`  
 **Responsibility:** Editable calendar for Editor view

**Exports:**

* `initEditorCalendar(container)` \- Returns FullCalendar instance

**Configuration:**

* Inherits from shared config  
* `editable: true` \- Can drag/resize  
* `selectable: true` \- Can select time ranges  
* `eventOverlap: false` \- Prevent overlaps  
* Includes menu button in toolbar  
* All event callbacks update `appState.editorSchedule`

**Event Callbacks:**

* `eventChange` → Mark dirty  
* `eventAdd` → Mark dirty  
* `eventRemove` → Mark dirty  
* `eventClick` → Open modal  
* `dateClick` → Create new event  
* `select` → Create event in range  
* `eventDrop` → Update event  
* `eventResize` → Update event

**Special Features:**

* Full CRUD operations  
* isDirty tracking on every change  
* Menu with file and server operations  
* beforeunload warning if unsaved changes

---

### **4.7. Schedule Adapter**

**File:** `components/calendar/schedule-adapter.mjs`  
 **Responsibility:** Bidirectional conversion between FullCalendar format and Schedule 1.0 JSON

**Exports:**

* `exportSchedule(calendar, options)` → Schedule 1.0 object  
* `importSchedule(calendar, scheduleJson)` → void

**Export Process:**

1. Iterate all events in FullCalendar  
2. Convert each to Schedule 1.0 format  
3. Deduplicate recurring event instances (one master event per series)  
4. Build complete schedule object with version and metadata

**Import Process:**

1. Clear all existing events  
2. Parse Schedule 1.0 JSON  
3. Convert each item to FullCalendar format  
4. Handle recurring events (expand to FullCalendar recurrence rules)  
5. Render in calendar

**Recurrence Transformation:**

* FullCalendar uses `daysOfWeek` as numbers (0=Sun, 1=Mon, ...)  
* Schedule 1.0 uses `daysOfWeek` as strings ("MON", "TUE", ...)  
* Adapter converts between formats

---

### **4.8. Status Bar**

**File:** `components/calendar/status-bar.mjs`  
 **Responsibility:** Visual indicator of synchronization state

**Exports:**

* `updateStatusBar(state, customMessage)` \- Update status display

**States:**

* `'clean'` → Green circle, "Synced with server"  
* `'dirty'` → Orange circle, "X unsaved changes"  
* `'syncing'` → Blue circle, "Saving..."  
* `'error'` → Red circle, "Save failed"

**Implementation:**

* Updates CSS classes on status bar element  
* Updates text content  
* Updates circle color

**HTML Structure (required in index.html):**

\<div id="status-bar" class="status-bar"\>  
    \<div id="status-circle"\>\</div\>  
    \<span id="status-text"\>\</span\>  
\</div\>

---

### **4.9. Menu**

**File:** `components/calendar/menu.mjs`  
 **Responsibility:** Create and manage calendar menu dropdown

**Exports:**

* `createMenu(calendar)` \- Create menu DOM and attach handlers  
* `toggleMenu()` \- Show/hide menu

**Menu Structure:**

New Schedule  
Load from File  
Save to File  
──────────────  
Get from Server  
Commit to Server

**Behavior:**

* Appears when clicking "..." button in calendar toolbar  
* Positioned relative to button  
* Closes when clicking outside  
* Only visible in Editor View  
* Delegates actions to menu-actions.mjs

**Implementation:**

* Creates dropdown element dynamically  
* Attaches to document.body  
* Positioned absolutely using getBoundingClientRect  
* Z-index above calendar

---

### **4.10. Menu Actions**

**File:** `components/calendar/menu-actions.mjs`  
 **Responsibility:** Handle all menu operations

**Exports:**

* `handleMenuAction(calendar, action)` \- Process menu action

**Actions:**

1. **'new'** \- New Schedule

   * Confirmation: "Remove all current events?"  
   * Clears calendar  
   * Marks as dirty  
2. **'load-local'** \- Load from File

   * Opens file picker (accepts .json)  
   * Reads file  
   * Validates Schedule 1.0 format  
   * Confirmation: "Replace all current events?"  
   * Imports schedule via schedule-adapter  
   * Marks as dirty  
3. **'save-local'** \- Save to File

   * Exports schedule via schedule-adapter  
   * Creates Blob from JSON  
   * Triggers download with filename: `schedule-YYYY-MM-DD.json`  
   * Does NOT mark as clean (working copy still exists)  
4. **'get-server'** \- Get from Server

   * Sends WebSocket message: `getSchedule`  
   * Confirmation if isDirty: "Discard unsaved changes?"  
   * Waits for `currentSchedule` response  
   * Imports schedule  
   * Marks as clean  
5. **'commit-server'** \- Commit to Server

   * Exports current schedule  
   * Updates status bar to 'syncing'  
   * Sends WebSocket message: `commitSchedule`  
   * Waits for `commitSuccess` or `commitError`  
   * Updates status bar accordingly  
   * Marks as clean on success

**Delegates to:**

* `schedule-adapter.mjs` for import/export  
* `websocket.mjs` for server communication  
* `status-bar.mjs` for visual feedback  
* `appState` for state updates

**File Operations Pattern:**

// Load from File  
const input \= document.createElement('input');  
input.type \= 'file';  
input.accept \= '.json';  
input.onchange \= (e) \=\> {  
    const file \= e.target.files\[0\];  
    const reader \= new FileReader();  
    reader.onload \= (ev) \=\> {  
        const json \= JSON.parse(ev.target.result);  
        importSchedule(calendar, json);  
    };  
    reader.readAsText(file);  
};  
input.click();

// Save to File  
const schedule \= exportSchedule(calendar);  
const blob \= new Blob(\[JSON.stringify(schedule, null, 2)\], { type: 'application/json' });  
const url \= URL.createObjectURL(blob);  
const a \= document.createElement('a');  
a.href \= url;  
a.download \= \`schedule-${new Date().toISOString().slice(0, 10)}.json\`;  
a.click();  
URL.revokeObjectURL(url);

---

### **4.11. Modal Component**

**Directory:** `components/calendar/modal/`  
 **Responsibility:** Event editing interface

**Structure:** Divided by responsibility into 4 files

#### **4.11.1. Modal Coordinator**

**File:** `modal.mjs`  
 **Responsibility:** Public API and lifecycle management

**Exports:**

* `openTaskModal(calendar, options)` \- Open modal for create/edit

**Handles:**

* Modal open/close  
* Save/delete coordination  
* Delegates to submodules for specific tasks

---

#### **4.11.2. Form Operations**

**File:** `form.mjs`  
 **Responsibility:** Form data population and extraction

**Exports:**

* `populateForm(event)` \- Fill form fields from event object  
* `extractFormData()` \- Build event object from form fields  
* `resetForm()` \- Clear all fields  
* `calculateDuration()` \- Update duration display  
* `toggleRecurring()` \- Show/hide recurrence fields  
* `updateUriHint(kind)` \- Update URI hint based on input kind  
* `autoGenerateSourceName()` \- Generate source name from title

---

#### **4.11.3. UI Interactions**

**File:** `ui.mjs`  
 **Responsibility:** UI event handlers

**Exports:**

* `initUIHandlers()` \- Setup all UI event listeners

**Handles:**

* Tab switching between modal tabs  
* Drag and drop modal repositioning  
* Color picker interactions  
* Clear color buttons

---

#### **4.11.4. Validation**

**File:** `validation.mjs`  
 **Responsibility:** Form validation logic

**Exports:**

* `validateForm(formData)` \- Validate all fields, return errors  
* `parseJsonField(raw, label)` \- Parse and validate JSON fields

**Validation Rules:**

* Required fields: title, source name, source URI  
* JSON fields: inputSettings, transform must be valid JSON  
* Input kind: must be in allowed list  
* Times: start must be before end

---

### **4.12. Events Module**

**File:** `components/calendar/calendar-events.mjs`  
 **Responsibility:** Centralized CRUD operations for calendar events

**Purpose:** Single source of truth for all event mutations

**Exports:**

* `updateEvent(calendar, eventData, existingEvent)` \- Create or update  
* `deleteEvent(eventToDelete)` \- Remove with confirmation  
* `prepareRecurringUpdate(event, info)` \- Prepare recurring event update data

**Pattern:** All calendar modifications flow through this module for consistency

---

### **4.13. Grid Actions**

**File:** `components/calendar/grid-actions.mjs`  
 **Responsibility:** Handle drag and resize operations on calendar events

**Exports:**

* `handleGridAction(calendar, info)` \- Process drag/resize callback

**Behavior:**

* For simple events: Update directly  
* For recurring events: Prompt user "Update all instances?" before applying  
* Delegates to `calendar-events.mjs` for actual mutations

---

### **4.14. Helpers**

**File:** `components/calendar/helpers.mjs`  
 **Responsibility:** Pure utility functions for date/time manipulation

**Exports:**

* `pad2(n)` \- Zero-pad numbers  
* `genId(prefix)` \- Generate unique IDs  
* `ensureHHMMSS(timeStr)` \- Normalize time format  
* `timeToMs(timeStr)` \- Convert time string to milliseconds  
* `msToTime(totalMs)` \- Convert milliseconds to time string  
* `toLocalDateString(value)` \- Format date as YYYY-MM-DD  
* `toLocalDateTimeString(value)` \- Format for datetime-local input  
* `formatTodayYYYYMMDD()` \- Get today's date string  
* `splitLocal(value)` \- Split datetime into {date, time}  
* `addSecondsLocalISO(localIso, seconds)` \- Add seconds to date  
* `daysOfWeekNumsToNames(nums)` \- Convert \[1,3,5\] → \["MON","WED","FRI"\]  
* `weekdaysNamesToNums(names)` \- Convert \["MON","WED"\] → \[1,3\]  
* `dateToHHMMSS(date)` \- Extract time from date

**Characteristics:** No side effects, no state, testable in isolation

---

### **4.15. Zoom**

**File:** `components/calendar/zoom.mjs`  
 **Responsibility:** Change calendar time slot granularity

**Exports:**

* `updateZoom(calendar, direction)` \- 'in' or 'out'

**Zoom Levels:** `['01:00:00', '00:30:00', '00:15:00', '00:05:00', '00:01:00', '00:00:30']`

**Behavior:**

* Maintains centered time when zooming  
* Calculates scroll position to keep view stable  
* Limits reached at finest/coarsest levels

---

### **4.16. Info Window**

**File:** `components/info-window/info-window.mjs`  
 **Responsibility:** Display real-time backend activity log

**Exports:**

* `initInfoWindow(container)` \- Initialize log display

**Subscriptions:**

* `ws:statusChange` → Update connection indicator  
* `ws:message` → Log all messages from backend

**Content Displayed:**

* Backend connection events (OBS connected/disconnected)  
* Schedule file changes detected  
* Program switches executed  
* Media source state changes  
* Virtual camera state  
* Validation errors  
* All backend operations in chronological order

**UI:**

* Scrollable log area with timestamps  
* Connection status indicator (colored circle \+ text)  
* Auto-scroll to latest messages

**Purpose:** Provides visibility into backend operations without needing server logs or OBS

---

### **4.17. Live Preview**

**File:** `components/live-preview/live-preview.mjs`  
 **Responsibility:** Real-time monitoring of OBS output via WHEP protocol

**Exports:**

* `initLivePreview(container)` \- Initialize video preview

**Implementation:**

* WHEPClient class (WebRTC connection to backend's WHEP endpoint)  
* Auto-reconnect on signal loss  
* Play/pause controls  
* Overlay for "Signal Lost" state  
* Mock tracks while no video available

**Purpose:**

* Display exactly what OBS is broadcasting in real-time  
* Verify correct program is playing  
* Check visual quality without opening OBS  
* Monitor scene transitions  
* Confirm schedule is working as expected

**Data Flow:**

OBS Virtual Camera → Backend MediaSource → WHEP Server → WebRTC → Browser Video Element

**Independent Operation:**

* No interaction with schedule or calendar state  
* Purely monitoring/observation  
* Does not control what's playing

---

## **5\. Event Preview System**

### **5.1. Overview**

The Event Preview System allows users to preview the actual content of scheduled programs using HLS (HTTP Live Streaming) generated on-demand by the backend via FFmpeg.

**Two Preview Contexts:**

1. **Editor Modal Preview** (Tab 5 of edit modal)

   * Full-featured preview when editing an event  
   * Available in Editor View only  
   * Part of the complete event editing workflow  
2. **Monitor Popup Preview**

   * Quick preview popup when clicking events in Monitor View calendar  
   * Read-only information display  
   * Minimal interface for fast content verification

**Backend Architecture:**

* Backend uses FFmpeg to generate HLS streams on-demand  
* Supports **any** `inputKind` that OBS can use  
* Stream lifecycle managed via WebSocket protocol  
* Automatic cleanup on timeout or explicit stop

---

### **5.2. Supported Input Types**

The preview system works for **all** `inputKind` values. The backend determines capability:

**Always Supported:**

* `ffmpeg_source` \- Local video/audio files  
* `vlc_source` \- Playlists  
* `image_source` \- Static images (single-frame HLS)  
* Remote URLs (HTTP/HTTPS streams)

**Backend-Dependent:**

* `browser_source` \- Requires headless browser rendering  
* `ndi_source` \- Requires NDI support in backend  
* `window_capture` / `display_capture` \- Requires GUI access  
* `color_source` / `text_source` \- Backend generates synthetic frames

**Strategy:**

* Frontend always allows preview attempt for any inputKind  
* Backend responds with `previewReady` if stream can be generated  
* Backend responds with `previewError` if not possible (with specific reason)

---

### **5.3. WebSocket Protocol**

**Request Preview:**

{  
    "action": "requestPreview",  
    "payload": {  
        "eventId": "evt-123",  
        "sourceUri": "/path/to/video.mp4",  
        "inputKind": "ffmpeg\_source",  
        "inputSettings": {}  
    }  
}

**Preview Ready (Success):**

{  
    "action": "previewReady",  
    "payload": {  
        "eventId": "evt-123",  
        "streamId": "stream-abc123",  
        "hlsUrl": "/stream/stream-abc123/playlist.m3u8"  
    }  
}

**Preview Error (Failure):**

{  
    "action": "previewError",  
    "payload": {  
        "eventId": "evt-123",  
        "error": "Cannot preview window\_capture on headless server",  
        "reason": "unsupported\_on\_platform"  
    }  
}

**Error Reasons:**

* `"file_not_found"` \- Source file doesn't exist  
* `"invalid_uri"` \- Malformed URI  
* `"unsupported_on_platform"` \- inputKind not supported on this backend  
* `"stream_generation_failed"` \- FFmpeg error

**Stop Preview:**

{  
    "action": "stopPreview",  
    "payload": {  
        "streamId": "stream-abc123"  
    }  
}

---

### **5.4. Editor Modal Preview (Tab 5\)**

**Location:** Edit Event Modal → Tab 5: Preview

**Implementation:**

* File: `components/calendar/modal/preview.mjs` (new file)  
* Integrated into modal component structure  
* Uses hls.js library for HLS playback

**UI Structure:**

┌─────────────────────────────────────────────┐  
│  Tab 5: Preview                             │  
├─────────────────────────────────────────────┤  
│                                             │  
│  ┌───────────────────────────────────────┐ │  
│  │                                       │ │  
│  │         \[▶ Preview Source\]            │ │  
│  │                                       │ │  
│  │  Click to load preview                │ │  
│  │                                       │ │  
│  └───────────────────────────────────────┘ │  
│                                             │  
│  Source: /path/to/video.mp4                 │  
│  Input Kind: ffmpeg\_source                  │  
│                                             │  
└─────────────────────────────────────────────┘

**After clicking Preview:**

┌─────────────────────────────────────────────┐  
│  Tab 5: Preview                             │  
├─────────────────────────────────────────────┤  
│                                             │  
│  ┌───────────────────────────────────────┐ │  
│  │  ╔═══════════════════════════════╗   │ │  
│  │  ║   \[Video Playing\]              ║   │ │  
│  │  ║   ━━━━━━━●───────── 00:45     ║   │ │  
│  │  ║   \[⏸\] \[🔊\]                    ║   │ │  
│  │  ╚═══════════════════════════════╝   │ │  
│  └───────────────────────────────────────┘ │  
│                                             │  
│  \[⏹ Stop Preview\]                           │  
│                                             │  
└─────────────────────────────────────────────┘

**Behavior:**

* Preview button enabled for all inputKinds  
* Click "Preview Source" → sends `requestPreview`  
* Shows loading spinner while backend generates stream  
* On success: Loads HLS stream with hls.js  
* On error: Displays error message from backend  
* Video player has standard HTML5 controls (play/pause/seek/volume)  
* "Stop Preview" button sends `stopPreview` to backend  
* Stream automatically stopped when modal closes

**Error Display:**

┌─────────────────────────────────────────────┐  
│  Tab 5: Preview                             │  
├─────────────────────────────────────────────┤  
│                                             │  
│  ┌───────────────────────────────────────┐ │  
│  │  ⚠️ Preview Error                     │ │  
│  │                                       │ │  
│  │  Cannot preview window\_capture        │ │  
│  │  on headless server                   │ │  
│  │                                       │ │  
│  │  \[Retry\]                              │ │  
│  └───────────────────────────────────────┘ │  
│                                             │  
└─────────────────────────────────────────────┘

---

### **5.5. Monitor Popup Preview**

**Location:** Monitor View → Click any event in calendar → Popup appears

**Implementation:**

* Directory: `components/monitor-preview/`  
* Files: `monitor-preview.mjs`, `monitor-preview.css`  
* Exports: `openMonitorPreview(event)`, `closeMonitorPreview()`

**Trigger:**

// File: components/calendar/calendar-monitor.mjs  
eventClick: (info) \=\> {  
    openMonitorPreview(info.event);  
}

**UI Structure:**

\<\!-- Monitor Event Preview Popup \--\>  
\<div id="monitor-preview-popup" class="monitor-popup"\>  
    \<div class="popup-header"\>  
        \<h3 id="popup-title"\>Bloomberg Live\</h3\>  
        \<button class="popup-close"\>✕\</button\>  
    \</div\>  
      
    \<div class="popup-content"\>  
        \<\!-- Video Preview Area \--\>  
        \<div class="preview-video-container"\>  
            \<video id="popup-video" controls\>\</video\>  
            \<button id="popup-play-preview" class="btn-preview"\>  
                ▶ Preview Source  
            \</button\>  
            \<div class="preview-status"\>Click to load preview\</div\>  
        \</div\>  
          
        \<\!-- Event Info (Read-only, Minimal) \--\>  
        \<div class="event-info"\>  
            \<div class="info-row"\>  
                \<span class="label"\>Source:\</span\>  
                \<span class="value"\>/media/bloomberg.mp4\</span\>  
            \</div\>  
            \<div class="info-row"\>  
                \<span class="label"\>Input Kind:\</span\>  
                \<span class="value"\>ffmpeg\_source\</span\>  
            \</div\>  
            \<div class="info-row"\>  
                \<span class="label"\>Schedule:\</span\>  
                \<span class="value"\>14:00 \- 15:00 (Mon, Wed, Fri)\</span\>  
            \</div\>  
        \</div\>  
    \</div\>  
      
    \<div class="popup-footer"\>  
        \<button class="btn-secondary" id="goto-editor"\>  
            Edit in Editor View  
        \</button\>  
        \<button class="btn-secondary" id="popup-close-btn"\>  
            Close  
        \</button\>  
    \</div\>  
\</div\>

**Visual Appearance:**

┌─────────────────────────────────────────────┐  
│  Bloomberg Live                        \[✕\]  │  
├─────────────────────────────────────────────┤  
│  ┌───────────────────────────────────────┐ │  
│  │  ╔═══════════════════════════════╗   │ │  
│  │  ║   \[Video Preview\]              ║   │ │  
│  │  ║   \[▶ Preview Source\]           ║   │ │  
│  │  ╚═══════════════════════════════╝   │ │  
│  └───────────────────────────────────────┘ │  
│                                             │  
│  Source: /media/bloomberg.mp4               │  
│  Input: ffmpeg\_source                       │  
│  Time: 14:00 \- 15:00 (Mon, Wed, Fri)        │  
│                                             │  
│  \[Edit in Editor View\]  \[Close\]             │  
└─────────────────────────────────────────────┘

**Information Displayed (Minimal):**

* Event title  
* Source URI  
* Input kind  
* Timing info (start, end, recurring days if applicable)  
* Preview video player

**NOT Displayed:**

* Description, tags, class names  
* Input settings JSON  
* Transform JSON  
* Behavior settings  
* Color settings

**Rationale:** If user needs detailed information, they should open the event in Editor View.

**Behavior:**

* Modal-style overlay (darkened background)  
* Centered on screen  
* Click outside or "✕" button to close  
* "Preview Source" button works same as Editor modal preview  
* Uses same WebSocket protocol  
* Stream cleanup on close

**"Edit in Editor View" Button:**

// When clicked:  
1\. Close popup  
2\. Send stopPreview (cleanup stream)  
3\. Switch to Editor View (change active tab)  
4\. Open edit modal for this event  
5\. User can now edit the event

---

### **5.6. HLS Player Implementation**

**Library:** hls.js v1.5.x (or latest stable)

**Loading:**

\<\!-- In index.html \--\>  
\<script src="vendor/hls.js/hls.min.js"\>\</script\>

**Usage Pattern:**

// File: components/monitor-preview/monitor-preview.mjs  
// (and similar in modal/preview.mjs)

let hls \= null;  
let currentStreamId \= null;

function playPreview(hlsUrl, streamId) {  
    const video \= document.getElementById('popup-video');  
      
    if (Hls.isSupported()) {  
        hls \= new Hls();  
        hls.loadSource(hlsUrl);  
        hls.attachMedia(video);  
        hls.on(Hls.Events.MANIFEST\_PARSED, () \=\> {  
            video.play();  
        });  
        hls.on(Hls.Events.ERROR, (event, data) \=\> {  
            console.error('HLS error:', data);  
            showPreviewError('Stream playback failed');  
        });  
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {  
        // Native HLS support (Safari)  
        video.src \= hlsUrl;  
        video.addEventListener('loadedmetadata', () \=\> {  
            video.play();  
        });  
    } else {  
        showPreviewError('HLS not supported in this browser');  
    }  
      
    currentStreamId \= streamId;  
}

function stopPreview() {  
    if (hls) {  
        hls.destroy();  
        hls \= null;  
    }  
      
    if (currentStreamId) {  
        websocket.sendMessage('stopPreview', { streamId: currentStreamId });  
        currentStreamId \= null;  
    }  
}

---

### **5.7. Stream Lifecycle**

**Creation:**

User clicks "Preview Source"  
  ↓  
Frontend: requestPreview with event details  
  ↓  
Backend: Validates source URI  
  ↓  
Backend: Starts FFmpeg process to generate HLS  
  ↓  
Backend: Responds with previewReady \+ HLS URL  
  ↓  
Frontend: Loads HLS stream with hls.js  
  ↓  
Video plays

**Cleanup (Multiple Triggers):**

1. **User clicks "Stop Preview"**

   * Frontend sends `stopPreview`  
   * Backend kills FFmpeg process  
   * Cleans up HLS segments  
2. **User closes modal/popup**

   * Frontend sends `stopPreview` automatically  
   * Backend cleanup  
3. **Backend timeout**

   * If no activity for 5 minutes  
   * Backend automatically kills stream  
   * Prevents orphaned processes  
4. **WebSocket disconnect**

   * Backend detects client disconnect  
   * Cleans up all streams for that client

---

### **5.8. Error Handling**

**File Not Found:**

User requests preview  
  ↓  
Backend checks: sourceUri doesn't exist  
  ↓  
Backend: previewError "File not found: /path/to/video.mp4"  
  ↓  
Frontend: Display error in preview area  
  ↓  
Show "Retry" button (in case file appears)

**Unsupported Input Kind:**

User requests preview of window\_capture  
  ↓  
Backend running headless (no GUI)  
  ↓  
Backend: previewError "unsupported\_on\_platform"  
  ↓  
Frontend: Show clear error message  
  ↓  
No retry option (won't work)

**Stream Generation Failed:**

Backend starts FFmpeg  
  ↓  
FFmpeg crashes or fails  
  ↓  
Backend: previewError "stream\_generation\_failed"  
  ↓  
Frontend: Show error \+ Retry button

**Network Issues:**

HLS stream loading  
  ↓  
Network interruption  
  ↓  
hls.js fires ERROR event  
  ↓  
Frontend: Display "Stream interrupted"  
  ↓  
Show Retry button

---

### **5.9. Module Structure Updates**

**New Files:**

components/  
├── monitor-preview/  
│   ├── monitor-preview.mjs      \# Popup for Monitor View  
│   └── monitor-preview.css  
│  
└── calendar/  
    └── modal/  
        ├── modal.mjs  
        ├── form.mjs  
        ├── ui.mjs  
        ├── validation.mjs  
        └── preview.mjs          \# ✨ NEW \- Tab 5 preview logic

**Updated Files:**

components/calendar/calendar-monitor.mjs  
  \- Add eventClick handler to open popup

components/calendar/modal/modal.mjs  
  \- Import preview.mjs  
  \- Initialize Tab 5 preview functionality

---

### **5.10. CSS Styling**

**Popup Overlay:**

.monitor-popup-overlay {  
    position: fixed;  
    top: 0;  
    left: 0;  
    right: 0;  
    bottom: 0;  
    background: rgba(0, 0, 0, 0.5);  
    z-index: 1000;  
    display: flex;  
    align-items: center;  
    justify-content: center;  
}

.monitor-popup {  
    background: var(--bg-panel);  
    border-radius: var(--radius);  
    box-shadow: var(--shadow);  
    width: 600px;  
    max-width: 90vw;  
    max-height: 80vh;  
    overflow: auto;  
}

**Video Container:**

.preview-video-container {  
    position: relative;  
    background: \#000;  
    border-radius: var(--radius);  
    overflow: hidden;  
    aspect-ratio: 16 / 9;  
}

.preview-video-container video {  
    width: 100%;  
    height: 100%;  
    object-fit: contain;  
}

.btn-preview {  
    position: absolute;  
    top: 50%;  
    left: 50%;  
    transform: translate(-50%, \-50%);  
    padding: 12px 24px;  
    font-size: 1rem;  
    background: var(--primary);  
    color: white;  
    border: none;  
    border-radius: var(--radius);  
    cursor: pointer;  
}

---

## **6\. Data Flows**

### **5.1. Application Startup**

1\. Browser loads index.html  
   ↓  
2\. main.mjs executes:  
   \- Import and initialize state.mjs  
   \- Import and call initViewSwitcher()  
   \- Import and call initMonitorView()  
   \- Import and call initEditorView()  
   \- Import and call websocket.connect()  
   ↓  
3\. websocket.mjs establishes connection  
   \- Dispatch 'ws:statusChange' → "Connecting"  
   \- On open → "Connected"  
   ↓  
4\. initMonitorView():  
   \- initLivePreview()  
   \- initInfoWindow()  
   \- initMonitorCalendar() ← Creates calendar instance \#1  
   ↓  
5\. initEditorView():  
   \- initEditorCalendar() ← Creates calendar instance \#2  
   \- createMenu(calendar) ← Attach menu to calendar  
   \- Initially hidden (display: none)  
   ↓  
6\. WebSocket automatically sends: getSchedule  
   ↓  
7\. Server responds: currentSchedule  
   ↓  
8\. appState.updateServerSchedule(payload)  
   ↓  
9\. Monitor calendar auto-refreshes  
   ↓  
10\. appState.syncEditorWithServer()  
    \- Editor calendar syncs with server

---

### **5.2. User Edits in Editor View**

1\. User switches to Editor tab  
   ↓  
2\. User drags event to new time  
   ↓  
3. FullCalendar fires 'eventDrop' callback
   ↓
4. handleGridAction() processes change and updates the Editor Calendar's internal data.
   ↓
5. The handler sets the global state: `AppState.isDirty = true;` and `AppState.changeCount++`.
   ↓
6. The handler calls the global updater: `updateUI()`.
   ↓
7. The `updateUI()` function reads `AppState` and updates all dependent elements at once:
    - Status bar shows "1 unsaved change".
    - Revert & Commit buttons are enabled.
    - Dirty indicator appears on the Editor tab.
    - The `beforeunload` browser event is activated.

---

### **5.3. User Saves Schedule to Local File**

1\. User clicks "..." button in calendar toolbar  
   ↓  
2\. Menu appears with options  
   ↓  
3\. User clicks "Save to File"  
   ↓  
4\. handleMenuAction(calendar, 'save-local') executes:  
   \- schedule \= exportSchedule(calendar)  
   \- Create Blob from JSON  
   \- Trigger download: schedule-YYYY-MM-DD.json  
   ↓  
5\. Browser downloads file  
   ↓  
6\. Status remains unchanged (still dirty if was dirty)  
   \- Saving to file is export, not sync with server

---

### **5.4. User Loads Schedule from Local File**

1\. User clicks "..." button in calendar toolbar  
   ↓  
2\. Menu appears  
   ↓  
3\. User clicks "Load from File"  
   ↓  
4\. handleMenuAction(calendar, 'load-local') executes:  
   \- Open file picker  
   ↓  
5\. User selects .json file  
   ↓  
6\. Read file contents  
   ↓  
7\. Parse and validate JSON  
   ↓  
8\. If valid Schedule 1.0 format:  
   \- Confirm: "Replace all current events?"  
   \- If YES:  
     \- importSchedule(calendar, json)  
     \- appState.updateEditorSchedule(json)  
     \- isDirty \= true (loaded file \!= server)  
     \- Status bar: "Unsaved changes"  
   ↓  
9\. If invalid format:  
   \- alert('Invalid schedule format')

---

### **5.5. User Commits Changes to Server**

1\. User clicks "Commit to Server" button  
   (or selects from menu)  
   ↓  
2\. updateStatusBar('syncing', 'Saving...')  
   ↓  
3\. schedule \= exportSchedule(editorCalendar)  
   ↓  
4\. websocket.sendMessage('commitSchedule', schedule)  
   ↓  
5\. Server validates → Writes schedule.json  
   ↓  
6\. SUCCESS PATH:  
   \- Server responds: 'commitSuccess'  
   \- appState.markEditorClean()  
   \- updateStatusBar('clean', 'Synced with server')  
   \- Buttons disable  
   ↓  
7\. ERROR PATH:  
   \- Server responds: 'commitError' with message  
   \- updateStatusBar('error', 'Save failed')  
   \- alert(error.message)  
   \- User can fix and retry

---

### **5.6. Server Hot-Reload (External Change)**

1\. schedule.json changes on disk  
   \- Manual edit OR  
   \- Another client commits  
   ↓  
2\. Backend FileWatcher detects change  
   ↓  
3\. Scheduler reloads schedule internally  
   ↓  
4\. WebServer broadcasts: 'scheduleChanged' to all clients  
   ↓  
5\. Frontend receives message  
   ↓  
6\. appState.updateServerSchedule(payload)  
   ↓  
7\. Monitor calendar auto-refreshes ✅  
   ↓  
8\. Editor calendar NOT affected ✅  
   \- Isolation maintained  
   \- User's work not interrupted  
   ↓  
9\. Passive notification shown:  
   "Schedule updated on server"  
   ↓  
10\. In Editor view, "Sync from Server" button available  
    \- User decides when to pull changes  
    \- No blocking confirmation dialog

**Key Design Decision:**

* Monitor view always shows reality (auto-updates)  
* Editor view is isolated sandbox (no auto-updates)  
* User has explicit control over when to sync

---

### **5.7. User Reverts Changes**

1\. User clicks "Revert Changes" in Editor view  
   ↓  
2\. Confirm: "Discard all unsaved changes?"  
   ↓  
3\. If YES:  
   \- appState.syncEditorWithServer()  
   \- Editor calendar reloads from server  
   \- appState.isDirty \= false  
   \- updateStatusBar('clean')  
   \- Buttons disable  
   ↓  
4\. If NO:  
   \- No action  
   \- Keep working

---

### **5.8. User Switches Views**

Monitor View Active:  
\- Monitor calendar visible (synced with server)  
\- Preview \+ Activity Log visible  
\- No edit controls  
\- Current program highlighted

User clicks "Editor" tab:  
  ↓  
  \- Monitor view hides (display: none)  
  \- Editor view shows  
  \- Editor calendar visible (may have unsaved changes)  
  \- Status bar shows sync state  
  \- Action buttons visible  
  \- Menu button visible in calendar toolbar

User clicks "Monitor" tab:  
  ↓  
  \- Editor view hides  
  \- Monitor view shows  
  \- Monitor calendar visible (always current)  
  \- Preview \+ Activity Log resume  
  \- Current program highlight updates

---

### **6.9. Event Preview Request (Both Views)**

User in Editor View editing event OR  
User in Monitor View clicks event  
  ↓  
User clicks "Preview Source" button  
  ↓  
Frontend gathers event data:  
  \- eventId  
  \- sourceUri (from form or event object)  
  \- inputKind  
  \- inputSettings (optional)  
  ↓  
websocket.sendMessage('requestPreview', payload)  
  ↓  
Show loading spinner: "Generating preview..."  
  ↓  
Backend receives requestPreview  
  ↓  
Backend validates sourceUri exists/accessible  
  ↓  
SUCCESS PATH:  
  Backend starts FFmpeg to generate HLS  
  ↓  
  Backend responds: previewReady with hlsUrl and streamId  
  ↓  
  Frontend receives previewReady  
  ↓  
  Hide loading spinner  
  ↓  
  Initialize hls.js with hlsUrl  
  ↓  
  Video element starts playing  
  ↓  
  Store streamId for cleanup  
    
ERROR PATH:  
  Backend cannot generate stream (file not found, unsupported, etc.)  
  ↓  
  Backend responds: previewError with error message and reason  
  ↓  
  Frontend receives previewError  
  ↓  
  Hide loading spinner  
  ↓  
  Display error message to user  
  ↓  
  Show Retry button (if applicable)

CLEANUP:  
  User closes modal/popup OR clicks Stop Preview  
  ↓  
  Frontend sends: stopPreview with streamId  
  ↓  
  Frontend destroys hls.js instance  
  ↓  
  Backend kills FFmpeg process  
  ↓  
  Backend cleans up HLS segments

---

### **6.10. Monitor Popup "Edit in Editor View"**

User in Monitor View  
  ↓  
Clicks event in calendar  
  ↓  
Popup opens showing event preview  
  ↓  
User sees something needs editing  
  ↓  
Clicks "Edit in Editor View" button  
  ↓  
Frontend actions (in order):  
  1\. closeMonitorPreview()  
     \- Send stopPreview if stream active  
     \- Destroy hls.js instance  
     \- Remove popup from DOM  
  ↓  
  2\. Switch to Editor View  
     \- Update tab UI  
     \- Show Editor calendar  
     \- Hide Monitor components  
  ↓  
  3\. openTaskModal(editorCalendar, { event })  
     \- Open full edit modal  
     \- Populate with event data  
  ↓  
User now in Editor View with modal open  
  ↓  
Can make changes and commit

---

### **6.11. Current Program Tracking**

Backend scheduler evaluates every second:  
  ↓  
Backend sends WebSocket message:  
  'scheduler.state.targetProgram'  
  ↓  
Frontend receives message:  
  ↓  
appState.updateCurrentProgram(payload)  
  ↓  
Monitor calendar re-renders:  
  \- Removes 'current-program' class from previous  
  \- Adds 'current-program' class to new current event  
  \- Applies pulsing glow animation  
  \- Shows "▶ LIVE" badge  
  ↓  
User sees which event is broadcasting NOW

---

## **7\. WebSocket Message Protocol**

### **7.0. Preview Protocol**

See **Section 5.3** for complete Event Preview WebSocket protocol including:

* `requestPreview` (client → server)  
* `previewReady` (server → client, success)  
* `previewError` (server → client, failure)  
* `stopPreview` (client → server, cleanup)

---

### **6.1. Outgoing Messages (Client → Server)**

**getSchedule:**

{  
    "action": "getSchedule",  
    "payload": {}  
}

**commitSchedule:**

{  
    "action": "commitSchedule",  
    "payload": {  
        "version": "1.0",  
        "scheduleName": "string",  
        "schedule": \[ /\* Schedule 1.0 events \*/ \]  
    }  
}

---

### **6.2. Incoming Messages Processed by Frontend**

**currentSchedule:**

{  
    "action": "currentSchedule",  
    "payload": {  
        "version": "1.0",  
        "scheduleName": "Main Schedule",  
        "schedule": \[ /\* events \*/ \]  
    }  
}

**Trigger:** Response to explicit getSchedule request  
 **Action:**

* Update `appState.serverSchedule`  
* Monitor calendar refreshes  
* If first load: Editor syncs with server

---

**scheduleChanged:**

{  
    "action": "scheduleChanged",  
    "payload": {  
        "version": "1.0",  
        "scheduleName": "string",  
        "schedule": \[ /\* new schedule \*/ \]  
    }  
}

**Trigger:** Backend FileWatcher detected schedule.json change  
 **Action:**

* Update `appState.serverSchedule`  
* Monitor calendar auto-refreshes  
* Editor calendar NOT affected (isolation)  
* Show passive notification: "Schedule updated on server"

---

**commitSuccess:**

{  
    "action": "commitSuccess",  
    "payload": {}  
}

**Action:**

* `appState.markEditorClean()`  
* `updateStatusBar('clean', 'Synced with server')`  
* Show success notification

---

**commitError:**

{  
    "action": "commitError",  
    "payload": {  
        "message": "Validation failed: Event 3 missing source.uri"  
    }  
}

**Action:**

* `updateStatusBar('error', 'Save failed')`  
* `alert(payload.message)`  
* User can fix issues and retry

---

**scheduler.state.targetProgram:**

{  
    "action": "scheduler.state.targetProgram",  
    "payload": {  
        "timestamp": "2025-10-10T17:06:35Z",  
        "targetProgram": {  
            "id": "evt-123",  
            "title": "Bloomberg Live",  
            "start": "2025-10-10T17:00:00Z",  
            "end": "2025-10-10T18:00:00Z"  
        },  
        "nextProgram": {  
            "id": "evt-124",  
            "title": "Weather Report",  
            "start": "2025-10-10T18:00:00Z"  
        }  
    }  
}

**Trigger:** Published every second by scheduler  
 **Action:**

* Update `appState.currentProgram`  
* Monitor calendar highlights current event  
* Info window logs the update

**`obs.program.changed`**

**payload:** `{ timestamp, previousProgram?, newProgram, trigger, seekOffsetMs?, correlationId? }`

**Trigger:** Emitted whenever OBS finishes switching Program.

**Action:** Update `appState.currentProgram` and `appState.nextProgram` as needed; re-render Monitor calendar highlight; log to Activity.

---

### **6.3. Backend State Events (Logged Only)**

The backend publishes numerous operational events via WebSocket. The frontend receives ALL these events but only LOGS them in the Info Window. These events do NOT trigger UI state changes in calendars.

**Event Categories:**

* Media Source lifecycle (ready/lost/stopped/failed)  
* OBS system (connected/disconnected/scene changed)  
* Virtual camera (started/stopped)  
* Streaming/recording state  
* WebSocket client connections  
* WebRTC peer connections

**Display Format:**

\[17:06:25\] webserver.lifecycle.started: port 8080  
\[17:06:30\] obs.system.connected: OBS v30.0.0  
\[17:06:31\] obs.virtualcam.started  
\[17:06:31\] mediasource.lifecycle.ready: OBS Virtual Camera  
\[17:06:35\] scheduler.state.targetProgram: Bloomberg Live  
\[17:10:00\] obs.stream.state.changed: streaming started

**Purpose:** Passive monitoring and debugging visibility

---

## **8\. Schedule 1.0 Format**

The `schedule.json` file defines what programs (events) should play in OBS and when. This is the data format that both backend and frontend understand.

### **7.1. Complete Schema**

{  
  "version": "1.0",  
  "scheduleName": "Main Schedule",  
  "schedule": \[  
    {  
      "id": "unique-id-string",  
      "title": "Program Name",  
      "enabled": true,  
      "general": {  
        "description": "Optional description",  
        "tags": \["tag1", "tag2"\],  
        "classNames": \["css-class"\],  
        "textColor": "\#FFFFFF",  
        "backgroundColor": "\#3788D8",  
        "borderColor": "\#FF0000"  
      },  
      "source": {  
        "name": "obs\_source\_name",  
        "inputKind": "ffmpeg\_source",  
        "uri": "/path/to/video.mp4",  
        "inputSettings": {  
          "local\_file": "/path/to/video.mp4",  
          "looping": true  
        },  
        "transform": {  
          "positionX": 0,  
          "positionY": 0,  
          "scaleX": 1.0,  
          "scaleY": 1.0  
        }  
      },  
      "timing": {  
        "start": "2025-01-15T14:00:00Z",  
        "end": "2025-01-15T15:00:00Z",  
        "isRecurring": false,  
        "recurrence": {  
          "daysOfWeek": \["MON", "WED", "FRI"\],  
          "startRecur": "2025-01-01",  
          "endRecur": "2025-12-31"  
        }  
      },  
      "behavior": {  
        "onEndAction": "hide",  
        "preloadSeconds": 5  
      }  
    }  
  \]  
}

---

### **7.2. Field Descriptions**

**Root Level:**

* `version`: Always "1.0" (for future compatibility)  
* `scheduleName`: Human-readable schedule name  
* `schedule`: Array of event objects

**Event Object \- Core Fields:**

* `id`: Unique identifier (UUID recommended)  
* `title`: Display name in calendar AND default OBS source name  
* `enabled`: If false, backend ignores this event

**general Section (Visual Only \- Backend Ignores):**

* `description`: Notes about the event  
* `tags`: Array of keywords for categorization  
* `classNames`: CSS classes applied in calendar view  
* `textColor`, `backgroundColor`, `borderColor`: Visual styling

**source Section (What OBS Will Display):**

* `name`: Technical name in OBS (must be unique)  
* `inputKind`: OBS source type (ffmpeg\_source, browser\_source, etc.)  
* `uri`: Path to media file or URL  
* `inputSettings`: JSON object with settings specific to inputKind  
* `transform`: Position and scale in OBS scene

**timing Section (When to Show):**

* `start`: ISO 8601 datetime with 'Z' suffix (UTC)  
* `end`: ISO 8601 datetime with 'Z' suffix (UTC)  
* `isRecurring`: If true, event repeats on specified days  
* `recurrence.daysOfWeek`: Array of day codes: "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"  
* `recurrence.startRecur`: First date to apply recurrence (YYYY-MM-DD)  
* `recurrence.endRecur`: Last date to apply recurrence (YYYY-MM-DD)

**behavior Section (Backend Actions):**

* `onEndAction`: What to do when event ends ("hide", "stop", "none")  
* `preloadSeconds`: Seconds in advance to prepare source

---

### **7.3. OBS inputKind Types**

Common values for `source.inputKind`:

* `ffmpeg_source` \- Video/audio files  
* `browser_source` \- Web pages  
* `image_source` \- Static images  
* `vlc_source` \- VLC playlist  
* `ndi_source` \- NDI stream  
* `color_source` \- Solid color  
* `text_ft2_source` / `text_gdiplus` \- Text overlays  
* `window_capture` \- Window capture  
* `display_capture` \- Screen capture

Each inputKind has specific required fields in inputSettings. Consult OBS documentation for details.

---

### **7.4. FullCalendar Mapping**

**Recurring events in Schedule 1.0:**

{  
  "timing": {  
    "start": "1970-01-01T14:00:00Z",  
    "end": "1970-01-01T15:00:00Z",  
    "isRecurring": true,  
    "recurrence": {  
      "daysOfWeek": \["MON", "WED", "FRI"\],  
      "startRecur": "2025-01-01",  
      "endRecur": "2025-12-31"  
    }  
  }  
}

**Maps to FullCalendar:**

{  
  daysOfWeek: \[1, 3, 5\],  // Numbers: 0=Sun, 1=Mon, ...  
  startTime: "14:00:00",  
  endTime: "15:00:00",  
  startRecur: "2025-01-01",  
  endRecur: "2025-12-31"  
}

**Critical:** Schedule Adapter must handle this conversion correctly in both directions.

---

### **7.5. Validation Rules**

**Required Fields:**

* `id` \- Must be unique  
* `title` \- Cannot be empty  
* `source.name` \- Cannot be empty, must be unique  
* `source.inputKind` \- Must be valid OBS input type  
* `source.uri` \- Cannot be empty for most input types  
* `timing.start` \- Must be valid ISO 8601 with 'Z'  
* `timing.end` \- Must be valid ISO 8601 with 'Z', must be after start

**Optional Fields with Defaults:**

* `enabled` \- Defaults to true  
* `general.*` \- All optional, used only for display  
* `source.inputSettings` \- Defaults to {}  
* `source.transform` \- Defaults to {}  
* `behavior.onEndAction` \- Defaults to "hide"  
* `behavior.preloadSeconds` \- Defaults to 0

**Recurring Event Rules:**

* If `timing.isRecurring` is true:  
  * `timing.recurrence.daysOfWeek` must have at least one day  
  * `timing.start` and `timing.end` represent time-of-day only  
  * Date part ignored (use 1970-01-01 as template)  
  * `timing.recurrence.startRecur` and `endRecur` define date range

---

## **9\. User Interface Components**

### **8.1. Overall Layout**

**Two-Tab Structure:**

┌─────────────────────────────────────────────────┐  
│  \[📺 Monitor\]  \[📝 Editor ●\]      ← Tab bar     │  
├─────────────────────────────────────────────────┤  
│                                                 │  
│  Active view content (Monitor OR Editor)        │  
│  Only one visible at a time                     │  
│                                                 │  
└─────────────────────────────────────────────────┘

**Monitor Tab Active:**

┌─────────────────────────────────────────────────┐  
│  \[📺 Monitor ✓\]  \[📝 Editor\]                    │  
├──────────────────┬──────────────────────────────┤  
│  Left Sidebar    │  Monitor Calendar            │  
│  ┌────────────┐  │  ┌────────────────────────┐ │  
│  │  🎥 Live   │  │  │ 14:00 Bloomberg ◄ NOW  │ │  
│  │  Preview   │  │  │ 15:00 Weather          │ │  
│  │            │  │  │ 16:00 Sports           │ │  
│  └────────────┘  │  └────────────────────────┘ │  
│  ┌────────────┐  │                             │  
│  │  📊 Log    │  │  🔒 Read-only               │  
│  │  \[✓\] Conn  │  │  ⚡ Synced with server      │  
│  │  \[10:30\]   │  │                             │  
│  └────────────┘  │                             │  
└──────────────────┴──────────────────────────────┘

**Editor Tab Active:**

┌─────────────────────────────────────────────────┐  
│  \[📺 Monitor\]  \[📝 Editor ✓ ●\]                  │  
├─────────────────────────────────────────────────┤  
│  📝 Schedule Editor                             │  
│  ┌────────────────────────────────────────────┐ │  
│  │ Toolbar: \[prev\]\[next\]\[today\] \+ \- ...      │ │  
│  │ 14:00 Bloomberg                            │ │  
│  │ 15:00 Weather (modified) ⚠️                │ │  
│  │ 16:30 Sports (moved) ⚠️                    │ │  
│  └────────────────────────────────────────────┘ │  
│                                                 │  
│  \[⚠️ 2 unsaved changes\]                         │  
│  \[↩️ Revert\] \[🔄 Sync\] \[💾 Commit to Server\]    │  
└─────────────────────────────────────────────────┘

---

### **8.2. Notification Strategy**

**Pattern:** Native browser alerts and confirms  
 **Rationale:** Zero code to maintain, universally understood, works everywhere

**Usage:**

Validation errors:

alert('Cannot save: Event missing source.uri');

Confirmations:

if (\!confirm('Delete this event?')) return;

Revert confirmation:

if (confirm('Discard all unsaved changes?')) {  
    appState.syncEditorWithServer();  
}

Error handling:

alert('Save failed: ' \+ errorMessage);

**Future Enhancement:** Can be upgraded to custom toast notifications later. Start simple.

---

### **8.3. Status Bar (Editor View Only)**

**Visual States:**

* **Clean:** Green circle \+ "Synced with server"  
* **Dirty:** Orange circle \+ "2 unsaved changes"  
* **Syncing:** Blue circle \+ "Saving..."  
* **Error:** Red circle \+ "Save failed"

**Location:** Bottom of Editor calendar panel  
 **Always visible:** User always knows sync state

**HTML Structure:**

\<div id="status-bar" class="status-bar clean"\>  
    \<div id="status-circle"\>\</div\>  
    \<span id="status-text"\>Synced with server\</span\>  
\</div\>

---

### **8.4. Action Buttons (Editor View Only)**

**Three Buttons:**

1. **Revert Changes** (btn-secondary)

   * Disabled when clean  
   * Enabled when dirty  
   * Confirmation required  
   * Discards all changes, reloads from server  
2. **Sync from Server** (btn-secondary)

   * Always enabled  
   * Confirmation required if dirty  
   * Pulls latest server schedule  
3. **Commit to Server** (btn-primary)

   * Disabled when clean  
   * Enabled when dirty  
   * Sends changes to server  
   * Shows syncing state during operation

---

### **8.5. Calendar Menu (Editor View Only)**

**Trigger:** Three-dot button (...) in FullCalendar toolbar

**Menu Items:**

New Schedule  
Load from File  
Save to File  
──────────────  
Get from Server  
Commit to Server

**Implementation:**

* Custom dropdown positioned relative to button  
* Created dynamically by `menu.mjs`  
* Attached to document.body with absolute positioning  
* Z-index above calendar  
* Closes when clicking outside

**Visual Design:**

.calendar-menu {  
    position: absolute;  
    z-index: 1000;  
    background: var(--bg-panel);  
    border: 1px solid var(--border);  
    border-radius: var(--radius);  
    box-shadow: var(--shadow);  
    width: 160px;  
}

.menu-item {  
    padding: 8px 12px;  
    font-size: .875rem;  
    cursor: pointer;  
    color: var(--text-primary);  
    transition: background-color 0.2s;  
}

.menu-item:hover {  
    background: var(--surface);  
}

---

### **9.6. Edit Modal Structure**

**Five Tabs:**

**Tab 1: General**

* Title (required)  
* Enabled checkbox  
* Description  
* Tags (space-separated)  
* Class Names  
* Text Color (color picker with clear)  
* Background Color (color picker with clear)  
* Border Color (color picker with clear, hint: "For recurring events only")

**Tab 2: Source (OBS Configuration)**

* Input Name (required) \- Technical OBS source name  
* Input Type (required) \- Dropdown of OBS input kinds  
* URI (required) \- File path or URL (with hints based on input type)  
* Settings (JSON) \- Text area for inputSettings object  
* Transform (JSON) \- Text area for transform object

**Tab 3: Timing**

* Start Date/Time (required)  
* End Date/Time (required)  
* Duration Display (auto-calculated, read-only)  
* Recurring checkbox  
* Days of Week (checkboxes for MON-SUN, shown when recurring)  
* Recurrence Start (date picker, shown when recurring)  
* Recurrence End (date picker, shown when recurring)  
* Hint: "For recurring events only the time part of Start/End is used"

**Tab 4: Behavior (Advanced)**

* Preload Seconds (number input)  
* On End Action (dropdown: hide/stop/none)

**Tab 5: Preview**

* Video preview element with HLS player  
* "Preview Source" button to load stream  
* Standard video controls (play/pause/seek/volume)  
* "Stop Preview" button when playing  
* Loading spinner during stream generation  
* Error display if preview fails  
* See **Section 5.4** for complete Tab 5 implementation details

**Modal Features:**

* Draggable by header  
* Tabbed interface  
* Color pickers with clear buttons  
* Auto-generate source name from title  
* Duration auto-calculation  
* JSON validation for inputSettings and transform  
* HLS preview in Tab 5

---

### **9.7. Monitor Event Preview Popup**

**Trigger:** Click any event in Monitor View calendar

**Purpose:** Quick preview of event content without leaving Monitor View

**Implementation:** See **Section 5.5** for complete details

**Visual Summary:**

┌─────────────────────────────────────────────┐  
│  Bloomberg Live                        \[✕\]  │  
├─────────────────────────────────────────────┤  
│  \[Video Player with HLS Stream\]             │  
│  \[▶ Preview Source\]                         │  
├─────────────────────────────────────────────┤  
│  Source: /media/bloomberg.mp4               │  
│  Input: ffmpeg\_source                       │  
│  Time: 14:00 \- 15:00 (Mon, Wed, Fri)        │  
├─────────────────────────────────────────────┤  
│  \[Edit in Editor View\]  \[Close\]             │  
└─────────────────────────────────────────────┘

**Information Displayed (Minimal):**

* Event title  
* Source URI and input kind  
* Schedule timing  
* HLS video preview

**Features:**

* Modal-style overlay  
* Click outside or ✕ to close  
* "Edit in Editor View" switches to Editor and opens full modal  
* Same HLS preview functionality as Editor modal  
* Automatic stream cleanup on close

---

### **9.8. Current Program Highlight (Monitor View Only)**

**Visual Indicators:**

* **Box shadow:** Pulsing green glow around event  
* **Badge:** "▶ LIVE" in top-right corner  
* **Animation:** Pulse effect (2s cycle)  
* **Z-index:** Above other events

**Implementation:**

.fc-event.current-program {  
    box-shadow: 0 0 0 3px var(--color-current-program);  
    animation: pulse-glow 2s ease-in-out infinite;  
}

.fc-event.current-program::before {  
    content: "▶ LIVE";  
    position: absolute;  
    top: \-2px;  
    right: \-2px;  
    background: var(--color-current-program);  
    color: white;  
    font-size: 0.65rem;  
    font-weight: bold;  
    padding: 2px 6px;  
    border-radius: 3px;  
}

---

### **9.9. Dirty Indicator (Editor Tab)**

**Visual:** Small orange dot on Editor tab label

**Shows when:**

* `appState.isDirty === true`  
* User has uncommitted changes

**Hides when:**

* `appState.isDirty === false`  
* Changes committed or reverted

**Implementation:**

\<button class="view-tab" data-view="editor"\>  
    \<span\>📝 Editor\</span\>  
    \<span class="dirty-indicator" style="display: none;"\>●\</span\>  
\</button\>

---

## **10\. Error Handling**

### **10.1. Error Categories**

| Error Type | Handling | User Experience |
| ----- | ----- | ----- |
| Validation | alert() with message | Immediate feedback, stays in modal to fix |
| Network | Status bar \+ console | Visual indicator, auto-retry for connection |
| Server rejection | alert() \+ status bar | Shows specific error message |
| Parse error | alert() \+ console.error | Shows specific JSON error |
| File read error | alert() with details | Shows file read failure reason |
| Preview error | Display in preview area | Shows backend error message, optional retry |

---

### **10.2. Validation**

**Where:** Modal form before saving event to FullCalendar

**What:**

* Required fields (title, source.name, source.uri)  
* JSON fields well-formed (inputSettings, transform)  
* inputKind in allowed list  
* Times: start before end

**Server Validation:** Server is authoritative. Even if frontend validates, server validates again. Server rejection shows alert with error message.

---

### **10.3. Preview Error Handling**

**File Not Found:**

* Display: "Preview Error: File not found: /path/to/video.mp4"  
* Show Retry button (in case file becomes available)

**Unsupported on Platform:**

* Display: "Preview Error: Cannot preview window\_capture on headless server"  
* No Retry button (won't work)

**Stream Generation Failed:**

* Display: "Preview Error: Stream generation failed"  
* Show Retry button

**HLS Playback Error:**

* Display: "Playback Error: Stream interrupted"  
* Show Retry button

See **Section 5.8** for complete preview error handling details.

---

### **10.4. Logging**

**Console:**

* All WebSocket messages  
* State transitions (isDirty changes, view switches)  
* All errors with context  
* File operations (load/save)  
* Preview stream requests/responses  
* HLS player events and errors

**Info Window:**

* User-visible activity log  
* Connection status changes  
* Server messages  
* Backend operational events

---

## **11\. CSS Architecture**

### **10.1. Root Variables (Single Source of Truth)**

:root {  
  /\* Colors \*/  
  \--primary: \#3b82f6;  
  \--primary-hover: \#2563eb;  
  \--success: \#10b981;  
  \--danger: \#ef4444;  
  \--secondary: \#64748b;

  /\* Greyscale & Layout \*/  
  \--bg-main: \#f0f2f5;  
  \--bg-panel: \#ffffff;  
  \--surface: \#f8fafc;  
  \--border: \#e2e8f0;

  /\* Text \*/  
  \--text-primary: \#1c1e21;  
  \--text-muted: \#64748b;

  /\* Sizing & Shadow \*/  
  \--radius: 8px;  
  \--shadow: 0 4px 12px rgba(0, 0, 0, 0.1);

  /\* View-specific \*/  
  \--view-header-height: 56px;  
  \--monitor-sidebar-width: 350px;  
  \--editor-actions-height: 60px;

  /\* Status indicators \*/  
  \--color-current-program: \#22c55e;  
  \--color-dirty: \#f59e0b;  
}

---

### **10.2. Layout System**

**Base Layout:**

\#app {  
  display: flex;  
  flex-direction: column;  
  height: 100vh;  
}

.view-switcher {  
  height: var(--view-header-height);  
  /\* Tab navigation bar \*/  
}

.app-view {  
  display: none;  
  flex: 1;  
  overflow: hidden;  
}

.app-view.active {  
  display: flex;  
}

**Monitor Layout:**

.monitor-layout {  
  display: grid;  
  grid-template-columns: var(--monitor-sidebar-width) 1fr;  
  gap: 1rem;  
  height: 100%;  
}

**Editor Layout:**

.editor-layout {  
  display: flex;  
  flex-direction: column;  
  height: 100%;  
}

\#editor-calendar {  
  flex: 1;  
  min-height: 0;  
}

.editor-actions {  
  height: var(--editor-actions-height);  
  /\* Action buttons \*/  
}

---

### **11.3. Component Styling**

**Each component has its own CSS file:**

* `view-switcher.css` \- Tab navigation  
* `monitor-view.css` \- Monitor layout  
* `editor-view.css` \- Editor layout  
* `calendar.css` \- Calendar overrides  
* `modal.css` \- Modal styling  
* `monitor-preview.css` \- Monitor popup preview  
* `info-window.css` \- Activity log  
* `live-preview.css` \- Video preview

**Import Order (in index.html):**

\<link rel="stylesheet" href="main.css"\>  
\<link rel="stylesheet" href="components/view-switcher/view-switcher.css"\>  
\<link rel="stylesheet" href="components/monitor-view/monitor-view.css"\>  
\<link rel="stylesheet" href="components/editor-view/editor-view.css"\>  
\<link rel="stylesheet" href="components/calendar/calendar.css"\>  
\<link rel="stylesheet" href="components/calendar/modal.css"\>  
\<link rel="stylesheet" href="components/monitor-preview/monitor-preview.css"\>  
\<link rel="stylesheet" href="components/info-window/info-window.css"\>  
\<link rel="stylesheet" href="components/live-preview/live-preview.css"\>

---

## **12\. Browser Compatibility**

**Minimum Requirements:**

* ES6 modules (import/export)  
* WebSocket API  
* CSS Grid  
* CustomEvent  
* Fetch API  
* WebRTC (for live preview)  
* File API (for load/save local files)  
* Media Source Extensions (for HLS playback)

**Supported Browsers:**

* Chrome/Edge 90+  
* Firefox 88+  
* Safari 14+

**HLS Support:**

* Chrome/Firefox: Via hls.js library  
* Safari: Native HLS support (fallback)

**No Polyfills:** Modern browsers only.

---

## **13\. Security**

**XSS Prevention:**

* All user input sanitized  
* FullCalendar escapes event data  
* Never use innerHTML with user content

**WebSocket Security:**

* Inherits authentication from backend WebServer (Basic Auth if configured)  
* Uses WSS when backend has TLS enabled  
* No additional frontend authentication

**File Operations Security:**

* File picker restricts to .json files  
* JSON parsing wrapped in try-catch  
* Validates Schedule 1.0 format before importing  
* No arbitrary code execution

**HLS Stream Security:**

* Streams served via HTTP from backend  
* Backend validates sourceUri before generating stream  
* Backend enforces timeout on idle streams  
* Stream URLs are ephemeral (contain streamId)  
* No direct file system access from frontend

**Validation:**

* Server is authoritative  
* Frontend validation is UX convenience only  
* Never trust client-side validation alone

---

## **14\. Implementation Guidelines**

### **13.1. File Header Convention**

Every file must start with:

// File: \[full-path-from-frontend-root\]

Examples:

// File: main.mjs  
// File: components/calendar/calendar-editor.mjs  
// File: components/calendar/modal/form.mjs

---

### **13.2. Module Pattern**

**Exports:**

* Use named exports for multiple functions  
* Use default export only for service objects

**Imports:**

* Import only what you need  
* Keep imports organized at top of file

**Example:**

// File: components/calendar/helpers.mjs

export function pad2(n) {  
  return String(n).padStart(2, '0');  
}

export function genId(prefix \= 'evt-') {  
  return \`${prefix}${Math.random().toString(36).slice(2, 9)}\`;  
}

---

### **13.3. State Management Pattern**

The application uses a simple, manual state management pattern for maximum clarity.

**1. The State (`app-state.mjs`):**
A single, exported JavaScript object holds simple flag variables.

```javascript
// File: shared/app-state.mjs
export const AppState = {
    isDirty: false,
    // ...
};
```

**2. The Update Logic (`ui-updater.mjs`):**
A single, exported function is responsible for reading the state and updating the DOM.

```javascript
// File: shared/ui-updater.mjs
import { AppState } from './app-state.mjs';

export function updateUI() {
    // Logic to update status bar based on AppState.isDirty
    // Logic to enable/disable buttons based on AppState.isDirty
    // etc.
}
```

**3. The Usage Pattern:**
Any component that needs to change the application's state must perform two steps: modify the state, then call the updater.

```javascript
// In any component file
import { AppState } from './shared/app-state.mjs';
import { updateUI } from './shared/ui-updater.mjs';

function handleSomethingThatChangesState() {
    // Step 1: Modify the state directly
    AppState.isDirty = true;

    // Step 2: Manually trigger the UI update
    updateUI();
}
```

---

### **13.4. Event Communication**

**Dispatch custom events:**

document.dispatchEvent(new CustomEvent('view:changed', {  
    detail: { view: 'editor' }  
}));

**Listen to custom events:**

document.addEventListener('view:changed', (e) \=\> {  
    console.log('Switched to:', e.detail.view);  
});

---

### **13.5. Component Initialization**

**Pattern:**

// File: components/example/example.mjs

export function initExample(container) {  
    // 1\. Get DOM elements  
    const element \= container.querySelector('\#target');  
      
    // 2\. Setup state  
    let localState \= {};  
      
    // 3\. Subscribe to global state  
    appState.subscribe('key', callback);  
      
    // 4\. Setup event listeners  
    element.addEventListener('click', handler);  
      
    // 5\. Return cleanup function (optional)  
    return () \=\> {  
        // Cleanup code  
    };  
}

---

## **15\. Glossary**

| Term | Definition |
| ----- | ----- |
| **Monitor View** | Read-only tab showing live schedule from server \+ OBS preview \+ activity log |
| **Editor View** | Editable tab with working copy of schedule for experimenting with changes |
| **Server State** | Schedule currently active on backend (schedule.json) |
| **Working Copy** | User's draft schedule in Editor view (may differ from server) |
| **Current Program** | Event actively broadcasting right now (highlighted in Monitor view) |
| **isDirty** | Boolean flag indicating Editor has uncommitted changes |
| **Commit** | Action to save Editor's working copy to server (becomes new server state) |
| **Revert** | Action to discard Editor changes and reload from server |
| **Sync from Server** | Explicit user action to update Editor with latest server state |
| **Load from File** | Import schedule from local .json file into Editor (marks as dirty) |
| **Save to File** | Export Editor's schedule to downloadable .json file |
| **Event Preview** | HLS video stream of a scheduled program's content, generated on-demand by backend |
| **Monitor Popup** | Quick preview popup shown when clicking events in Monitor calendar |
| **Live Preview** | Real-time video feed of OBS output (what's broadcasting now) via WHEP |
| **Source of Truth** | Authoritative version of data \- always schedule.json on server |
| **Hot-Reload** | Backend detecting schedule.json file change and notifying frontend |
| **Schedule 1.0** | JSON format specification for schedule.json |
| **WHEP** | WebRTC-HTTP Egress Protocol for live video preview |
| **HLS** | HTTP Live Streaming protocol for event content preview |
| **OBS** | Open Broadcaster Software \- streaming/recording software controlled by backend |
| **inputKind** | OBS source type (ffmpeg\_source, browser\_source, etc.) |
| **streamId** | Unique identifier for an active HLS preview stream |

---

## **END OF SPECIFICATION**

**Version:** 1.1  
 **Last Updated:** 2025-10-11

This specification is complete and self-contained. It describes the frontend as a finished, production-ready system.
