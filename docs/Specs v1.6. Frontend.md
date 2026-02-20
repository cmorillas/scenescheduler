# Scene Scheduler for OBS - Frontend Specification

**Status:** Prescriptive
**Version:** 1.6
**Date:** 2025-10-28

---

## 1. Overview

### 1.1. Purpose

This specification defines the frontend component of Scene Scheduler for OBS - a web-based application that provides visual editing and real-time monitoring capabilities for OBS automation schedules.

### 1.2. System Context

The frontend operates as a browser-based interface to an autonomous OBS automation backend. The backend independently controls OBS Studio, reads schedule configuration, and switches programs according to schedule. The frontend provides optional visual tools for schedule management and operational monitoring.

### 1.3. Core Responsibilities

**Schedule Editor**
- Provide visual calendar interface for editing schedule configuration
- Support creating, modifying, and deleting scheduled programs
- Enable import/export of schedule files
- Maintain isolated editing workspace with explicit commit workflow

**Operational Monitor**
- Display real-time backend activity logs
- Show current connection states (Server, OBS, Preview)
- Provide live video preview of broadcast output
- Display current schedule state (read-only view)

### 1.4. Key Architectural Principles

**Dual State Management**: The system SHALL maintain separate schedule states:
- Server state: Official schedule active on backend
- Working state: Local editing sandbox that can diverge

**Explicit Synchronization**: Changes in the editor SHALL NOT affect the live schedule until explicitly committed by user action.

**Reactive Architecture**: The system SHALL use centralized state management with subscribe/notify patterns for UI updates.

**Non-Goals**: The frontend SHALL NOT:
- Control OBS directly
- Command backend to switch programs
- Require authentication (inherited from backend)
- Support offline editing
- Provide real-time collaboration

---

## 2. System Architecture

### 2.1. Component Overview

```
┌──────────────────────────────────────────────┐
│  Backend (Autonomous)                        │
│  └─ WebServer (HTTP/WebSocket/WHEP)         │
└──────────────────────────────────────────────┘
                    ↕
┌──────────────────────────────────────────────┐
│  Frontend Application                        │
│  ├─ State Management (reactive)              │
│  ├─ Monitor View                             │
│  │  ├─ Activity Log                          │
│  │  ├─ Live Preview (WHEP)                   │
│  │  └─ Read-only Calendar                    │
│  └─ Editor View                              │
│     ├─ Editable Calendar                     │
│     ├─ Schedule Operations                   │
│     └─ Status Indicators                     │
└──────────────────────────────────────────────┘
```

### 2.2. State Management Model

**Centralized State**

The application SHALL maintain a single source of truth for application state with the following structure:

- **Connection State**: WebSocket connection status
- **OBS State**: Backend-to-OBS connection status and version
- **Preview State**: VirtualCam stream availability
- **Schedule State (Dual)**:
  - Server schedule: Official schedule from backend
  - Working schedule: Local editing copy
- **Editor State**: Dirty tracking, change count, sync status
- **View State**: Current active view (monitor or editor)

**Reactive Updates**

Components SHALL subscribe to specific state paths and receive automatic notifications when those paths change. State updates SHALL trigger re-rendering of subscribed components only.

### 2.3. Dual View Architecture

**Monitor View (Read-only)**

Purpose: Observe actual system state
- Calendar displays server schedule with no editing capabilities
- Activity log shows backend operation events
- Live preview displays current broadcast output
- Modal interactions show event details in read-only mode

**Editor View (Editable)**

Purpose: Safe workspace for schedule modifications
- Calendar displays working copy with full editing capabilities
- Changes automatically tracked as "dirty" state
- Explicit synchronization actions (commit, revert, load)
- Isolated from server updates when uncommitted changes exist

### 2.4. Communication Protocols

**WebSocket Protocol**

The frontend SHALL communicate with backend using JSON messages over WebSocket:

Message format: `{"action": "string", "payload": {}}`

**WHEP Protocol**

The frontend SHALL use WebRTC-HTTP Egress Protocol (WHEP) for live video streaming:
- POST to `/whep` endpoint with SDP offer
- Receive SDP answer for WebRTC connection
- Handle 503 responses gracefully (stream unavailable is expected state)

