# Scene Scheduler Beta 0.2 User Manual

---

## 🚀 1. Quick Installation (For the Impatient)

This section gets you up and running in minutes. Follow these 4 essential steps:

### Step 1: OBS Preparation

**Enable WebSocket:**
- In OBS Studio, go to **Tools → WebSocket Server Settings**
- Check the **"Enable WebSocket server"** checkbox
- Note the port (default is **4455**)
- Set a **secure password** and write it down
- Click **"Apply"** then **"OK"**

**Create Required Scenes:**
- In OBS, create two new empty scenes:
  - Right-click in the Scenes panel
  - Select **"Add" → "Scene"**
  - Create a scene named **Schedule** (will be the visible main scene)
  - Create another scene named **Schedule_Temp** (temporary staging scene)
  - **Important:** These scenes must be completely empty at startup

### Step 2: Minimal Configuration

- Unzip the downloaded .zip file to a folder of your choice
- Open the **config.json** file with a text editor (Notepad, Notepad++, VS Code, etc.)
- Fill in these required fields in the "obs" section:

```json
"obs": {
  "host": "localhost",
  "port": 4455,
  "password": "your_obs_password",
  "scheduleScene": "Schedule",
  "scheduleSceneTmp": "Schedule_Temp"
}
```

- Save the config.json file

### Step 3: Execution

- Double-click **scenescheduler.exe** (Windows) or run `./scenescheduler` (Linux/Mac)
- A terminal window with logs will open. **Don't close it** - this window must remain open
- Wait for the message **"WebServer running on port 8080"**
- Open your web browser (Chrome, Firefox, Edge) and go to: **http://localhost:8080**

### Step 4: Web Interface and Workflow

**Understanding the Interface:**

The interface has two main views:

- **Monitor View** (read-only): For observing the system in real-time
  - Backend activity log
  - Live stream preview
  - Calendar with active server schedule

- **Editor View** (full editing): For modifying the schedule
  - Editable calendar
  - Actions menu (⋯) with schedule operations

**Status Indicators:**

At the top you'll see three status indicators:

- **Server** (WebSocket): Green = Connected to backend | Red = Disconnected
- **OBS**: Green = Backend connected to OBS | Red = OBS disconnected
- **Preview**: Green = Stream available | Orange = Connecting | Red = Unavailable

**Basic Editing Workflow:**

1. **Load:** Switch to Editor View, click the **⋯** menu and select **"Get from Server"** to load the current schedule
2. **Edit:**
   - Click on the calendar to create new events
   - Double-click existing events to edit them
   - Drag events to move them
3. **Save:** When finished, return to the **⋯** menu and select **"Commit to Server"**. Changes will be automatically applied in OBS

---

## 2. Introduction

### Welcome to Scene Scheduler!

Scene Scheduler is a powerful tool designed to fully automate your OBS Studio production. It allows you to plan ahead what content will be shown and when, creating a broadcast schedule similar to that of a professional television channel.

The system works with a very intuitive web calendar where you can add, move, and edit events visually. Once the schedule is saved, Scene Scheduler takes care of changing sources in OBS automatically, precisely, and without visual cuts, ensuring continuous 24/7 operation.

### Key Features

- **Total Automation:** Once configured, Scene Scheduler manages all scene changes without manual intervention
- **Dual Web Interface:** Monitor View (observation) and Editor View (full modification)
- **Triple Status System:** Independent indicators for Server, OBS, and Preview
- **Live Preview:** Ultra-low latency WebRTC streaming with WHEP protocol
- **Seamless Transitions:** 5-step staging system that ensures smooth transitions without visual artifacts
- **Recurring Events:** Schedule events that repeat daily, weekly, or on specific days
- **Hot-Reload:** Schedule changes are applied automatically without restarting
- **Automatic Reconnection:** Intelligent reconnection system with state synchronization
- **24/7 Operation:** Designed to run continuously without interruptions

### Who is this manual for?

This manual is aimed at Scene Scheduler end users. We'll guide you step by step, from initial setup to daily schedule management, without requiring technical programming knowledge. We'll cover:

- Installation and initial configuration
- Using the calendar to create and manage events
- Configuring different types of sources (videos, images, web pages)
- Troubleshooting common issues
- Best practices for efficient operation

---

## 3. Web Interface - Overview

The Scene Scheduler web interface provides two specialized views for different purposes.

### 3.1. Dual View System

**Monitor View (Read-Only)**

Purpose: Observe the current system state without modifying anything.

Components:
- **Activity Log:** Shows all backend events in real-time
  - Connections and disconnections
  - Program changes
  - Schedule reload
  - VirtualCam events
- **Live Preview:** WebRTC stream of what OBS is broadcasting
  - WHEP protocol for ultra-low latency
  - Playback controls (play/pause)
  - On-demand connection (only when playing)
- **Read-Only Calendar:** Visualization of the active server schedule
  - Cannot create or edit events
  - Shows current program highlighted
  - Clicking events opens read-only modal

**Editor View (Full Editing)**

Purpose: Safe workspace for modifying the schedule.

Components:
- **Editable Calendar:** Complete editing functionality
  - Create events: Click or click-and-drag
  - Modify events: Double-click to open editor
  - Move events: Drag to new position
  - Change duration: Drag borders
  - Delete events: Delete key or button in modal
- **Actions Menu (⋯):** Main schedule operations
  - New Schedule: Clear calendar
  - Load from File: Import schedule from local JSON
  - Save to File: Export current schedule to JSON
  - Get from Server: Load active server schedule
  - Commit to Server: Save changes to server
- **Status Bar:** Shows synchronization status
  - Green "Synced with server": No pending changes
  - Orange "X unsaved changes": Unsaved changes
  - Blue "Saving...": Operation in progress
  - Red: Error message

### 3.2. Triple Status Indicator System

At the top of the web interface you'll find three independent indicators showing connection status:

**Server Indicator (WebSocket)**

