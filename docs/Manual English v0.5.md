# Scene Scheduler — User Manual

**Version:** 0.5  
**Date:** February 2026  
**Application:** Scene Scheduler for OBS Studio

---

## Table of Contents

1. [Getting Started](#1-getting-started)
2. [Configuration](#2-configuration)
3. [Web Interface](#3-web-interface)
4. [Managing Events](#4-managing-events)
5. [Source Types](#5-source-types)
6. [Live Preview](#6-live-preview)
7. [How It Works Internally](#7-how-it-works-internally)
8. [Schedule JSON Reference](#8-schedule-json-reference)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Getting Started

### 1.1 What is Scene Scheduler?

Scene Scheduler is an external automation tool for OBS Studio that runs your broadcast on a time-based schedule — like a television station's programming grid. You define what content plays at what time, and Scene Scheduler handles the transitions automatically, 24/7, without manual intervention.

**Key features:**
- Time-based automation with recurring events (daily/weekly)
- Visual calendar interface (FullCalendar) accessible from any browser on your network
- Real-time source preview via WebRTC and HLS
- Automatic OBS source staging for smooth, glitch-free transitions
- Optional default backup source for idle periods

### 1.2 Prerequisites

1. **OBS Studio** (version 28.0+) with the **WebSocket plugin v5** (included by default in OBS 28+)
2. **Operating System**: Linux (tested on Ubuntu 20.04+), Windows 10/11, or macOS
3. **Network**: OBS and Scene Scheduler must be reachable via network

### 1.3 Installation

**Linux:**
```bash
tar -xzf scenescheduler-linux-amd64.tar.gz
cd scenescheduler
chmod +x build/scenescheduler
```

**Windows:**
1. Extract `scenescheduler-windows-amd64.zip` to a folder (e.g., `C:\scenescheduler\`)
2. Open Command Prompt in that folder

### 1.4 Quick Start

**Step 1 — Configure OBS WebSocket:**
1. Open OBS Studio → **Tools** → **WebSocket Server Settings**
2. Enable "Enable WebSocket server"
3. Set a password (recommended) and note the port (default: 4455)

**Step 2 — Edit `config.json`:**
```json
{
  "obs": {
    "host": "localhost",
    "port": 4455,
    "password": "your-obs-password",
    "scheduleScene": "_SCHEDULER",
    "scheduleSceneAux": "_SCHEDULER_AUX"
  },
  "webServer": {
    "port": "8080",
    "user": "admin",
    "password": "your-web-password",
    "hlsPath": "hls"
  },
  "paths": {
    "logFile": "logs.txt",
    "schedule": "schedule.json"
  }
}
```

**Step 3 — Start Scene Scheduler:**
```bash
# Linux
./build/scenescheduler

# Windows
scenescheduler.exe
```

**Step 4 — Open the web interface:**
- Same machine: `http://localhost:8080`
- Other devices: `http://<server-ip>:8080`

If you configured `user` and `password`, the browser will prompt for HTTP Basic Authentication credentials.

---

## 2. Configuration

Scene Scheduler uses a single `config.json` file located in the same directory as the executable.

### 2.1 OBS Connection (`obs`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | `"localhost"` | OBS hostname or IP |
| `port` | integer | `4455` | OBS WebSocket port |
| `password` | string | `""` | OBS WebSocket password |
| `reconnectInterval` | integer | `15` | Seconds between reconnection attempts |
| **`scheduleScene`** | string | — | **Required.** Primary scene managed by the scheduler |
| **`scheduleSceneAux`** | string | — | **Required.** Auxiliary "staging" scene for preloading sources |
| `sourceNamePrefix` | string | `"_sched_"` | Prefix for sources created by the scheduler |

Both `scheduleScene` and `scheduleSceneAux` are created automatically in OBS if they don't exist.

### 2.2 Web Server (`webServer`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | string | `"8080"` | HTTP server port |
| `user` | string | `""` | HTTP Basic Auth username (empty = auth disabled) |
| `password` | string | `""` | HTTP Basic Auth password |
| `hlsPath` | string | `"hls"` | Directory for HLS preview files (must be relative) |
| `enableTls` | boolean | `false` | Enable HTTPS |
| `certFilePath` | string | `""` | TLS certificate path (required if TLS enabled) |
| `keyFilePath` | string | `""` | TLS private key path (required if TLS enabled) |

### 2.3 Media Source (`mediaSource`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `videoDeviceIdentifier` | string | `""` | Video capture device identifier |
| `audioDeviceIdentifier` | string | `""` | Audio capture device identifier |
| `quality` | string | `"low"` | Quality preset: `"low"`, `"medium"`, `"high"` |

These settings control the WebRTC live preview stream shown in the Monitor view.

### 2.4 Paths (`paths`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `logFile` | string | `""` | Log output file path |
| `schedule` | string | `"schedule.json"` | Schedule file path |

### 2.5 Scheduler (`scheduler`)

| Field | Type | Description |
|-------|------|-------------|
| `defaultSource` | object | Optional backup source for idle periods |
| `defaultSource.name` | string | OBS source name |
| `defaultSource.inputKind` | string | OBS input type (e.g., `"image_source"`) |
| `defaultSource.uri` | string | Content path or URL |
| `defaultSource.inputSettings` | object | Additional OBS input settings |
| `defaultSource.transform` | object | Position/scale/crop transform |

### 2.6 Validation

On startup, Scene Scheduler validates:
- ✅ `obs.scheduleScene` and `obs.scheduleSceneAux` are present (fatal if missing)
- ✅ `webServer.hlsPath` is a safe relative path (no `..` or absolute paths)
- ✅ TLS cert/key paths present when `enableTls` is true
- ⚠️ Warning if `obs.password` is empty
- ⚠️ Warning if `webServer.user` or `webServer.password` is empty

---

## 3. Web Interface

Scene Scheduler is a **single-page application** served at the root URL (`http://<host>:<port>/`). There is no separate `/editor.html` page.

### 3.1 Switching Views

The header contains a **View dropdown** with two options:
- **📺 Monitor** — Read-only calendar with live preview and activity log
- **📝 Editor** — Editable calendar for managing events

### 3.2 Connection Status Indicators

The header displays **three** independent connection indicators:

| Indicator | Green | Red |
|-----------|-------|-----|
| **Server** | WebSocket connected to backend | Disconnected (auto-reconnects every 5s) |
| **OBS** | Backend connected to OBS Studio | OBS not reachable |
| **Preview** | VirtualCam active, live preview available | No preview stream |

### 3.3 Monitor View

The Monitor view is designed for passive observation. It contains:

- **Left sidebar:**
  - **Live Preview** — WebRTC video feed from the server's camera/microphone (requires `mediaSource` config)
  - **Activity Log** — Real-time log of server events (connections, schedule loads, errors)

- **Main area:**
  - **Calendar** (read-only) — Shows all scheduled events as colored blocks on a timeline
  - **Current event highlighting** — The currently active event is highlighted in green (`#22c55e`)
  - Clicking an event in Monitor view opens a **preview popup** (read-only) showing the source URI, input kind, and a Preview button

### 3.4 Editor View

The Editor view provides full control over the schedule:

- **Calendar** (editable) — Click on a time slot to create a new event, or click an existing event to edit it
- **Drag & resize** — Events can be moved or resized directly on the calendar
- **Sync indicator** — The header shows "Synced" (green) or "Unsaved" (orange) to indicate if the current schedule matches the server

### 3.5 Editor Status

The Editor view shows a sync status indicator in the header:
- **Synced** — Your local calendar matches the server
- **Unsaved** — You have local changes not yet saved to the server

---

## 4. Managing Events

### 4.1 Creating an Event

1. In **Editor view**, click on an empty time slot in the calendar
2. The **Task Editor** modal opens with five tabs
3. Fill in the required fields (at minimum: Title, Start/End time)
4. Click **Save Changes**

### 4.2 Event Modal — Five Tabs

#### Tab 1: General

| Field | Description |
|-------|-------------|
| **Description** | Optional text description of the event |
| **Tags** | Space-separated tags for organization |
| **ClassNames** | CSS class names for custom styling |
| **Text Color** | Color picker for event text |
| **Background Color** | Color picker for event background |
| **Border Color** | Color picker for event border (for recurring events) |

#### Tab 2: Source

Defines the OBS source that will be created when this event triggers.

| Field | Description |
|-------|-------------|
| **Input Name** * | Technical name for the OBS source (e.g., `"YT_Chillhop"`) |
| **Input Kind** * | Source type dropdown (see [Section 5](#5-source-types)) |
| **URI** * | Content path or URL |
| **Settings (JSON)** | Additional OBS input settings as raw JSON |
| **Transform (JSON)** | Position, scale, crop as raw JSON |

#### Tab 3: Timing

| Field | Description |
|-------|-------------|
| **Start** * | Start date and time (datetime picker with seconds) |
| **End** * | End date and time |
| **Duration** | Automatically calculated (read-only) |
| **Recurring** | Toggle to enable weekly recurrence |

When **Recurring** is enabled:
- **From / Until** — Date range for the recurring series
- **Week Days** — Checkboxes for Mon–Sun
- Only the **time** portion of Start/End is used; the dates come from the recurrence range

#### Tab 4: Behavior

| Field | Description |
|-------|-------------|
| **Preload seconds** | How many seconds before the event to start staging the source (default: 0) |
| **On end action** | What happens when the event ends: `hide` (default), `none`, or `stop` |

#### Tab 5: Preview

- **Preview Source** button — Generates an HLS preview stream of the configured source
- Video player shows the preview inline
- Requires the `hls-generator` companion tool to be available

### 4.3 The Title Field

The **Title** field appears above the tabs, always visible. It serves as both the event display name on the calendar and a quick identifier.

### 4.4 Enabled/Disabled Toggle

Next to the title, a **toggle switch** controls whether the event is active. Disabled events remain in the schedule but are not executed by the scheduler.

### 4.5 Editing Events

Click any event on the Editor calendar to reopen the modal with all fields populated.

### 4.6 Deleting Events

The **Delete** button appears at the bottom-left of the modal. Deleting an event removes it immediately.

### 4.7 Drag and Resize

In the Editor calendar:
- **Drag** an event to move it to a different time
- **Resize** by dragging the bottom edge to change duration

### 4.8 Saving to Server

After making changes in the Editor, the schedule is automatically sent to the server via WebSocket (`commitSchedule` action). The server saves it to the `schedule.json` file.

---

## 5. Source Types

The **Input Kind** dropdown in the Source tab offers these OBS source types:

| Input Kind | Use Case | URI Example |
|------------|----------|-------------|
| `ffmpeg_source` | Local video/audio files, RTMP/RTSP/RTP/SRT streams | `/path/to/video.mp4` or `rtmp://server/stream` |
| `browser_source` | Web pages, HTML overlays, embedded video players | `https://www.youtube.com/embed/...` |
| `vlc_source` | VLC media playlists | `/path/to/playlist.m3u` |
| `ndi_source` | NDI network video streams | NDI source name |
| `image_source` | Static images | `/path/to/image.png` |

### 5.1 Input Settings (JSON)

The **Settings** field accepts raw JSON that is passed directly to OBS when creating the source. Common examples:

**Browser source with custom dimensions:**
```json
{
  "css": "body { background-color: rgba(0, 0, 0, 0); margin: 0px auto; overflow: hidden; }",
  "height": 1080,
  "width": 1920
}
```

### 5.2 Transform (JSON)

The **Transform** field accepts raw JSON for positioning the source in the scene:

```json
{
  "PositionX": 100,
  "PositionY": 50,
  "ScaleX": 0.5,
  "ScaleY": 0.5
}
```

---

## 6. Live Preview

### 6.1 Monitor View — Live Preview

The Monitor view's left sidebar shows a **WebRTC live preview** of the server's camera and microphone. This uses the WHEP protocol (`/whep/` endpoint) and requires:
- `mediaSource.videoDeviceIdentifier` and `audioDeviceIdentifier` configured in `config.json`
- The Preview status indicator to be green (VirtualCam active)

### 6.2 Source Preview (in Event Modal)

The **Preview** tab in the event modal lets you test a source before saving:
1. Configure the source in the **Source** tab (Input Kind + URI)
2. Switch to the **Preview** tab
3. Click **▶ Preview Source**
4. The server generates an HLS stream using the `hls-generator` companion tool
5. The video plays inline in the modal

The `hls-generator` binary must be placed in the same directory as the `scenescheduler` executable.

### 6.3 Monitor Preview Popup

In the Monitor view, clicking on a calendar event opens a **preview popup** showing:
- Source URI and Input Kind
- A **▶ Preview Source** button to generate an HLS preview
- An **Edit in Editor View** button to jump to the event editor

---

## 7. How It Works Internally

### 7.1 Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Production Server                      │
│                                                          │
│  ┌──────────────┐    WebSocket     ┌──────────────────┐ │
│  │  OBS Studio  │ ◄──────────────  │  Scene Scheduler │ │
│  │              │   (localhost)     │     (Backend)    │ │
│  └──────────────┘                  └────────┬─────────┘ │
│                                             │            │
│                                    HTTP (0.0.0.0:8080)   │
└─────────────────────────────────────────────┼────────────┘
                                              │
                        Network (LAN)         │
                    ┌─────────────────────────┼──────────┐
                    │                         │          │
               ┌────▼─────┐           ┌──────▼──┐  ┌───▼────┐
               │  Laptop  │           │ Tablet  │  │ Phone  │
               │ (Editor) │           │(Monitor)│  │(Monitor│
               └──────────┘           └─────────┘  └────────┘
```

### 7.2 Communication

The backend exposes four HTTP endpoints:

| Endpoint | Protocol | Purpose |
|----------|----------|---------|
| `/ws` | WebSocket | Real-time bidirectional communication |
| `/whep/` | HTTP (WebRTC) | Live camera/mic preview via WHEP |
| `/hls/` | HTTP (static) | HLS preview stream files |
| `/` | HTTP (static) | Frontend application |

### 7.3 WebSocket Protocol

Messages use the format `{ "action": "string", "payload": {} }`.

**Client → Server:**

| Action | Payload | Description |
|--------|---------|-------------|
| `getSchedule` | `{}` | Request current schedule |
| `commitSchedule` | Schedule v1.0 JSON | Save schedule changes |
| `getStatus` | `{}` | Request OBS and preview status |

**Server → Client:**

| Action | Payload | Description |
|--------|---------|-------------|
| `currentSchedule` | Schedule v1.0 JSON | Full schedule data |
| `log` | string | Activity log message |
| `obsConnected` | `{ obsVersion, timestamp }` | OBS connection established |
| `obsDisconnected` | `{ timestamp }` | OBS connection lost |
| `virtualCamStarted` | `{}` | Live preview stream available |
| `virtualCamStopped` | `{}` | Live preview stream stopped |
| `currentStatus` | `{ obsConnected, obsVersion, virtualCamActive }` | Initial status on connect |
| `previewReady` | `{ hlsUrl }` | Source preview HLS stream ready |
| `previewError` | `{ error }` | Source preview failed |
| `previewStopped` | `{ reason }` | Source preview auto-stopped |

### 7.4 Source Staging Process

When a scheduled event's time arrives:

1. **Stage** — Source is created in `scheduleSceneAux` (invisible to viewers), configured with all settings and transforms
2. **Activate** — Source is moved from the auxiliary scene to `scheduleScene`
3. **Scene Switch** — OBS transitions to `scheduleScene`
4. **Cleanup** — Temporary staging elements removed from `scheduleSceneAux`
5. **Monitor** — Source remains active until the event ends, then the configured `onEndAction` executes (`hide`, `stop`, or `none`)

### 7.5 Default Backup Source

When no event is scheduled (idle period), the `scheduler.defaultSource` (if configured) activates automatically, providing a standby image or content.

---

## 8. Schedule JSON Reference

The schedule file (`schedule.json`) follows the **Schedule v1.0** format:

```json
{
  "version": "1.0",
  "scheduleName": "Schedule",
  "schedule": [
    {
      "id": "evt-abc123",
      "title": "Morning News",
      "enabled": true,
      "general": {
        "description": "Daily morning broadcast",
        "tags": ["news", "morning"],
        "classNames": [],
        "textColor": "#ffffff",
        "backgroundColor": "#1f2fad",
        "borderColor": "#0fc233"
      },
      "source": {
        "name": "MorningStream",
        "inputKind": "ffmpeg_source",
        "uri": "rtmp://stream.example.com/live",
        "inputSettings": {},
        "transform": {}
      },
      "timing": {
        "start": "2025-01-01T07:00:00Z",
        "end": "2025-01-01T14:00:00Z",
        "isRecurring": true,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
          "startRecur": "2025-01-01",
          "endRecur": ""
        }
      },
      "behavior": {
        "onEndAction": "hide",
        "preloadSeconds": 0
      }
    }
  ]
}
```

### 8.1 Field Reference

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique event identifier (auto-generated) |
| `title` | Yes | Display name |
| `enabled` | Yes | Whether the event is active |
| `general.description` | No | Text description |
| `general.tags` | No | Array of tag strings |
| `general.classNames` | No | CSS classes for styling |
| `general.textColor` | No | Hex color for text |
| `general.backgroundColor` | No | Hex color for background |
| `general.borderColor` | No | Hex color for border |
| `source.name` | Yes | OBS source name |
| `source.inputKind` | Yes | OBS input type |
| `source.uri` | Yes | Content path or URL |
| `source.inputSettings` | No | Additional OBS settings (JSON object) |
| `source.transform` | No | Position/scale/crop (JSON object) |
| `timing.start` | Yes | ISO 8601 UTC start time |
| `timing.end` | Yes | ISO 8601 UTC end time |
| `timing.isRecurring` | Yes | Whether this is a recurring event |
| `timing.recurrence.daysOfWeek` | If recurring | Array of `"MON"` through `"SUN"` |
| `timing.recurrence.startRecur` | If recurring | Start date (`YYYY-MM-DD`) |
| `timing.recurrence.endRecur` | If recurring | End date (empty = indefinite) |
| `behavior.onEndAction` | No | `"hide"` (default), `"none"`, or `"stop"` |
| `behavior.preloadSeconds` | No | Seconds to preload before event start |

---

## 9. Troubleshooting

### 9.1 Connection Issues

| Problem | Solution |
|---------|----------|
| **Server indicator red** | Check that Scene Scheduler is running and the browser can reach the host/port |
| **OBS indicator red** | Verify OBS is running, WebSocket is enabled, and `obs.host`/`port`/`password` match |
| **Preview indicator red** | Enable VirtualCam in OBS (Tools → Start Virtual Camera) and configure `mediaSource` |

### 9.2 Configuration Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `obs.scheduleScene and obs.scheduleSceneAux are required` | Missing required fields | Add both fields to `config.json` |
| `webServer.hlsPath must be a relative path` | Absolute path used | Use a relative path like `"hls"` |
| `certFilePath and keyFilePath are required when TLS is enabled` | TLS enabled without certs | Provide cert/key paths or set `enableTls` to `false` |

### 9.3 Schedule Issues

| Problem | Solution |
|---------|----------|
| **Events not triggering** | Check `enabled` is `true` and the event time hasn't passed |
| **Recurring events not showing** | Verify `startRecur` date is in the past and `daysOfWeek` includes the current day |
| **Schedule not saving** | Check the Editor sync indicator; ensure WebSocket is connected |

### 9.4 Preview Issues

| Problem | Solution |
|---------|----------|
| **Preview button not working** | Ensure `hls-generator` binary is in the same directory as `scenescheduler` |
| **Live preview blank** | Check `mediaSource` config and VirtualCam status in OBS |

### 9.5 Restarting

Changes to `config.json` require a restart:

**Linux:**
```bash
pkill scenescheduler
./build/scenescheduler
```

**Windows:**
```cmd
REM Press Ctrl+C in the command prompt, then restart
scenescheduler.exe
```

Schedule changes (`schedule.json`) are applied in real-time via the web interface — no restart needed.

---

## License

Scene Scheduler is proprietary software. See LICENSE file for details.