---

## 3. Data Model

### 3.1. Schedule Format (Version 1.0)

The system SHALL support schedule documents with the following structure:

```json
{
  "version": "1.0",
  "scheduleName": "string",
  "schedule": [
    {
      "id": "string",
      "title": "string",
      "enabled": boolean,
      "general": {
        "description": "string",
        "tags": ["string"],
        "classNames": ["string"],
        "textColor": "#RRGGBB",
        "backgroundColor": "#RRGGBB",
        "borderColor": "#RRGGBB"
      },
      "source": {
        "name": "string",
        "inputKind": "string",
        "uri": "string",
        "inputSettings": {},
        "transform": {}
      },
      "timing": {
        "start": "ISO8601",
        "end": "ISO8601",
        "isRecurring": boolean,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", ...],
          "startRecur": "YYYY-MM-DD",
          "endRecur": "YYYY-MM-DD"
        }
      },
      "behavior": {
        "onEndAction": "hide|none|stop",
        "preloadSeconds": number
      }
    }
  ]
}
```

### 3.2. State Model

**Application State Structure**

```javascript
{
  websocket: {
    isConnected: boolean,
    status: string,
    statusText: string
  },
  obs: {
    connected: boolean,
    version: string,
    statusText: string
  },
  preview: {
    available: boolean,
    status: 'unavailable' | 'available' | 'connecting' | 'connected',
    statusText: string
    // 'unavailable': VirtualCam stopped or stream unavailable
    // 'available': VirtualCam active, can play (even if currently paused)
    // 'connecting': WebRTC connection in progress
    // 'connected': WebRTC connected, actively streaming
  },
  editor: {
    isDirty: boolean,
    changeCount: number,
    isSyncing: boolean,
    status: 'clean' | 'dirty' | 'syncing' | 'error',
    statusText: string
  },
  schedule: Schedule,           // Server state
  workingSchedule: Schedule,    // Editor state
  currentView: 'monitor' | 'editor'
}
```

---

## 4. Component Specifications

### 4.1. State Manager

**Responsibilities**
- Maintain centralized application state
- Provide subscription mechanism for reactive updates
- Manage dual schedule state (server and working)
- Track editor dirty state
- Coordinate state synchronization

**Key Operations**

`subscribe(path, callback)` - Register listener for state changes at specific path

`setSchedule(schedule, options)` - Update server schedule state
- SHALL update working schedule if no unsaved changes exist
- SHALL NOT update working schedule if dirty flag is true
- SHALL prompt user for confirmation if manual load with unsaved changes

`setWorkingSchedule(schedule)` - Update editor working copy
- SHALL mark editor as dirty
- SHALL update change count
- SHALL trigger subscribed component updates

### 4.2. WebSocket Service

**Responsibilities**
- Establish and maintain WebSocket connection to backend
- Translate between JSON messages and internal events
- Request initial state synchronization on connect
- Handle connection lifecycle (connect, disconnect, reconnect)

**Connection and Reconnection Flow**

On successful WebSocket connection (initial or reconnect), the service SHALL:
1. Send `getStatus` message to request current backend state
2. Send `getSchedule` message to load current schedule
3. Update connection state indicators
4. Enable appropriate UI controls

**Automatic Reconnection**

When connection is lost, the service SHALL:
1. Update status indicators to disconnected state
2. Automatically attempt reconnection after 5-second delay
3. Re-execute full connection flow on successful reconnect
4. Re-synchronize all state (status and schedule) to account for missed events
5. Suppress redundant error logging during reconnection attempts

This ensures accurate state display regardless of connection interruptions while minimizing console noise.

**Message Handling**

The service SHALL dispatch received messages to appropriate handlers based on action type.

### 4.3. Monitor View Components

**Activity Log**

SHALL display real-time backend operation messages:
- Connection status changes
- Schedule reload events
- Program switch notifications
- Error and warning messages

SHALL filter verbose messages (e.g., complete schedule payloads) for readability.

**Live Preview**

SHALL provide WHEP-based video player for OBS output:
- Display video element for stream rendering
- Provide play/pause controls
- Show connection status
- Handle stream unavailable states gracefully

**Stream Control Behavior:**