Shows the connection status between browser and backend:
- **Green:** Connected to backend server
- **Red:** Disconnected from backend server
- Tooltip: Shows connection status

When connection is lost, the system attempts to reconnect automatically every 5 seconds. Upon successful reconnection, all state (status and schedule) is re-synchronized to ensure updated information.

**OBS Indicator (Backend ↔ OBS)**

Shows the connection status between backend and OBS Studio:
- **Green:** Backend connected to OBS
- **Red:** Backend not connected to OBS
- Tooltip: Shows OBS version when connected

This indicator reflects whether the backend can communicate with OBS Studio through the obs-websocket protocol.

**Preview Indicator (VirtualCam Stream)**

Shows preview stream availability:
- **Green:** Stream available or actively connected
- **Orange:** WebRTC connection in progress
- **Red:** Stream unavailable (VirtualCam stopped)

Detailed states:
- **unavailable (Red):** VirtualCam stopped in OBS or stream unavailable (503)
- **available (Green):** VirtualCam active, stream available to play
- **connecting (Orange):** Establishing WebRTC connection
- **connected (Green):** WebRTC connected, stream actively playing

**Important note:** User pause/play actions don't change availability status. It only changes when the stream actually becomes unavailable (VirtualCam stops, network error, etc.).

### 3.3. Live Preview with WHEP

Scene Scheduler uses the WHEP protocol (WebRTC-HTTP Egress Protocol) for ultra-low latency video streaming.

**How it works:**

1. In OBS, click **"Start Virtual Camera"** (VirtualCam)
2. The backend captures this stream and prepares it for WebRTC distribution
3. In Monitor View, click the **Play** button on the player
4. The browser establishes WebRTC connection with the backend
5. The stream displays in the player with minimal latency

**Controls:**

- **Play:** Establishes WebRTC connection and starts playback
- **Pause:** Disconnects WebRTC session (maintains stream availability)
- **Volume:** Controls audio level

**Stream Behavior:**

- WebRTC connection is established **only when Play is pressed**
- When pausing, WebRTC session disconnects to free resources
- If stream remains available, you can play again immediately
- If VirtualCam stops in OBS, status changes to **unavailable** (red)
- Connection is maintained while stream is available and playing

**Error Handling:**

The system distinguishes between different states:
- **503 Service Unavailable:** Expected response when VirtualCam is not active (not logged as error)
- **Network errors:** Appropriate messages displayed
- **Remote stream ends:** Automatic disconnection

---

## 4. Detailed Installation and Configuration

To get Scene Scheduler running with all its features, follow these detailed steps.

### Step 1: System Requirements

Before installing, ensure you have:

- **Operating System:** Windows 10/11, macOS 10.15+, or Linux (Ubuntu 20.04+)
- **OBS Studio:** Version 28.0 or higher with WebSocket Plugin
- **Web Browser:** Chrome 90+, Firefox 88+, Edge 90+ or Safari 14+ (with WebRTC support)
- **RAM:** Minimum 4GB (8GB recommended)
- **Disk Space:** 100MB for application + space for logs

### Step 2: Unzip the Files

You'll receive a .zip file with the Scene Scheduler distribution. Unzip it to a permanent folder on your computer (avoid temporary or download folders). Inside you'll find:

**Essential Files:**
- `scenescheduler.exe` (Windows) or `scenescheduler` (Linux/Mac): The main program
- `config.json`: The main configuration file
- `schedule.json`: File where your calendar is saved (initially with examples)

**Generated Files:**
- `logs.txt`: Text file with logs (created automatically when running)
- Additional `.log` files may be created with date/time according to configuration

### Step 3: Configure the Connection (config.json)

Open the `config.json` file with a text editor. This file controls all aspects of Scene Scheduler. Let's review each section in detail:

#### 3.1. OBS Connection (Section "obs")

This is the most important section and must be configured correctly for Scene Scheduler to work.

Before starting:
- Open OBS Studio
- Go to **Tools → WebSocket Server Settings**
- Ensure **"Enable WebSocket server"** is checked
- Configure a port (default 4455) and a secure password
- Create the two required empty scenes in OBS

Configuration parameters:

```json
"obs": {
  "host": "localhost",              // Address of PC with OBS
  "port": 4455,                     // WebSocket port
  "password": "your_password",      // WebSocket password
  "reconnectInterval": 5,           // Seconds between retries
  "scheduleScene": "Schedule",      // Name of main scene
  "scheduleSceneTmp": "Schedule_Temp",  // Temporary scene
  "sourceNamePrefix": "SS_"         // Prefix for sources
}
```

Important notes:
- **host:** Use "localhost" if OBS is on the same PC. For remote control, use the IP of the PC with OBS
- **scheduleScene and scheduleSceneTmp:** Names must match EXACTLY with scenes in OBS
- **sourceNamePrefix:** All sources created by Scene Scheduler will have this prefix for identification

#### 3.2. Web Server (Section "webServer")

Configure web calendar interface access:

```json
"webServer": {
  "port": "8080",           // Port for web interface
  "user": "",               // User (empty = no authentication)
  "password": "",           // Password (empty = no authentication)
  "hlsPath": "hls",         // Directory for HLS previews
  "enableTls": false,       // HTTPS enabled/disabled
  "certFilePath": "",       // Path to SSL certificate
  "keyFilePath": ""         // Path to SSL key
}
```

Common configurations:
- **Local access without security:** Leave user and password empty
- **Protected access:** Set user and password to require authentication
- **HTTPS:** Configure `enableTls: true` and provide certificate files

Notes about hlsPath:
- **Must be a relative path** (e.g., "hls", "data/previews")
- Absolute paths not allowed (e.g., "/etc/hls") for security
- Directory traversal not permitted (e.g., "../hls")

#### 3.3. Scheduler (Section "scheduler")

Define what to show when no events are scheduled:

```json
"scheduler": {
  "defaultSource": {
    "name": "standby_image",
    "inputKind": "image_source",
    "uri": "C:/images/standby.png",
    "inputSettings": {
      "file": "C:/images/standby.png"
    },
    "transform": {
      "positionX": 0,
      "positionY": 0,
      "scaleX": 1.0,
      "scaleY": 1.0
    }
  }
}
```

Default source types:
- Static image: `inputKind: "image_source"`
- Looping video: `inputKind: "ffmpeg_source"`
- Web page: `inputKind: "browser_source"`

#### 3.4. Live Preview (Section "mediaSource")

Configure capture for preview:

```json
"mediaSource": {
  "videoDeviceIdentifier": "OBS Virtual Camera",
  "audioDeviceIdentifier": "default",
  "quality": "low"  // "low", "medium", or "high"
}
```

Step-by-step configuration:
- In OBS, click **"Start Virtual Camera"**
- Run `scene-scheduler -list-devices` to see available devices
- Copy exact device name to `videoDeviceIdentifier`

#### 3.5. File Paths (Section "paths")

Define important file locations:

```json
"paths": {
  "logFile": "./scene-scheduler.log",   // Log file
  "schedule": "./schedule.json"         // Schedule file
}
```

---

## 5. Schedule File (schedule.json) - Complete Format

The `schedule.json` file is the heart of Scene Scheduler. It contains all your programming and must follow a strict JSON format. Below is the complete structure with all available fields.

### 5.1. General File Structure

The complete file is wrapped in an object containing metadata and the events array:

```json
{
  "version": "1.0",
  "scheduleName": "My Streaming Schedule",
  "schedule": [
    // Array of events (programs)
  ]
}
```

Main fields:
- **version:** Format version (currently "1.0")
- **scheduleName:** Descriptive name of your schedule
- **schedule:** Array containing all scheduled events

### 5.2. Event Structure

Each element in the schedule array is an object with the following structure:

```json
{
  "id": "evt-001",
  "title": "Morning Program",
  "enabled": true,
  "general": { /* ... */ },
  "source": { /* ... */ },
  "timing": { /* ... */ },
  "behavior": { /* ... */ }
}
```

### 5.3. Main Event Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | Yes | Unique event identifier (e.g., "evt-001") |
| title | string | Yes | Descriptive title that appears in calendar |
| enabled | boolean | Yes | If true event will execute, if false it's ignored |
| general | object | No | Visual configuration and metadata |
| source | object | Yes | Defines what content to show in OBS |
| timing | object | Yes | Defines when the event executes |
| behavior | object | No | Automatic behaviors |

### 5.4. Section "general" - Appearance and Metadata

```json
"general": {
  "description": "Morning news with the production team",
  "tags": ["news", "daily", "priority"],
  "classNames": ["high-priority", "news-segment"],
  "textColor": "#FFFFFF",
  "backgroundColor": "#2196F3",
  "borderColor": "#1976D2"
}
```

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| description | string | Descriptive text of event | "Interview segment" |
| tags | array[string] | Tags for categorization | ["interview", "live"] |
| classNames | array[string] | Custom CSS classes | ["premium-content"] |
| textColor | string | Hexadecimal text color | "#FFFFFF" |
| backgroundColor | string | Hexadecimal background color | "#FF5722" |
| borderColor | string | Hexadecimal border color | "#E64A19" |

### 5.5. Section "source" - Content Configuration

This section defines exactly what content OBS will show:

```json
"source": {
  "name": "morning_news_feed",
  "inputKind": "ffmpeg_source",
  "uri": "C:/Videos/morning_news.mp4",
  "inputSettings": {
    "local_file": true,
    "looping": false,
    "restart_on_activate": true
  },
  "transform": {
    "positionX": 0,
    "positionY": 0,
    "scaleX": 1.0,
    "scaleY": 1.0
  }
}
```

Source fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Unique technical source name (no spaces) |
| inputKind | string | Yes | OBS source type (see available types) |
| uri | string | Yes* | Content location (path or URL) |
| inputSettings | object | No | Type-specific source configuration |
| transform | object | No | Position and transformation in scene |

Available inputKind types:
- `ffmpeg_source`: Local videos and streams
- `browser_source`: Web pages and HTML content
- `image_source`: Static images
- `vlc_source`: Playback with VLC

### 5.6. Section "timing" - Temporal Scheduling

**IMPORTANT:** The start and end fields must use ISO 8601 format with timezone (Z for UTC):

```json
"timing": {
  "start": "2024-03-15T09:00:00Z",
  "end": "2024-03-15T10:30:00Z",
  "isRecurring": false,
  "recurrence": {
    "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
    "startRecur": "2024-01-01",
    "endRecur": "2024-12-31"
  }
}
```

Timing fields:

| Field | Type | Format | Description |
|-------|------|--------|-------------|
| start | string | ISO 8601 | Start date/time: YYYY-MM-DDTHH:MM:SSZ |
| end | string | ISO 8601 | End date/time: YYYY-MM-DDTHH:MM:SSZ |
| isRecurring | boolean | - | If true, event repeats |
| recurrence | object | - | Recurrence configuration (if applicable) |

Recurrence fields:

| Field | Type | Format | Description |
|-------|------|--------|-------------|
| daysOfWeek | array | - | Repeat days: ["MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"] |
| startRecur | string | YYYY-MM-DD | First date of recurring series |
| endRecur | string | YYYY-MM-DD | Last date of recurring series |

Note on recurring events: For repeating events, the start and end fields define only the time of day (the time part), while repeat dates are controlled with startRecur and endRecur.

### 5.7. Section "behavior" - Automatic Behavior

```json
"behavior": {
  "onEndAction": "hide",
  "preloadSeconds": 30
}
```

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| onEndAction | string | "hide", "stop", "none" | Action when event ends |
| preloadSeconds | number | 0-300 | Seconds to preload before start |

### 5.8. Complete schedule.json Example