SHALL automatically establish WebRTC connection on play button activation.

SHALL disconnect WebRTC session on pause, while maintaining preview availability state. User pause is a playback control action, not a stream availability change.

SHALL update preview status to unavailable only when stream is genuinely unavailable (VirtualCam stopped, connection error, remote stream ended).

SHALL allow immediate reconnection via play button when stream remains available.

**Monitor Calendar**

SHALL display read-only calendar view of server schedule:
- Show all scheduled programs
- Highlight currently active program
- Allow clicking events to view details (read-only modal)
- Update automatically when server schedule changes

### 4.4. Editor View Components

**Editor Calendar**

SHALL provide full editing capabilities:
- Create new events via time range selection
- Modify existing events via drag, resize, or modal edit
- Delete events
- Prevent event overlaps
- Support recurring event patterns

SHALL mark all modifications as dirty state.

**Schedule Operations**

The editor SHALL provide the following operations:

`New Schedule` - Clear all events with confirmation if unsaved changes exist

`Load from File` - Import schedule from local JSON file with validation

`Save to File` - Export working schedule to local JSON file

`Get from Server` - Replace working copy with server schedule
- SHALL prompt for confirmation if unsaved changes exist

`Commit to Server` - Send working schedule to backend for persistence
- SHALL validate schedule structure
- SHALL update status indicators during save
- SHALL clear dirty flag on success

**Status Bar**

SHALL display current editor state:

| State | Indicator | Description |
|-------|-----------|-------------|
| Clean | Green | "Synced with server" |
| Dirty | Orange | "X unsaved changes" |
| Syncing | Blue | "Saving..." |
| Error | Red | Error message text |

### 4.5. Event Modal

**Dual Mode Operation**

The modal SHALL support two distinct modes:

**Edit Mode** (Editor view)
- All form fields enabled
- Save and Delete buttons visible
- Full interaction with all controls
- Form validation on save

**Read-only Mode** (Monitor view)
- All form fields disabled
- Only Close button visible
- Display-only interaction
- No validation needed

**Form Structure**

The modal SHALL organize event properties into tabs:

**General Tab**
- Title, description, tags
- Visual properties (colors, CSS classes)
- Enabled/disabled toggle

**Source Tab**
- Input name (OBS identifier)
- Input kind (source type)
- URI (media location)
- Input settings (JSON)
- Transform (JSON for position/size)

**Timing Tab**
- Start/end date-time
- Recurring toggle
- Recurrence pattern (days, date range)

**Behavior Tab**
- On-end action
- Preload time

**Preview Tab**
- Video preview player
- Preview controls (play/stop)
- Source information display
- HLS.js integration for media sources
- Direct video playback for browser sources

### 4.6. Source Preview Component

**Responsibility:** Provide on-demand preview of program sources within the modal before scheduling.

**Unified HLS Workflow:**

The preview component SHALL use HLS streaming for ALL source types (browser_source, ffmpeg_source, media_source, vlc_source, image_source):

**Implementation Flow:**

1. **User clicks "Preview Source"**
   - Frontend displays loading spinner with source-specific message
   - browser_source: "Loading browser engine (5-10 seconds)..."
   - Other sources: "Generating preview..."

2. **Frontend sends `startPreview` WebSocket message**
   - Payload: `{inputKind, uri, inputSettings}`

3. **Backend spawns hls-generator process**
   - For browser_source: CEF renders web page (5-10s initialization)
   - For other sources: FFmpeg transcodes to HLS

4. **Backend verifies playlist has segments**
   - Waits for `#EXTINF:` tag in playlist.m3u8
   - Prevents empty playlist errors (levelEmptyError)

5. **Backend responds with `previewReady`**
   - Payload: `{hlsUrl}` (e.g., `/hls/preview-1/playlist.m3u8`)

6. **Frontend loads HLS stream using HLS.js**
   - Hides loading spinner
   - Shows video player with controls
   - Auto-plays stream

7. **Automatic timeout after 30 seconds**
   - Backend sends `previewStopped` message: `{reason: "Preview automatically stopped after 30 seconds"}`
   - Frontend receives notification BEFORE cleanup
   - HLS.js destroyed gracefully (prevents 404 errors)
   - UI shows blue info message for 5 seconds
   - Button auto-resets