```json
{
  "version": "1.0",
  "scheduleName": "Web TV Channel Programming",
  "schedule": [
    {
      "id": "morning-news-001",
      "title": "Morning News",
      "enabled": true,
      "general": {
        "description": "Morning news summary with latest updates",
        "tags": ["news", "informative", "daily"],
        "classNames": ["news-program", "high-priority"],
        "textColor": "#FFFFFF",
        "backgroundColor": "#1E88E5",
        "borderColor": "#1565C0"
      },
      "source": {
        "name": "morning_news_source",
        "inputKind": "browser_source",
        "uri": "https://news.example.com/live",
        "inputSettings": {
          "url": "https://news.example.com/live",
          "width": 1920,
          "height": 1080,
          "fps": 30,
          "css": "body { overflow: hidden; }"
        },
        "transform": {
          "positionX": 0,
          "positionY": 0,
          "scaleX": 1.0,
          "scaleY": 1.0
        }
      },
      "timing": {
        "start": "2024-03-15T09:00:00Z",
        "end": "2024-03-15T10:00:00Z",
        "isRecurring": true,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
          "startRecur": "2024-03-01",
          "endRecur": "2024-12-31"
        }
      },
      "behavior": {
        "onEndAction": "hide",
        "preloadSeconds": 30
      }
    },
    {
      "id": "lunch-break-002",
      "title": "Break Screen",
      "enabled": true,
      "general": {
        "description": "Static image during lunch hours",
        "tags": ["break", "image", "daily"],
        "classNames": ["break-screen"],
        "textColor": "#000000",
        "backgroundColor": "#4CAF50",
        "borderColor": "#388E3C"
      },
      "source": {
        "name": "lunch_break_image",
        "inputKind": "image_source",
        "uri": "C:/Images/lunch_break.png",
        "inputSettings": {
          "file": "C:/Images/lunch_break.png",
          "unload": false
        },
        "transform": {
          "positionX": 0,
          "positionY": 0,
          "scaleX": 1.0,
          "scaleY": 1.0
        }
      },
      "timing": {
        "start": "2024-03-15T12:00:00Z",
        "end": "2024-03-15T13:00:00Z",
        "isRecurring": true,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI"],
          "startRecur": "2024-03-01",
          "endRecur": "2024-12-31"
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

### 5.9. Important Format Notes

- **ISO 8601 Date/Time Format:**
  - Always use format YYYY-MM-DDTHH:MM:SSZ
  - The T separates date and time
  - The Z at the end indicates UTC (universal time)
  - Example: 2024-03-15T09:00:00Z = March 15, 2024, 9:00 AM UTC

- **Unique IDs:** Each event must have a unique id throughout the file

- **JSON Validation:** File must be valid JSON (watch for commas and quotes)

- **Optional fields:** Only id, title, enabled, source, and timing are mandatory

- **Disabled events:** Events with `enabled: false` remain in the file but don't execute

---

## 6. Managing Your Schedule

### 6.1. Fundamental Concepts

Before creating events, it's important to understand these concepts:

- **Event/Program:** A unit of content with start and end time
- **Source:** The actual content OBS will show (video, image, web)
- **Scene:** The container in OBS where sources are placed
- **Recurrence:** Events that repeat automatically according to a pattern
- **Server Schedule:** The official schedule the backend is executing
- **Working Schedule:** Local copy in the Editor that may diverge from server

### 6.2. The Actions Menu (⋯)

Located in the top right corner of the calendar in Editor View, it contains the main actions:

#### Menu Options:

**1. New Schedule**
- Completely clears the calendar
- **Warning:** This action cannot be undone
- Useful for starting a schedule from scratch

**2. Load from File**
- Loads a schedule from a .json file on your PC
- Allows maintaining multiple schedules and switching between them
- Doesn't affect active server schedule until "Commit"

**3. Save to File**
- Saves current schedule to a .json file
- Useful for backups
- Includes all events and their configurations

**4. Get from Server ⭐ (Main Action)**
- Loads active schedule from server
- Synchronizes your calendar with what Scene Scheduler is using
- Always use this when starting to be synchronized
- If there are unsaved changes, will ask for confirmation

**5. Commit to Server ⭐ (Main Action)**
- Saves all changes to the server
- Changes are applied immediately in OBS
- Scene Scheduler automatically reloads the new schedule
- Updates status to "Synced with server"

### 6.3. Creating and Modifying Events

#### Create a New Event:

**Method 1: Simple Click**
- Click on any empty space in the calendar
- Creation modal opens with selected time
- Event will have default 1-hour duration

**Method 2: Click and Drag**
- Click and hold at the start time
- Drag to desired end time
- Release to create event with exact duration

#### Modify Existing Events:

**Edit Details:**
- Double-click on event to open full editor
- Modify any parameter and save changes

**Move in Time:**
- Click and drag event to new position
- Event maintains its original duration

**Change Duration:**
- Position cursor on event's bottom edge
- Drag up or down to adjust duration

**Delete Events:**
- Select event by clicking on it
- Press Delete key
- Or open editor and use "Delete" button

### 6.4. The Edit Modal: Detailed Configuration

The edit modal is where you configure all aspects of an event. It's organized into five tabs:

#### "General" Tab - Basic Information

**1. Title (Required)**
- Descriptive event name
- Shows in calendar and logs
- Examples: "Morning News", "Promotional Video", "Technical Break"

**2. Enabled**
- Checkbox to enable/disable event
- Disabled events remain in calendar but don't execute
- Useful for temporary scheduling or testing

**3. Description**
- Optional descriptive text
- Internal notes about the event
- Doesn't affect operation, only informative

**4. Tags**
- Tags separated by spaces
- Facilitate search and categorization
- Examples: "news", "advertising", "educational"

**5. ClassNames**
- Custom CSS classes for advanced styling
- For advanced users who want to customize appearance

**6. Colors**
- **Text Color:** Text color in calendar
- **Background Color:** Event background color
- **Border Color:** Border color (useful for recurring events)
- Use color selector or enter hexadecimal codes

#### "Source" Tab - Content Configuration

This is the most important tab, defining what content OBS will show.

**1. Input Name (Required)**
- Unique technical source name in OBS
- Don't use spaces or special characters
- Scene Scheduler will automatically add configured prefix
- Examples: "video_intro", "image_break", "web_news"

**2. Input Kind (Required)**
- Type of OBS source to create
- Common options:
  - `ffmpeg_source`: Videos and media streams
  - `image_source`: Static images
  - `browser_source`: Web pages and HTML
  - `vlc_source`: Videos with VLC (if installed)

**3. URI (Required depending on type)**
- Content location
- For local files: Full path (e.g., C:/videos/intro.mp4)
- For web content: Full URL (e.g., https://example.com)
- For images: Path to image file

**4. Input Settings (JSON)**

Type-specific source configuration. Examples by type:

For ffmpeg_source (videos):
```json
{
  "local_file": true,
  "is_local_file": true,
  "looping": true,
  "restart_on_activate": true,
  "clear_on_media_end": false
}
```

For browser_source (web):
```json
{
  "url": "https://example.com",
  "width": 1920,
  "height": 1080,
  "fps": 30,
  "css": "body { background-color: transparent; }"
}
```

For image_source:
```json
{
  "file": "C:/images/logo.png",
  "unload": false
}
```

**5. Transform (JSON)**

Source position and transformation in scene:

```json
{
  "positionX": 0,           // Horizontal position (pixels)
  "positionY": 0,           // Vertical position (pixels)
  "scaleX": 1.0,            // Horizontal scale (1.0 = 100%)
  "scaleY": 1.0,            // Vertical scale
  "rotation": 0,            // Rotation in degrees
  "cropTop": 0,             // Top crop (pixels)
  "cropBottom": 0,          // Bottom crop
  "cropLeft": 0,            // Left crop
  "cropRight": 0            // Right crop
}
```

#### "Timing" Tab - Temporal Scheduling

Define when and how the event is scheduled.

**For Single Events:**

**1. Start Date/Time**
- Exact start date and time
- Use date/time selector or type directly
- Format: YYYY-MM-DD HH:MM:SS

**2. End Date/Time**
- Exact end date and time
- Must be after start time
- Defines total event duration

**For Recurring Events:**

**1. Recurring (Checkbox)**
- Activates recurrence mode
- Changes behavior of date fields

**2. Recurrence Pattern**
- **Days of Week:** Select days it repeats
  - Monday to Sunday available
  - Can select multiple days
- **Time:** For recurring events, only Start/End time is used
- **Date Range:**
  - **Start Recur:** First date of series
  - **End Recur:** Last date of series

Recurrence Examples:
- **Daily at 9 AM:** All days checked, Start: 09:00, End: 10:00
- **Monday to Friday:** Only weekdays checked
- **Weekends:** Only Saturday and Sunday checked

#### "Behavior" Tab - Advanced Behavior

**1. Preload Seconds**
- Seconds in advance to prepare source
- Useful for heavy videos or network streams
- Value 0 = load right at change moment

**2. On End Action**
- What to do when event ends:
  - **hide:** Hide source (default)
  - **stop:** Stop and release resources
  - **none:** Do nothing (keep visible)

#### "Preview" Tab - Source Preview System

The Preview tab allows you to test and verify your source content before scheduling it. This helps ensure your sources work correctly and display as expected.

**How Source Preview Works:**

All source types (videos, images, web pages, streams) use the same preview system:

1. **Click "Preview Source" button**
   - A loading spinner appears with a message
   - For web pages: "Loading browser engine (5-10 seconds)..."
   - For other sources: "Generating preview..."

2. **Backend generates HLS stream**
   - The system creates a temporary video stream of your source
   - For web pages, it renders the page using a browser engine
   - For videos/streams, it transcodes to web-compatible format

3. **Video player loads automatically**
   - Once ready, the preview plays automatically in the modal
   - You can see exactly how the source will appear in OBS
   - Video controls allow pause/play/seek

4. **Automatic timeout after 30 seconds**
   - Previews stop automatically to save resources
   - A blue notification appears: "Preview automatically stopped after 30 seconds"
   - Click "Preview Source" again if you need more time

5. **Manual stop**
   - Click "Stop Preview" button to stop early
   - Changing tabs or closing the modal also stops the preview

**Preview Features:**

- **All source types supported:** Videos, images, web pages, RTSP streams, RTMP streams, etc.
- **Real-time rendering:** See exactly what will appear in OBS
- **Resource efficient:** Automatic timeout prevents resource leaks
- **Loading feedback:** Clear visual indicators during generation
- **Error handling:** Friendly error messages if source is invalid

**Preview Tips:**

- **Web pages (browser_source):** Allow 5-10 seconds for browser engine initialization
- **Network streams:** Preview helps verify the stream URL is working
- **Local files:** Verify the file path is correct and file plays properly
- **Resolution testing:** Check that transformations display correctly

**Common Preview Messages:**

- **Loading spinner:** Preview is being generated (wait a few seconds)
- **Blue info message:** Preview stopped automatically after 30 seconds (normal behavior)
- **Orange warning message:** Error occurred (check your URI or settings)
- **HLS playback error:** Network issue or invalid source format

**Important Notes:**

- Preview uses backend resources - don't leave multiple previews running
- The 30-second timeout ensures system resources are freed automatically
- Preview quality may differ slightly from final OBS output
- Some sources (especially web pages) may take longer to initialize

### 6.5. Best Practices for Scheduling

#### Efficient Organization:

- **Use descriptive names:** Facilitates quick identification
- **Color coding:** Assign colors by category (e.g., blue for news, green for advertising)
- **Consistent tags:** Create a tag system and use it consistently
- **Document with descriptions:** Add important notes in description field

#### Avoiding Problems:

- **Don't overlap events:** Scene Scheduler will execute the most recent
- **Verify file paths:** Ensure all files exist
- **Test before broadcasting:** Use disabled events to test
- **Regular backup:** Save schedule copies frequently

#### Resource Optimization:

- **Reuse sources:** Use same Input Name for repeating content
- **Strategic preload:** Configure preload only where necessary
- **Periodic cleanup:** Remove old events you no longer need

---

## 7. How the Switching System Works

### 7.1. The Safe Switching Process

Scene Scheduler uses a sophisticated "staging" system to ensure transitions without visual artifacts. This 5-step process ensures your audience never sees cuts, black screens, or errors during transitions.

#### The 5 Switching Steps:

**Step 1: STAGING (Preparation)**
- New source is created in temporary scene (Schedule_Temp)
- Fully configured but remains hidden
- All transformations applied (position, scale, etc.)
- If fails: Process stops without affecting current broadcast

**Step 2: PROMOTION**
- Prepared element is duplicated to main scene (Schedule)
- Still remains hidden in main scene
- Verification that duplication was successful
- If fails: Complete rollback executed

**Step 3: ACTIVATION**
- New element becomes visible in main scene
- This is the exact moment of change for the audience
- Change is instantaneous and seamless
- If fails: Rollback and previous content maintained

**Step 4: CLEANUP (Staging Cleanup)**
- Temporary element removed from Schedule_Temp
- Unnecessary resources freed
- Temporary scene ready for next change

**Step 5: RETIREMENT (Previous Removal)**
- Previous program hidden in main scene
- Completely removed after hiding
- All previous content resources freed

### 7.2. Staging System Advantages

**1. Seamless Transitions**
- No black frames between transitions
- No flickers or visual artifacts
- Audience sees clean, instantaneous change

**2. Safety and Rollback**
- If something fails, current content continues
- Each step validates before continuing
- Automatic rollback system in case of error

**3. Early Preparation**
- Heavy sources load before the change
- Videos and streams have time to buffer
- Reduces system load at change moment

### 7.3. Change Logs and Diagnostics

The backend terminal shows detailed information for each change:

**Information Messages (Debug):**
- Creating source in TEMP scene: Staging start
- Duplicating to MAIN scene: Successful promotion
- Activating in MAIN scene: Visible change
- Cleanup completed: Process finished

**Warning Messages:**
- Source already exists: Reusing existing source
- Transform partially applied: Some parameters not applied
- Cleanup skipped: Elements not found for cleanup

**Error Messages:**
- Failed to create source: Couldn't create source
- Duplication failed: Error promoting to main scene
- Activation failed - rollback initiated: Change aborted

---

## 8. Common Use Cases

### 8.1. Online TV/Radio Broadcasting

Typical configuration:

```json
{
  "title": "Morning Show",
  "source": {
    "inputKind": "ffmpeg_source",
    "uri": "rtmp://server/live/stream"
  },
  "timing": {
    "isRecurring": true,
    "recurrence": {
      "daysOfWeek": ["MON","TUE","WED","THU","FRI"],
      "startRecur": "2024-01-01",
      "endRecur": "2024-12-31"
    }
  }
}
```

### 8.2. Information Displays

For lobbies, waiting rooms, stores:

```json
{
  "title": "Daily Information",
  "source": {
    "inputKind": "browser_source",
    "uri": "https://yourcompany.com/info-screen",
    "inputSettings": {
      "width": 1920,
      "height": 1080,
      "fps": 30
    }
  },
  "timing": {
    "start": "08:00:00",
    "end": "20:00:00",
    "isRecurring": true
  }
}
```

### 8.3. Gaming/Event Streaming

For scheduled tournaments and events:

```json
{
  "title": "CS:GO Tournament - Semifinals",
  "source": {
    "inputKind": "game_capture",
    "inputSettings": {
      "capture_mode": "window",
      "window": "Counter-Strike: Global Offensive"
    }
  },
  "timing": {
    "start": "2024-03-15T19:00:00",
    "end": "2024-03-15T23:00:00"
  }
}
```

### 8.4. Educational Content

Scheduled classes and tutorials:

```json
{
  "title": "Math Class - Algebra",
  "source": {
    "inputKind": "ffmpeg_source",
    "uri": "C:/Classes/algebra_lesson_5.mp4",
    "inputSettings": {
      "local_file": true,
      "looping": false,
      "restart_on_activate": true
    }
  }
}
```

---

## 9. Appendix and Troubleshooting

### A.1. Complete config.json Reference

This section details all available options in the config.json file, grouped by section.

#### Section "obs" - OBS Connection

| Key | Description | Required | Default Value | Type |
|-----|-------------|----------|---------------|------|
| host | Address of PC running OBS | No | "localhost" | string |
| port | OBS WebSocket server port | No | 4455 | integer |
| password | WebSocket password. Empty = no auth | No | "" | string |
| reconnectInterval | Seconds between reconnection attempts | No | 5 | integer |
| scheduleScene | Name of main visible scene | Yes | N/A | string |
| scheduleSceneTmp | Name of temporary staging scene | Yes | N/A | string |
| sourceNamePrefix | Prefix to identify managed sources | No | "SS_" | string |

Important notes:
- scheduleScene and scheduleSceneTmp names must match EXACTLY with OBS scenes
- sourceNamePrefix is used to identify and automatically clean orphaned sources

#### Section "webServer" - Web Server

| Key | Description | Required | Default Value | Type |
|-----|-------------|----------|---------------|------|
| port | Port for web interface | No | "8080" | string |
| user | User for basic authentication | No | "" | string |
| password | Password for basic authentication | No | "" | string |
| hlsPath | Directory for HLS previews (relative) | No | "hls" | string |
| enableTls | Enable HTTPS | No | false | boolean |
| certFilePath | Path to SSL certificate | Conditional* | "" | string |
| keyFilePath | Path to SSL private key | Conditional* | "" | string |

*Required if enableTls is true

Security configurations:
- No protection: Leave user and password empty (local use only)
- Basic authentication: Set user and password
- HTTPS: Configure enableTls: true with valid certificates

hlsPath restrictions:
- Only relative paths to execution directory allowed
- Absolute paths not accepted (e.g., "/var/hls")
- Directory traversal not permitted (e.g., "../data")

#### Section "scheduler" - Scheduler

| Key | Description | Required | Default Value | Type |
|-----|-------------|----------|---------------|------|
| defaultSource | Source to show when no events | No | null | object |

defaultSource structure:

```json
{
  "name": "string",           // Source name
  "inputKind": "string",      // Type (image_source, ffmpeg_source, etc.)
  "uri": "string",            // Path or URL of content
  "inputSettings": {},        // Type-specific configuration
  "transform": {}             // Position and transformation
}
```

#### Section "mediaSource" - Preview

| Key | Description | Required | Default Value | Type |
|-----|-------------|----------|---------------|------|
| videoDeviceIdentifier | Video device name | No | "" | string |
| audioDeviceIdentifier | Audio device name | No | "default" | string |
| quality | Encoding quality | No | "low" | string |

Quality values: "low", "medium", "high"

#### Section "paths" - System Paths

| Key | Description | Required | Default Value | Type |
|-----|-------------|----------|---------------|------|
| logFile | Log file | No | "./scene-scheduler.log" | string |
| schedule | Schedule file | No | "./schedule.json" | string |

### A.2. Command Line Tool

Scene Scheduler includes useful command line tools:

#### List Devices (-list-devices)

To find exact device identifiers:

Windows:
```
scene-scheduler.exe -list-devices
```

Linux/Mac:
```
./scene-scheduler -list-devices
```

Example output:
```
----------- Available Media Devices -----------
INFO: Use the 'Friendly Name' or 'DeviceID' for your config.

Device #0:
  - Kind          : Video Input
  - Friendly Name : OBS Virtual Camera
  - DeviceID      : v4l2:/dev/video6

Device #1:
  - Kind          : Audio Input
  - Friendly Name : Monitor of Built-in Audio Analog Stereo
  - DeviceID      : alsa:pulse_

----------------------------------------------
```

Copy the exact "Friendly Name" or "DeviceID" to your config.json.

#### Validate Configuration (-validate)

Verify your configuration is valid:

```
./scene-scheduler -validate
```

#### Debug Mode (-debug)

Start with detailed logging for diagnostics:

```
./scene-scheduler -debug
```

### A.3. Complete config.json Example

Here's a fully functional example with all sections:

```json
{
  "scheduler": {
    "defaultSource": {
      "name": "standby_screen",
      "inputKind": "image_source",
      "uri": "C:/Scene-Scheduler/assets/standby.png",
      "inputSettings": {
        "file": "C:/Scene-Scheduler/assets/standby.png",
        "unload": false
      },
      "transform": {
        "positionX": 0,
        "positionY": 0,
        "scaleX": 1.0,
        "scaleY": 1.0
      }
    }
  },
  "mediaSource": {
    "videoDeviceIdentifier": "OBS Virtual Camera",
    "audioDeviceIdentifier": "default",
    "quality": "medium"
  },
  "webServer": {
    "port": "8080",
    "user": "admin",
    "password": "secure_password_123",
    "hlsPath": "hls",
    "enableTls": false,
    "certFilePath": "",
    "keyFilePath": ""
  },
  "obs": {
    "host": "localhost",
    "port": 4455,
    "password": "obs_websocket_password",
    "reconnectInterval": 5,
    "scheduleScene": "Schedule",
    "scheduleSceneTmp": "Schedule_Temp",
    "sourceNamePrefix": "SS_"
  },
  "paths": {
    "logFile": "./scene-scheduler.log",
    "schedule": "./schedule.json"
  }
}
```

### A.4. Common Troubleshooting

#### Startup Problems

**Application closes immediately:**
- **Cause:** Error in config.json
- **Solution:**
  - Verify JSON syntax (commas, quotes, braces)
  - Ensure scheduleScene and scheduleSceneTmp are defined
  - Run with -validate to see specific errors

**Error "Cannot parse config file":**
- **Cause:** Malformed JSON
- **Solution:** Use online JSON validator or editor with syntax highlighting

**Message "Scene Scheduler has expired":**
- **Cause:** Beta version expired
- **Solution:** Contact developer for updated version

#### OBS Connection Problems

**"Failed to connect to OBS":**
- **Causes and solutions:**
  - OBS not running → Start OBS first
  - WebSocket not enabled → Tools > WebSocket Server Settings
  - Incorrect port → Verify it matches OBS
  - Incorrect password → Check password on both sides
  - Firewall blocking → Add exception for Scene Scheduler

**"Scene not found":**
- **Cause:** Scenes don't exist in OBS
- **Solution:** Create scenes exactly as in config.json

**Intermittent connection:**
- **Cause:** Unstable network or OBS overloaded
- **Solution:** Increase reconnectInterval to 10-15 seconds

#### Web Interface Problems

**Can't access calendar:**
- **Verifications:**
  - Terminal shows "WebServer running on port 8080"
  - Using correct URL: http://localhost:[port]
  - Firewall not blocking port
  - If authentication enabled, using correct credentials

**Calendar doesn't load:**
- **Cause:** Problems with embedded web server
- **Solution:** Restart Scene Scheduler and verify port isn't in use

**WebSocket constantly disconnecting:**
- **Causes:**
  - Proxy or VPN interfering
  - Browser extensions blocking WebSockets
  - Inactivity timeout
- **Solution:** Try incognito mode or different browser

#### Event Problems

**Events don't execute:**
- **Verifications:**
  - Event is enabled (enabled: true)
  - Date/time is correct
  - No overlapping events
  - Source file/URL exists

**Error creating source:**
- **Common causes:**
  - Unsupported source type
  - File not found
  - Inaccessible URL
  - Invalid JSON settings

**Videos not playing:**
- **Solution:**
  - Verify file exists and isn't corrupt
  - Use absolute paths, not relative
  - For ffmpeg_source, add: "local_file": true
  - Install necessary system codecs

#### Performance Problems

**High CPU usage:**
- **Causes:**
  - Too many active browser_source events
  - Very high resolution videos
  - Complex transforms
- **Solutions:**
  - Reduce preview quality
  - Optimize videos before use
  - Close unnecessary calendar tabs

**Constantly increasing memory:**
- **Cause:** Sources not releasing properly
- **Solution:**
  - Restart Scene Scheduler daily
  - Use onEndAction: "stop" for heavy videos

#### Source Preview Problems

**Preview takes too long to load:**
- **For web pages (browser_source):**
  - Normal: 5-10 seconds for browser engine initialization
  - Solution: Be patient, message indicates expected wait time
- **For video files:**
  - Check file exists and path is correct
  - Verify file isn't corrupted
  - Try with smaller test file first
- **For network streams:**
  - Check stream URL is accessible
  - Verify network connection
  - Some streams may require authentication

**Preview shows "Preview automatically stopped after 30 seconds":**
- **Status:** This is normal behavior, not an error
- **Reason:** Automatic timeout prevents resource accumulation
- **Solution:** Click "Preview Source" again if you need to see more

**Preview shows orange warning "HLS playback error":**
- **Causes:**
  - Invalid source URI or file path
  - Unsupported video format
  - Network stream not accessible
  - Insufficient backend resources
- **Solutions:**
  - Verify URI/path is correct
  - Check file format is supported (MP4, MKV, etc.)
  - Test stream URL in VLC or similar player
  - Check backend logs for specific error

**Preview shows "levelEmptyError - No Segments found":**
- **Cause:** Trying to load preview before segments are ready
- **Status:** Should be resolved automatically in v1.6+
- **Solution:** Wait a few more seconds and it will load

**Preview button doesn't respond:**
- **Verifications:**
  - Server status indicator is green (connected)
  - Backend terminal shows no errors
  - Try refreshing browser page (Ctrl+Shift+R)
- **Solution:** Check that hls-generator binary is in same directory as scenescheduler

**Multiple previews causing system slowdown:**
- **Cause:** Each preview uses backend resources
- **Solution:**
  - Stop current preview before starting a new one
  - System automatically stops previews after 30 seconds
  - Close modal when finished previewing

### A.5. Common Error Messages and Solutions

| Error Message | Meaning | Solution |
|---------------|---------|----------|
| Config file not found | config.json doesn't exist | Create or restore file |
| Invalid JSON in config | Incorrect JSON syntax | Validate JSON |
| Schedule file not found | schedule.json doesn't exist | Will be created automatically |
| OBS connection refused | OBS rejects connection | Verify port and password |
| Scene does not exist | Scene not found in OBS | Create required scenes |
| Source creation failed | Couldn't create source | Verify type and parameters |
| WebSocket upgrade failed | WS handshake error | Check network configuration |
| Permission denied | No file permissions | Run as administrator |
| Port already in use | Port occupied | Change port or close other application |

---

## 10. Best Practices and Recommendations

### 10.1. Initial Setup

- **Plan your structure:** Before starting, design your schedule on paper
- **Test locally:** Configure and test everything locally before production
- **Document your configuration:** Keep notes of your specific setup
- **Configuration backup:** Save copies of config.json and schedule.json

### 10.2. Daily Operation

- **Morning sync:** Always use "Get from Server" when starting the day
- **Frequent saving:** Do "Commit to Server" after important changes
- **Regular monitoring:** Check Monitor View periodically
- **Logs for diagnostics:** Review logs if something doesn't work as expected

### 10.3. Maintenance

- **Weekly cleanup:** Remove old events from calendar
- **Content updates:** Verify all referenced files exist
- **Scheduled restart:** Consider restarting Scene Scheduler weekly
- **Regular backups:** Export your schedule to file regularly

### 10.4. Security

- **Strong passwords:** Use secure passwords for WebSocket and web
- **Limited access:** In production, use web server authentication
- **Secure network:** For remote access, consider using VPN
- **File permissions:** Limit who can modify config.json

---

## 11. Glossary of Terms

- **Backend:** The server part of Scene Scheduler that manages logic
- **Commit:** Save changes to server for application
- **Editor View:** Edit view with editable calendar and actions menu
- **EventBus:** Internal system for module communication
- **Frontend:** The web calendar interface
- **Hot-reload:** Automatic reload without restarting application
- **Input/Source:** Content source in OBS (video, image, web)
- **Modal:** Event editing window
- **Monitor View:** Read-only view with activity log and live preview
- **Prefix:** Text added to beginning of source names
- **Rollback:** Revert changes if something fails
- **Scene:** Container in OBS where sources are placed
- **Scheduler:** The scheduler that evaluates what to show
- **Server Schedule:** Official active schedule in backend
- **Staging:** Safe preparation before visible change
- **VirtualCam:** OBS virtual camera for video output
- **WebSocket:** Protocol for real-time communication
- **WHEP:** WebRTC-HTTP Egress Protocol for low latency streaming
- **Working Schedule:** Local copy of schedule in Editor that may diverge from server

---

## 12. Contact and Support

### Help Resources

- **Technical documentation:** Consult complete technical specifications
- **Application logs:** Review log file for error details
- **OBS Community:** For OBS Studio specific issues

### Version Information

- **Current version:** Beta 0.1
- **Release date:** October 2025

### Version Features

**Implemented in Beta 0.1:**
- Dual view system (Monitor/Editor)
- Triple status indicator system
- WebRTC preview with WHEP protocol
- Automatic schedule hot-reload
- 5-step staging system
- Automatic reconnection with state synchronization
- Real-time activity log
- Complete recurring event management

**Known limitations:**
- Limited REST API
- No event templates
- Manual backup only

### Upcoming Features (Roadmap)

- Template system for common events
- Complete REST API for external integration
- Scheduled automatic backup
- Broadcast statistics and analytics
- Support for multiple simultaneous scenes
- Visual transform editor
- Import from Google Calendar/iCal

---

**Scene Scheduler Beta 0.2 - User Manual**
© 2025 - All rights reserved