8. **Manual stop (user clicks "Stop Preview" OR changes tab OR closes modal)**
   - Frontend sends `stopPreview` WebSocket message (no payload)
   - Backend kills process and cleans up files

9. **WebSocket disconnect**
   - Backend automatically cleans up any active preview

**Preview State Management:**

```javascript
previewState = {
  currentState: 'idle' | 'loading' | 'playing' | 'error' | 'unsupported',
  currentSource: {
    inputKind: string,
    uri: string,
    inputSettings: object
  },
  hls: HLS instance | null
}
```

**Component Interface:**

```javascript
// Module: frontend/public/components/calendar/modal/preview.mjs

// Initialize preview module (called when modal opens)
initPreview()

// Update preview info with source data
updatePreviewInfo(source: {inputKind, uri, inputSettings})

// Cleanup preview resources (called on modal close or tab change)
cleanupPreview()

// Handle backend responses (via custom events)
handlePreviewReady(hlsUrl)
handlePreviewError(errorMsg)
handlePreviewStopped(reason)  // New in v1.6
```

**Loading Feedback:**

The component SHALL display visual feedback during preview generation:

```javascript
// CSS: frontend/public/components/calendar/modal.css
.preview-loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.preview-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
```

```html
<!-- HTML: frontend/public/index.html -->
<div id="modal-preview-loading" class="preview-loading-container" style="display: none;">
  <div class="preview-spinner"></div>
  <div id="modal-preview-loading-text" class="preview-loading-text">Loading preview...</div>
</div>
```

**Button State Styling:**

The preview button SHALL use distinct color states for different notifications:

```css
/* Base preview button */
.btn-preview {
  background: var(--primary);
  color: #fff;
  min-width: 200px;
}

/* Info state - blue for timeout notifications */
.btn-preview.info-state {
  background: #0ea5e9; /* sky-500 */
  color: #fff;
  cursor: default;
}

/* Warning state - amber for errors */
.btn-preview.warning-state {
  background: #f59e0b; /* amber-500 */
  color: #fff;
  cursor: default;
}
```

**Cleanup Requirements:**

The preview component SHALL automatically stop and cleanup in these scenarios:
1. **User clicks "Stop Preview"**: Explicit stop action
2. **User changes tab**: Leaving Preview tab triggers cleanup
3. **User closes modal**: Modal close triggers cleanup
4. **WebSocket disconnects**: Backend automatically cleans up server-side

Cleanup actions:
- Send `stopPreview` WebSocket message to backend
- Destroy HLS.js instance to free resources
- Stop and clear video element
- Reset preview state to idle

**Error Handling:**

The component SHALL handle:
- Backend timeout (no response after 30 seconds - playlist generation)
- HLS loading failures (empty playlist, network errors)
- HLS fatal errors (automatic cleanup via stopPreview())
- Network errors
- Invalid source URIs
- Automatic preview timeout (30 seconds after start)
- Display user-friendly error/info messages in preview button

**HLS Fatal Error Detection:**

```javascript
hls.on(Hls.Events.ERROR, (event, data) => {
  if (data.fatal) {
    console.error('HLS fatal error:', data);
    // Cleanup and notify backend to kill the process
    stopPreview();
    // Show error to user
    handlePreviewError('HLS playback error: ' + data.type);
  }
});
```

**Automatic Timeout Handling:**

```javascript
function handlePreviewStopped(reason) {
  console.log('Preview stopped by server:', reason);

  // Cleanup HLS.js gracefully (prevents 404 errors)
  if (hls) {
    hls.destroy();
    hls = null;
  }

  // Reset video element
  if (dom.video) {
    dom.video.src = '';
    dom.video.load();
  }

  // Update UI to show the reason with info styling (blue background)
  setState('idle');
  dom.playBtn.classList.add('info-state');
  dom.playBtn.textContent = `ℹ ${reason}`;
  dom.playBtn.disabled = true;

  // Auto-reset after 5 seconds
  setTimeout(() => {
    dom.playBtn.classList.remove('info-state');
    dom.playBtn.textContent = '▶ Preview Source';
    dom.playBtn.disabled = false;
  }, 5000);
}
```

### 4.7. Status Indicators

The application header SHALL display three independent status indicators:

**Server Indicator** - WebSocket connection state
- Green: Connected to backend
- Red: Disconnected from backend

**OBS Indicator** - Backend-to-OBS connection state
- Green: Backend connected to OBS
- Red: Backend not connected to OBS
- Tooltip shows OBS version when connected

**Preview Indicator** - VirtualCam stream availability
- Green: Stream available (can play/replay) or actively connected
- Orange: WebRTC connection in progress
- Red: Stream not available (VirtualCam stopped)
- Note: User pause/play actions do not change availability state

---

## 5. Communication Protocols

### 5.1. WebSocket Messages

**Client to Server**

`getStatus` - Request current backend state (OBS connection, VirtualCam state)

`getSchedule` - Request current schedule

`commitSchedule` - Save schedule with full payload

`startPreview` - Request HLS preview for a source (payload: inputKind, uri, inputSettings)

`stopPreview` - Stop active source preview (no payload - tracked by WebSocket connection ID)

**Server to Client**

`currentStatus` - Response with OBS and VirtualCam state

`obsConnected` - OBS connection established notification

`obsDisconnected` - OBS connection lost notification

`virtualCamStarted` - VirtualCam stream now available

`virtualCamStopped` - VirtualCam stream no longer available

`currentSchedule` - Schedule data payload

`scheduleChanged` - Broadcast when schedule modified (hot-reload)

`commitSuccess` - Schedule save succeeded

`commitError` - Schedule save failed with error message

`previewReady` - Source preview HLS stream ready (payload: hlsUrl)

`previewError` - Source preview generation failed (payload: error)

`previewStopped` - Source preview automatically stopped (payload: reason) - **New in v1.6**

`log` - Backend activity message for display

### 5.2. WHEP Streaming

**Stream Request Flow**

1. User clicks play button
2. Frontend creates WebRTC peer connection
3. Frontend generates SDP offer
4. Frontend POSTs offer to `/whep` endpoint
5. Backend responds with SDP answer (or 503 if unavailable)
6. Frontend establishes WebRTC connection
7. Video element displays stream

**Error Handling**

503 responses SHALL be treated as expected state (stream not ready), not errors.

Network failures SHALL display appropriate error messages.

Stream SHALL automatically stop when `virtualCamStopped` message received.

---

## 6. User Interface Requirements

### 6.1. Visual Feedback

**Connection States**

All connection indicators SHALL provide immediate visual feedback using color coding and text labels.

**Editing State**

The editor status bar SHALL always display current synchronization state.

Unsaved changes SHALL be clearly indicated with change count.

**Current Program**

The monitor calendar SHALL visually highlight the currently broadcasting program with distinct styling.

**Loading States**

Operations with network activity SHALL show appropriate loading indicators.

### 6.2. User Workflows

**Schedule Editing Workflow**

1. User switches to Editor view
2. User modifies schedule (add/edit/delete events)
3. System marks state as dirty
4. User commits changes
5. System validates and sends to server
6. System marks state as clean on success

**Monitoring Workflow**

1. User switches to Monitor view
2. User observes activity log for backend operations
3. User clicks preview play button
4. System establishes WHEP connection
5. User views live broadcast output

**Event Inspection**

1. User clicks event in calendar (either view)
2. Modal opens with event details
3. In Monitor view: read-only display
4. In Editor view: full editing capabilities

### 6.3. Responsive Behavior

The calendar SHALL adapt display density based on current zoom level.

The layout SHALL support typical desktop browser dimensions.

Modal dialogs SHALL be draggable for repositioning.

---

## 7. Operational Requirements

### 7.1. Startup Behavior

On application load, the system SHALL:
1. Establish WebSocket connection
2. Request current status (OBS, VirtualCam)
3. Request current schedule
4. Initialize calendar views
5. Restore last active view from local storage

### 7.2. State Synchronization

**Automatic Synchronization**

Server schedule updates SHALL automatically refresh Monitor view calendar.

Server schedule updates SHALL NOT update Editor view when unsaved changes exist.

**Manual Synchronization**

User-initiated "Get from Server" SHALL prompt for confirmation if unsaved changes exist.

User-initiated commit SHALL show progress and final result (success or error).

### 7.3. Error Handling

**Validation Errors**

Form validation SHALL prevent saving invalid event data.

Server-side validation errors SHALL display error messages to user.

**Connection Errors**

WebSocket disconnection SHALL update connection indicator.

Failed operations SHALL display user-friendly error messages.

### 7.4. Data Persistence

The editor working copy SHALL be volatile (in-memory only).

Closing or refreshing browser with unsaved changes SHALL prompt user confirmation.

Local storage MAY be used for user preferences (last view, zoom level).

---

## 8. Technical Constraints

### 8.1. Browser Compatibility

The application SHALL target modern browsers with WebRTC and WebSocket support.

The application SHALL use native browser features over polyfills where possible.

### 8.2. Performance

Calendar rendering SHALL handle hundreds of events efficiently.

State updates SHALL only trigger re-renders of affected components.

WebRTC video SHALL provide low-latency playback.

### 8.3. Security

Authentication SHALL be inherited from backend WebServer configuration.

The application SHALL NOT store credentials locally.

All backend communication SHALL use WebSocket (upgradable to WSS with TLS).

---

## 9. Glossary

| Term | Definition |
|------|------------|
| **Working Schedule** | Local editing copy that can diverge from server |
| **Server Schedule** | Official schedule active on backend |
| **Dirty State** | Condition where working schedule differs from server schedule |
| **Commit** | Action to send working schedule to server for persistence |
| **Revert** | Action to discard working changes and reload from server |
| **WHEP** | WebRTC-HTTP Egress Protocol for low-latency streaming |
| **Monitor View** | Read-only observation interface |
| **Editor View** | Editable workspace for schedule modifications |

---

**END OF SPECIFICATION**

This specification defines requirements and architecture. Implementation details are left to designer discretion within these constraints.

---

**Last Updated:** 2025-10-17
**Document Version:** 1.5

**Changelog v1.5:**
- **Source Preview Implementation:**
  - Complete implementation in `frontend/public/components/calendar/modal/preview.mjs`
  - State management: `{currentState, currentSource, hls}`
  - Event-driven architecture with custom events `preview:ready` and `preview:error`
  - HLS.js v1.5.15 integration via CDN
  - Automatic cleanup on stop, tab change, modal close, and disconnect
- **WebSocket Protocol Updates (Section 5.1):**
  - Changed from `requestSourcePreview`/`stopSourcePreview` to `startPreview`/`stopPreview`
  - Removed `requestID` - backend tracks by client IP automatically
  - Changed from `sourcePreviewReady`/`sourcePreviewError` to `previewReady`/`previewError`
  - Simplified payload structures: `{inputKind, uri, inputSettings}` → `{hlsUrl}` or `{error}`
- **Modal Integration:**
  - Preview initialization in `modal.mjs` on modal open
  - Tab change cleanup in `ui.mjs` when leaving Preview tab
  - WebSocket message handlers in `websocket.mjs`
  - Preview tab in modal HTML with video element and controls

**Changelog v1.3:**
- **Source Preview Component (Section 4.6):**
  - Added complete specification for source preview functionality
  - Updated Event Modal with Preview Tab
  - Documented dual preview mode: browser sources (direct) vs media sources (HLS)
  - Specified HLS.js integration for media source previews
  - Defined automatic cleanup scenarios (stop, tab change, modal close, disconnect)
- **WebSocket Protocol Updates (Section 5.1):**
  - Changed from `requestSourcePreview`/`stopSourcePreview` to `startPreview`/`stopPreview`
  - Removed `requestID` - backend tracks by client IP automatically
  - Changed from `sourcePreviewReady`/`sourcePreviewError` to `previewReady`/`previewError`
  - Simplified payload structures for clarity
- **Implementation Details:**
  - Preview module: `frontend/public/components/calendar/modal/preview.mjs`
  - State management: `{currentState, currentSource, hls}`
  - Event-driven architecture: Custom events `preview:ready` and `preview:error`
  - Integration points: `modal.mjs` (init), `ui.mjs` (tab change cleanup)
  - HLS.js CDN: v1.5.15 from jsdelivr
