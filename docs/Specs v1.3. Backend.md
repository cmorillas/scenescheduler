# **Scene Scheduler for OBS - Backend Specification**

**Status:** Prescriptive
**Version:** 1.3
**Date:** 2025-10-16

This document defines the architectural blueprint for the Scene Scheduler backend system. It serves as:

* Complete specification for implementing the system from scratch
* Architectural reference for system modifications
* Design contract independent of any specific implementation

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Technology Stack](#2-technology-stack)
3. [System Configuration](#3-system-configuration)
4. [Data Model](#4-data-model)
5. [Backend Architecture](#5-backend-architecture)
6. [Event System](#6-event-system)
7. [Backend Modules](#7-backend-modules)
8. [WebServer Module](#8-webserver-module)
9. [Logging Guide](#9-logging-guide)
10. [Glossary](#10-glossary)
11. [Build and Distribution](#11-build-and-distribution)

---

## 1. Project Overview

### 1.1. Objective

The system SHALL function as an autonomous backend server controlling OBS Studio via obs-websocket, with a modular, reactive, event-based architecture designed for continuous 24/7 operation.

### 1.2. System Context

The backend is a complete autonomous system that:

- Controls OBS Studio via obs-websocket protocol
- Reads schedule.json as single source of truth for program scheduling
- Continuously evaluates which OBS scene should be visible
- Switches programs in OBS according to schedule
- Operates standalone without any frontend or human intervention

The backend optionally provides:

- Embedded web server serving static frontend application
- WebSocket API for bidirectional communication with browser clients
- Live preview streaming via WebRTC (WHEP protocol)
- Activity logging stream for real-time monitoring

The frontend is purely optional. The backend performs all scheduling and OBS control independently. When present, the frontend serves two purposes:

1. **Editor**: Visual calendar-based interface for editing schedule.json
2. **Monitor**: Real-time observation of backend operations (live preview, logs, connection status)

The frontend does NOT control OBS - it only edits the schedule and observes backend operations.

### 1.3. Architectural Principles

**State Responsibility**

Each module is responsible for its own state and decides when to act. Commands and events are declarations of "desired state", not imperative orders to "do this now".

**Design Principles**

- **Source of Truth**: schedule.json is the single source of truth for scheduling
- **Event-Based Architecture**: Central EventBus decouples all modules
- **Convergence to Desired State**: Modules react to events and converge toward target state
- **Idempotent Operations**: Actions can be repeated without causing inconsistencies

### 1.4. Two-Scene Staging Pattern

The system SHALL use two OBS scenes (scheduleScene and scheduleSceneTmp) with a staging mechanism for program switches.

**Problem Solved:** OBS exhibits visual glitches when creating/configuring scene items in a visible scene. Users see intermediate states: blank frames, incorrect positioning, loading artifacts.

**Benefits:**

- Zero visible artifacts during switches
- Professional broadcast quality
- Clear separation between preparation and presentation

---

## 2. Technology Stack

| Layer | Technology | Purpose |
| :---- | :---- | :---- |
| Language | Go | Primary language |
| Logging | log/slog | Level-based structured logging |
| GUI | Fyne (fyne.io/fyne/v2) | Native status and log window |
| HTTP Server | net/http | Serve static files and API |
| WebSocket | github.com/gorilla/websocket | Bidirectional communication |
| OBS Client | github.com/andreykaipov/goobs | obs-websocket protocol |
| Media Capture | github.com/pion/mediadevices | OBS virtual camera capture |
| Video Streaming | github.com/pion/webrtc/v4 | WebRTC preview (WHEP) |
| Event System | Custom eventbus package | Pub/Sub to decouple modules |
| File Watcher | github.com/fsnotify/fsnotify | Detect changes in schedule.json |

---

## 3. System Configuration

### 3.1. Configuration Structure

The system SHALL read configuration from config.json defining all module parameters.

**scheduler Section:**

```json
{
  "scheduler": {
    "defaultSource": {
      "name": "string",
      "inputKind": "string",
      "uri": "string",
      "inputSettings": {},
      "transform": {}
    }
  }
}
```

Defines fallback source when no program is scheduled. All fields follow the same schema as program sources in schedule.json.

**mediaSource Section:**

```json
{
  "mediaSource": {
    "videoDeviceIdentifier": "string",
    "audioDeviceIdentifier": "string",
    "quality": "low"
  }
}
```

Configures OBS virtual camera capture for live preview streaming.

**webServer Section:**

```json
{
  "webServer": {
    "port": "8080",
    "user": "string",
    "password": "string",
    "enableTls": false,
    "certFilePath": "string",
    "keyFilePath": "string"
  }
}
```

Defines HTTP server behavior. Basic authentication and TLS are optional.

**obs Section:**

```json
{
  "obs": {
    "host": "localhost",
    "port": 4455,
    "password": "string",
    "reconnectInterval": 5,
    "scheduleScene": "string",
    "scheduleSceneTmp": "string",
    "sourceNamePrefix": "string"
  }
}
```

scheduleScene and scheduleSceneTmp are REQUIRED. sourceNamePrefix identifies managed sources for safe cleanup.

**paths Section:**

```json
{
  "paths": {
    "ffmpeg": "string",
    "hlsBase": "string",
    "logFile": "string",
    "schedule": "schedule.json"
  }
}
```

Defines filesystem paths for resources and data files.

---

## 4. Data Model

### 4.1. Schedule Format (Version 1.0)

The system SHALL read schedule.json conforming to the following schema:

```json
{
  "version": "1.0",
  "scheduleName": "Schedule Name",
  "schedule": [
    {
      "id": "string",
      "title": "string",
      "enabled": true,
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
        "inputKind": "ffmpeg_source|browser_source|...",
        "uri": "string",
        "inputSettings": {},
        "transform": {}
      },
      "timing": {
        "start": "YYYY-MM-DDTHH:MM:SSZ",
        "end": "YYYY-MM-DDTHH:MM:SSZ",
        "isRecurring": false,
        "recurrence": {
          "daysOfWeek": ["MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"],
          "startRecur": "YYYY-MM-DD",
          "endRecur": "YYYY-MM-DD"
        }
      },
      "behavior": {
        "onEndAction": "hide|none|stop",
        "preloadSeconds": 0
      }
    }
  ]
}
```

**Field Semantics:**

- **enabled**: Backend control - if false, Scheduler ignores this program
- **title**: Display name for UI (frontend concern)
- **source.name**: Technical identifier for OBS operations (backend concern)
- **general**: Visual presentation fields (frontend concern)
- **timing**: When the program should be active (backend concern)
- **behavior**: Execution parameters (backend concern)

---

## 5. Backend Architecture

### 5.1. Standard Module File Organization

All backend modules SHALL follow this standardized file structure:

**Required Files:**

| File | Purpose | Critical Requirement |
| :---- | :---- | :---- |
| constructor.go | Struct definition, constructor, public API | MUST call subscribeToEvents() before returning |
| runner.go | Run(), Stop(), cleanup(), lifecycle | Main loop and graceful shutdown |
| events.go | Event subscriptions and handlers | subscribeToEvents() implementation |

**Constructor Pattern (MANDATORY):**

```go
func New(appCtx context.Context, cfg *Config, bus *EventBus) *Module {
    m := &Module{
        config: cfg,
        bus:    bus,
    }

    // CRITICAL: Create context before subscribing
    m.ctx, m.cancel = context.WithCancel(appCtx)

    // CRITICAL: Subscribe before returning
    m.subscribeToEvents()

    return m
}
```

**Why This Order Matters:**

- Context must exist before event handlers can use it
- Event subscriptions must complete before module is considered "ready"
- Prevents race conditions where events fire before module initialization
- Module is immediately operational when New() returns

**Optional Files:**

| File | When to Create | Content |
| :---- | :---- | :---- |
| types.go | ≥3 structs, enums or custom errors | DTOs, enums, error definitions |
| api.go | constructor.go exceeds ~200 lines | Public methods extracted |
| [theme].go | ≥3 related methods with cohesive purpose | Thematic grouping (scene.go, fsm.go, etc.) |
| misc.go | Private methods without clear theme | Helper methods with receiver |
| helpers.go | Pure functions (no receiver) | Stateless utility functions |

**File Organization Criteria:**

- **constructor.go**: Struct + New() + public methods (until ~200 lines)
- **api.go**: Public methods extracted when constructor.go grows too large
- **events.go**: All EventBus subscription and handler logic
- **runner.go**: Lifecycle management and main loop
- **misc.go**: Private methods with state (methods with receiver)
- **helpers.go**: Pure functions without state
- **[theme].go**: Create when ≥3 methods form cohesive group with meaningful complexity

**internal/ Integration Pattern:**

When a module has an internal/ component, a corresponding integration file MUST exist:

```text
mediasource/
├── feed.go      # REQUIRED: All interaction with internal/feed
└── internal/
    └── feed/
        └── *.go
```

The integration file SHALL contain ALL communication with the internal component: initialization, callbacks, public method wrappers, cleanup.

### 5.2. File Header Format (REQUIRED)

Every file MUST include this standardized header:

```go
// backend/<module>/<file>.go
//
// <Brief description of file purpose>
//
// Contents:
// - <Section 1 name>
// - <Section 2 name>
```

**Example:**

```go
// backend/scheduler/constructor.go
//
// Scheduler module constructor, configuration, and public API.
//
// Contents:
// - Type definitions
// - Constructor (New) - calls subscribeToEvents()
// - Public API methods
// - State management
```

### 5.3. Documentation Requirements

**Public Methods (REQUIRED):**

```go
// LoadSchedule reads and parses the schedule file from disk.
// It validates the schedule structure and returns an error if invalid.
// This method is idempotent and can be called multiple times safely.
//
// Returns:
//   - error: nil on success, validation error if schedule is malformed
func (s *Scheduler) LoadSchedule() error {
    // Implementation
}
```

**Event Handlers (REQUIRED):**

```go
// handleTargetProgramState converges OBS state to target program.
// Compares target with current state and switches if divergent.
// This operation is idempotent - receiving the same target multiple times
// causes no additional side effects.
//
// Topic: scheduler.state.targetProgram
func (o *OBSClient) handleTargetProgramState(e eventbus.Event) {
    // Implementation
}
```

**Struct Definitions (REQUIRED):**

```go
// Scheduler evaluates schedule.json and publishes target program state.
// It does not control OBS directly, only declares desired state.
type Scheduler struct {
    // --- Configuration ---
    config   *Config   // Module configuration from config.json
    bus      *EventBus // EventBus for module communication

    // --- Schedule Data ---
    schedule []Program // Current loaded schedule

    // --- Lifecycle Management ---
    ctx    context.Context
    cancel context.CancelFunc

    // --- Synchronization ---
    mu          sync.RWMutex
    stopOnce    sync.Once
    cleanupOnce sync.Once
}
```

### 5.4. context.Context Discipline

**Core Principles:**

Each module SHALL initialize its context in the constructor before subscribing to events. This guarantees event handlers can safely use the context from the moment of subscription.

**Implementation Pattern:**

```go
func New(appCtx context.Context, cfg *Config, bus *EventBus) *Module {
    m := &Module{config: cfg, bus: bus}

    // Create derived context BEFORE subscription
    m.ctx, m.cancel = context.WithCancel(appCtx)

    // Now safe to subscribe - handlers can use m.ctx
    m.subscribeToEvents()

    return m
}

func (m *Module) Run() {
    defer m.cleanup()

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.evaluate()
        case <-m.ctx.Done():
            return
        }
    }
}

func (m *Module) Stop() {
    m.stopOnce.Do(func() {
        if m.cancel != nil {
            m.cancel()
        }
    })
}

func (m *Module) cleanup() {
    m.cleanupOnce.Do(func() {
        m.unsubscribeAllEvents()
        // Other cleanup operations
    })
}
```

**Rules:**

1. Context MUST be created in constructor using context.WithCancel(appCtx)
2. Context creation MUST happen before calling subscribeToEvents()
3. Internal goroutines MUST use m.ctx for cancellation
4. Stop() MUST invoke m.cancel() to trigger graceful shutdown
5. Both Stop() and cleanup() MUST use sync.Once for idempotency

### 5.5. Module Communication Rules

**Top-Level Module Communication:**

All top-level modules (Scheduler, OBSClient, WebServer, MediaSource) SHALL communicate exclusively through EventBus. No direct references or method calls between top-level modules.

**Parent ↔ Internal Component Communication:**

Parent modules SHALL communicate with their internal/ components via callbacks or channels. Internal components MUST NOT use EventBus. Parent acts as ambassador, translating external events to internal notifications.

**Example:**

```go
// ✅ GOOD: Internal component with callback
type Handler struct {
    onChange func()  // Direct callback to parent
}

// Parent coordinates all external interaction
type WebServer struct {
    wsHandler *websocket.Handler
    bus       *EventBus
}

func (ws *WebServer) subscribeToEvents() {
    // Parent decides if internal should react to external events
    ws.bus.Subscribe("obs.disconnected", func(e Event) {
        ws.wsHandler.Pause()
    })
}
```

### 5.6. Lifecycle Invariants

1. New() MUST call subscribeToEvents() before returning
2. Stop() MUST be idempotent using sync.Once
3. cleanup() MUST always call unsubscribeAllEvents()
4. cleanup() MUST be idempotent using sync.Once
5. Run() MUST implement efficient for/select loop with context cancellation
6. Context MUST be initialized before event subscription

**Why Subscription in Constructor:**

```go
// ❌ RACE CONDITION: Events can be lost
func main() {
    mediaSource := mediasource.New(config, bus)
    webServer := webserver.New(config, bus)

    go mediaSource.Run(ctx)  // Starts publishing events
    go webServer.Run(ctx)    // Subscribes in Run()
    // If MediaSource publishes before WebServer subscribes → event lost
}

// ✅ NO RACE: All subscribed before any Run()
func main() {
    // All constructors subscribe internally
    mediaSource := mediasource.New(ctx, config, bus)  // Subscribed
    webServer := webserver.New(ctx, config, bus)      // Subscribed

    // Now safe to start publishing
    go mediaSource.Run()
    go webServer.Run()
}
```

---

## 6. Event System

### 6.1. Description

In-memory publish/subscribe bus that decouples top-level autonomous modules through event-based communication. EventBus is used exclusively for communication between top-level modules. Internal components within module/internal/ SHALL use direct coupling (callbacks, channels).

### 6.2. Event Contracts

| Topic | Emitter | Receiver | Purpose |
| :---- | :---- | :---- | :---- |
| scheduler.state.targetProgram | Scheduler | OBSClient | Declares which program should be active |
| obs.program.changed | OBSClient | WebServer, GUI | Confirms program change completed in OBS (ON AIR) |
| obs.system.connected | OBSClient | WebServer, Scheduler, GUI | OBS connection established, includes version |
| obs.system.disconnected | OBSClient | WebServer, Scheduler, GUI | OBS disconnection, includes reason |
| obs.virtualcam.started | OBSClient | WebServer | VirtualCam started broadcasting |
| obs.virtualcam.stopped | OBSClient | WebServer | VirtualCam stopped broadcasting |
| webserver.command.getSchedule | WebServer | Scheduler | Frontend requests current schedule |
| webserver.command.commitSchedule | WebServer | Scheduler | Frontend wants to save new schedule |
| webserver.command.getStatus | WebServer | OBSClient | Frontend requests current OBS/VirtualCam state |
| webserver.response.status | OBSClient | WebServer | Response with current connection status |
| mediasource.tracks.ready | MediaSource | WebServer | Video/audio tracks available for WHEP |
| mediasource.tracks.stopped | MediaSource | WebServer | Capture stopped, tracks no longer available |
| sourcepreview.request | WebServer | SourcePreview | Frontend requests preview for a source |
| sourcepreview.ready | SourcePreview | WebServer | HLS preview stream is ready, provides URL |
| sourcepreview.error | SourcePreview | WebServer | Preview generation failed with error |
| sourcepreview.stop | WebServer | SourcePreview | Frontend stops active preview |

### 6.3. TargetProgramState Contract

```go
type TargetProgramState struct {
    Timestamp      time.Time
    TargetProgram  *ProgramData // Program that should be active (nil = none)
    NextProgram    *ProgramData // Next scheduled program (for UI/logging)
    SeekOffset     time.Duration // Offset if program already started
}
```

**Semantics:**

- State declaration, not a command
- Published every evaluation cycle (every second)
- Receiver decides whether to act
- Multiple publications of same state cause no side effects (idempotent)

### 6.4. ProgramChanged Contract

The system SHALL emit obs.program.changed only if there is a real program change (idempotency) and after OBS has applied the change.

**Payload:**

```go
type ProgramChanged struct {
    Timestamp       time.Time
    PreviousProgram *ProgramData
    NewProgram      *ProgramData
    Trigger         string // "scheduler" | "manual" | "recovery"
    SeekOffsetMs    int
    CorrelationId   string
}
```

**Intent vs Confirmation:**

- scheduler.state.targetProgram represents the **intent** of desired state
- obs.program.changed represents the **confirmation** that change is complete and ON AIR
- Consumers requiring "ON AIR" state MUST subscribe to obs.program.changed

### 6.5. Status Query System

The system SHALL support initial state synchronization via query-response pattern. When a client connects, it MAY request current OBS and VirtualCam state. OBSClient SHALL respond with thread-safe state snapshot, enabling accurate status display without race conditions.

**Request Event:**

```go
type GetStatusRequested struct {
    ResponseChannel chan<- StatusResponse
}
```

**Topic:** webserver.command.getStatus

**Response Event:**

```go
type StatusResponse struct {
    IsConnected      bool
    OBSVersion       string
    VirtualCamActive bool
}
```

**Topic:** webserver.response.status

**Flow:**

```
1. Client connects (initial or reconnect) → WebSocket established
2. WebSocket handler publishes GetStatusRequested with response channel
3. OBSClient.GetCurrentStatus() queries internal state (thread-safe)
4. OBSClient sends StatusResponse through channel
5. WebServer sends 'currentStatus' message to client
6. Frontend updates status indicators immediately
```

**Reconnection Handling:**

The frontend SHALL automatically reconnect after connection loss and re-request status. This ensures accurate state synchronization even if backend state changed during disconnection (e.g., OBS connected, VirtualCam started). The backend SHALL respond identically to reconnect requests as to initial connections.

### 6.6. VirtualCam Event Broadcasting

The system SHALL broadcast VirtualCam state changes (started/stopped) to all connected clients. This enables real-time preview availability indication in the UI.

**Events:**

```go
type VirtualCamStarted struct {
    Timestamp time.Time
}

type VirtualCamStopped struct {
    Timestamp time.Time
}
```

**Trigger Conditions:**

- VirtualCam starts: User clicks "Start Virtual Camera" in OBS
- VirtualCam stops: User clicks "Stop Virtual Camera" in OBS

**Behavior:** WebServer receives these events and broadcasts corresponding WebSocket messages to all connected clients, enabling/disabling preview controls appropriately.

### 6.7. SourcePreview Event Contracts

The system SHALL support source preview requests and responses for on-demand HLS generation.

**Request Event:**

```go
type SourcePreviewRequest struct {
    ClientID    string
    SourceURI   string
    InputKind   string
    InputSettings map[string]interface{} // Optional OBS input settings
}
```

**Topic:** sourcepreview.request

**Ready Event:**

```go
type SourcePreviewReady struct {
    RequestID   string
    ClientID    string
    HLSURL      string    // e.g., "/hls/preview-{requestID}/playlist.m3u8"
    Timestamp   time.Time
}
```

**Topic:** sourcepreview.ready

**Error Event:**

```go
type SourcePreviewError struct {
    RequestID   string
    ClientID    string
    Error       string
    Timestamp   time.Time
}
```

**Topic:** sourcepreview.error

**Stop Event:**

```go
type SourcePreviewStop struct {
    RequestID   string
    ClientID    string
}
```

**Topic:** sourcepreview.stop

**Flow:**

```
1. Frontend clicks "Preview Source" in modal
2. WebSocket handler receives requestPreview message
3. WebServer publishes SourcePreviewRequest to EventBus
4. SourcePreview module receives event, spawns hls-generator
5. SourcePreview monitors for playlist.m3u8 file
6. SourcePreview publishes SourcePreviewReady with HLS URL
7. WebServer receives event, sends previewReady to client
8. Frontend loads HLS stream using HLS.js
9. User clicks "Stop Preview"
10. Frontend sends stopPreview WebSocket message
11. WebServer publishes SourcePreviewStop to EventBus
12. SourcePreview kills process and cleans up filesystem
```

---

## 7. Backend Modules

### 7.1. Scheduler

**Responsibility:** Evaluate which program should be active at each moment and publish target state. Does not decide if change is necessary, simply declares what should be broadcasting. Monitors schedule.json for hot-reload.

**Operation:**

The Scheduler SHALL execute every second:

1. Search for scheduled event for current instant
2. If no event, use scheduler.defaultSource from config.json
3. Publish TargetProgramState on EventBus

The Scheduler SHALL always publish target state, regardless of whether it changed:

- Does not maintain "active program" state
- Does not compare with previous states
- Event declares "this is what should be broadcasting now"
- Idempotent operation by design

**Frontend Communication:**

The Scheduler SHALL respond to commands from WebServer:

- **GetScheduleRequested**: Read current schedule from memory, send back through WebServer
- **CommitScheduleRequested**: Validate structure, write to schedule.json if valid, send success/error response

**Hot-Reload:** After successful commit, FileWatcher detects modification, Scheduler reloads, WebServer broadcasts scheduleChanged to all clients.

### 7.2. OBSClient

**Responsibility:** Keep OBS state convergent with target state published by Scheduler. It is the only module that knows and manages which program is currently active.

**Lifecycle:**

Upon reaching Connected state, the system SHALL:

1. Execute setupScene(): Ensure existence of scheduleScene and scheduleSceneTmp, cleanup scene items
2. Launch monitoring goroutines
3. Subscribe to OBS events

**Target State Management:**

On receiving TargetProgramState, OBSClient SHALL:

1. Compare received targetProgram with internal activeProgram
2. If targetProgram == activeProgram → Do nothing (already convergent)
3. If targetProgram ≠ activeProgram → Execute program change

This operation is idempotent - receiving same target state multiple times causes no side effects.

**Program Change Process:**

The system SHALL execute six-step staging process:

1. **Staging in TEMP**: Create input and scene item in scheduleSceneTmp (hidden), apply transforms
2. **Promotion to MAIN**: Duplicate scene item to scheduleScene
3. **Activation**: Make new item visible (public moment)
4. **Staging Cleanup**: Remove temporary scene item from scheduleSceneTmp
5. **Previous Retirement**: Hide and remove scene item and input from previous program
6. **State Update**: Set activeProgram = targetProgram (convergence achieved)

On failure: Rollback, clean resources, maintain previous activeProgram.

After successfully executing program change, OBSClient SHALL publish obs.program.changed with payload defined in ProgramChanged Contract.

**Status Query API:**

OBSClient SHALL provide thread-safe GetCurrentStatus() method:

```go
// GetCurrentStatus returns current connection status.
// Thread-safe and can be called from any goroutine.
func (o *OBSClient) GetCurrentStatus() ConnectionStatus {
    // Returns: IsConnected, OBSVersion, VirtualCamActive
}
```

Used by WebServer to respond to status queries from newly connected clients.

**VirtualCam Event Handlers:**

OBSClient SHALL monitor OBS WebSocket events for VirtualCam state changes and publish to EventBus:

- Detect VirtualcamStateChanged event from OBS
- If OutputActive: Publish obs.virtualcam.started
- If not OutputActive: Publish obs.virtualcam.stopped

### 7.3. MediaSource

**Responsibility:** Capture OBS virtual camera output and provide WebRTC tracks for WHEP streaming.

**Operation:**

The MediaSource SHALL:

1. Monitor for OBS Virtual Camera device availability
2. When device available, start capture of video and audio
3. Publish mediasource.tracks.ready with track references
4. When capture stops or device unavailable, publish mediasource.tracks.stopped

**Integration:** Communicates with WebServer exclusively through EventBus. No direct coupling.

### 7.4. SourcePreview

**Responsibility:** Generate on-demand HLS preview streams for individual program sources before they are scheduled. This allows users to verify source content in the editor. The module spawns and manages `hls-generator` processes for media sources. Completely independent from live WebRTC preview of OBS output.

**Architecture Decision:**

SourcePreview is a top-level module (not internal to WebServer) because:
- **Has complete ownership of a system resource type**: Manages hls-generator processes and temporary filesystem exclusively
- It has its own lifecycle with Run() loop for cleanup and monitoring
- It has complete domain responsibility: "Generate HLS previews of sources on demand"
- It is testable in isolation without mounting WebServer
- Future reusability: CLI tools or GUI could request previews directly
- Complexity warrants separation: ~800+ lines of code with process management, filesystem operations, session tracking, and timeout handling

**Technical Decisions:**

The following technical decisions have been made for implementation:

1. **File Structure** (6 files, all < 250 lines):
   - `types.go` - Structs, constants, errors
   - `constructor.go` - SourcePreview struct, New() function
   - `process.go` - hls-generator spawn/kill, filesystem operations
   - `session.go` - Session tracking, timeout management
   - `events.go` - EventBus subscriptions and handlers
   - `runner.go` - Run() loop, Stop(), cleanup()

2. **hls-generator Binary Location:**
   - Primary: Same directory as scenescheduler executable
   - Fallback: System PATH
   - NOT configurable in config.json
   - Discovery at module initialization
   - If not found: log error, allow startup, fail requests gracefully

3. **RequestID Generation:**
   - Unix timestamp (seconds): `1729123456`
   - If collision in same second: append suffix `-1`, `-2`, etc.
   - Helper function for human-readable logs: `requestIDToDate()`
   - Guarantees uniqueness without UUID dependency

4. **Timeouts (hardcoded constants):**
   ```go
   playlistWaitTimeout = 30 * time.Second  // Wait for playlist.m3u8
   sessionTimeout      = 5 * time.Minute   // Abandoned session
   cleanupInterval     = 30 * time.Second  // Run() loop ticker (unused)
   processKillTimeout  = 5 * time.Second   // SIGTERM → SIGKILL
   ```

5. **State Management:**
   - NO FSM (Finite State Machine)
   - Simple session tracking with implicit states
   - State determined by: Process existence, playlist file existence, map membership

6. **Filesystem Cleanup Strategy:**
   - **Startup**: Delete entire `/tmp/hls-previews/` directory (clean slate)
   - **Shutdown**: Kill all processes with Wait(), then delete entire directory
   - **Runtime**: No periodic cleanup (single instance guarantee)
   - Aggressive strategy justified: only one backend instance ever runs

7. **Playlist Detection:**
   - Polling method: Check file existence every 500ms
   - Timeout: 30 seconds maximum
   - NO fsnotify (filesystem watcher) - polling is sufficient for this use case

**Operation:**

The SourcePreview module SHALL execute the following workflows:

**On sourcepreview.request event:**

1. Generate unique requestID using Unix timestamp with collision handling
2. Create unique temporary directory: `{hlsBase}/preview-{requestID}/`
3. Find hls-generator binary (same directory as executable, fallback to PATH)
4. Spawn `hls-generator` process: `hls-generator {sourceURI} {tempDir}`
5. Poll filesystem every 500ms for `playlist.m3u8` creation (30 second timeout)
6. Publish `sourcepreview.ready` with requestID and HLS URL when playlist exists
7. Publish `sourcepreview.error` if binary not found, process fails, or timeout exceeded

**On sourcepreview.stop event:**

1. Lookup active session by requestID
2. Send SIGTERM to hls-generator process
3. Wait up to 5 seconds, then SIGKILL if still running
4. Delete entire temporary directory: `os.RemoveAll(tempDir)`
5. Remove session from activePreviews map

**Run() Loop Responsibilities:**

The SourcePreview Run() loop SHALL:

1. **Startup cleanup**: Delete entire `/tmp/hls-previews/` directory for clean slate
2. **Wait for shutdown**: Block on context cancellation (no periodic tasks needed)
3. Single instance guarantee means no runtime orphan detection required

**Shutdown Sequence:**

The cleanup() method SHALL:

1. Kill all active processes with SIGTERM
2. Wait for each process termination (Cmd.Wait())
3. Sleep 200ms safety buffer
4. Delete entire `/tmp/hls-previews/` directory
5. Clear activePreviews map

**State Management:**

The SourcePreview SHALL maintain internal state tracking all active preview sessions:

```go
type SourcePreview struct {
    logger           *logger.Logger
    bus              *eventbus.EventBus
    paths            *config.PathsConfig

    // Lifecycle
    ctx              context.Context
    cancelCtx        context.CancelFunc
    stopOnce         sync.Once
    cleanupOnce      sync.Once
    unsubscribeFuncs []func()

    // Active sessions
    mu               sync.RWMutex
    activePreviews   map[string]*PreviewSession // Key: requestID
}

type PreviewSession struct {
    RequestID   string
    ClientID    string        // For tracking which client requested it
    SourceURI   string
    InputKind   string
    TempDir     string
    Process     *ProcessHandle
    CreatedAt   time.Time
    LastAccess  time.Time     // Updated when client fetches HLS segments
}

type ProcessHandle struct {
    Cmd       *exec.Cmd
    PID       int
    StartedAt time.Time
}
```

**Integration with hls-generator:**

The SourcePreview module depends on the external `hls-generator` binary. Binary discovery follows this order:

1. Same directory as scenescheduler executable: `./hls-generator`
2. System PATH: `exec.LookPath("hls-generator")`

Configuration in `config.json`:

```json
{
  "paths": {
    "hlsBase": "/tmp/hls-previews"
  }
}
```

The module SHALL attempt binary discovery during initialization:
- If found: Log info with path
- If not found: Log error, continue startup, fail preview requests gracefully with error message

No panic or fatal error if binary not found - allows system to run without preview capability.

**Browser Source Handling:**

The SourcePreview module SHALL NOT handle browser_source types. Browser sources are rendered directly in the frontend by setting the URL in the video element's src attribute. The frontend is responsible for detecting inputKind and handling browser sources locally.

---

## 8. WebServer Module

### 8.1. Responsibility

The WebServer SHALL serve static files and provide communication APIs for browser clients. It acts as a bridge between HTTP/WebSocket/WebRTC protocols and internal backend systems.

### 8.2. Key Functions

**Static File Serving:**

- Serve frontend application files from configured directory
- Support TLS when enableTls is configured
- Optional Basic Authentication

**WebSocket Gateway:**

Translate between JSON messages over WebSocket (browser) and EventBus events (backend):

- Browser → WebSocket → WebServer → EventBus → Modules
- Modules → EventBus → WebServer → WebSocket → Browser

**WHEP Streaming:**

Provide WebRTC endpoint for streaming live OBS output:

- Receive video/audio tracks from MediaSource via EventBus
- Maintain current track references with mutex protection
- Serve WHEP-compatible WebRTC stream to browser clients
- Return 503 Service Unavailable if tracks not yet available (expected behavior)

**Log Streaming:**

Stream backend activity logs through WebSocket for real-time monitoring.

### 8.3. WebSocket Protocol

All messages SHALL use JSON structure: `{"action": "string", "payload": {}}`

**Client → Server Messages:**

| Action | Purpose |
| :---- | :---- |
| getStatus | Request current OBS connection and VirtualCam state |
| getSchedule | Request current schedule |
| commitSchedule | Save new schedule |
| requestSourcePreview | Request HLS preview for a program source (payload: sourceURI, inputKind, inputSettings) |
| stopSourcePreview | Stop currently active source preview (payload: requestID) |

**Server → Client Messages:**

| Action | Purpose |
| :---- | :---- |
| currentStatus | Initial state synchronization response (obsConnected, obsVersion, virtualCamActive) |
| obsConnected | OBS connection established (includes version, timestamp) |
| obsDisconnected | OBS connection lost (includes reason, timestamp) |
| virtualCamStarted | VirtualCam stream available (includes timestamp) |
| virtualCamStopped | VirtualCam stream unavailable (includes timestamp) |
| currentSchedule | Response to getSchedule |
| scheduleChanged | Broadcast when schedule.json modified |
| commitSuccess | Schedule save succeeded |
| commitError | Schedule save failed validation |
| programChanged | Program switch completed (broadcast) |
| sourcePreviewReady | Source preview HLS stream ready (includes requestID, hlsURL) |
| sourcePreviewError | Source preview generation failed (includes requestID, error) |

### 8.4. Initial State Synchronization

**Problem:** Events fire before client connects. When frontend client connects, backend may have already established connections and activated features. These events fired before client was listening.

**Solution:** On WebSocket connection establishment, WebServer SHALL automatically:

1. Create response channel
2. Publish GetStatusRequested event with channel
3. Wait for StatusResponse from OBSClient
4. Send currentStatus message to client with obsConnected, obsVersion, virtualCamActive
5. Client updates status indicators immediately

This ensures accurate state display without race conditions or missed events.

### 8.5. Event Subscriptions

WebServer SHALL subscribe to EventBus events for broadcasting:

- obs.system.connected → Broadcast obsConnected
- obs.system.disconnected → Broadcast obsDisconnected
- obs.virtualcam.started → Broadcast virtualCamStarted
- obs.virtualcam.stopped → Broadcast virtualCamStopped
- obs.program.changed → Broadcast programChanged
- Schedule file change → Broadcast scheduleChanged
- sourcepreview.ready → Broadcast sourcePreviewReady
- sourcepreview.error → Broadcast sourcePreviewError

### 8.6. HTTP Routes

The WebServer SHALL expose the following HTTP routes:

**Static Frontend** (embedded files):
```
GET /              → Serve embedded frontend (index.html)
GET /main.css      → Serve embedded CSS
GET /main.mjs      → Serve embedded JavaScript
GET /components/*  → Serve embedded components
```

**HLS Preview Streaming** (dynamic filesystem):
```
GET /hls/preview-{requestID}/playlist.m3u8  → HLS playlist
GET /hls/preview-{requestID}/*.ts           → HLS video segments
```

Implementation:
```go
hlsDir := pathsConfig.HLSBase // "/tmp/hls-previews"
http.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir(hlsDir))))
```

**WHEP Live Preview** (WebRTC):
```
POST   /whep   → Create WebRTC session
DELETE /whep   → Delete WebRTC session
```

**WebSocket** (bidirectional messaging):
```
GET /ws → Upgrade to WebSocket connection
```

**Notes:**
- NO CORS headers needed (all served from same origin)
- HLS files served directly from filesystem (not embedded)
- Frontend files served from embedded FS
- WHEP and WebSocket use dedicated handlers

### 8.7. Implementation Notes

- WebServer maintains list of active WebSocket connections
- Broadcasts events to all connections simultaneously
- Each client request gets individual response via its own connection
- No business logic - purely protocol translator
- Video/audio tracks received via EventBus and stored with mutex protection
- WHEP handler returns 503 if tracks not available (expected, not error)

---

## 9. Logging Guide

### 9.1. Log Levels

The system SHALL use structured logging with log/slog:

| Level | Use | Examples |
| :---- | :---- | :---- |
| Debug | Development/troubleshooting | Detailed state transitions, event payloads, internal calculations |
| Info | Normal operations | Connection established, program change, schedule reload, cleanup complete |
| Warn | Non-critical failures | Resource not found during cleanup, transform application failures |
| Error | Critical failures | Scene setup failed, duplication failures, connection loss |

---

## 10. Glossary

| Term | Definition |
| :---- | :---- |
| EventBus | In-memory publish/subscribe bus for decoupled communication |
| Hot-reload | Automatic configuration reload without restarting application |
| Input | Content source in OBS (video, image, browser, etc.) |
| Scene Item | Instance of an input placed in an OBS scene |
| Staging | Prior resource preparation in temporary scene before making visible |
| WHEP | WebRTC-HTTP Egress Protocol (low-latency streaming protocol) |

---

## 11. Build and Distribution

### 11.1. Build Commands

**Development Build:**

```sh
go build -o scenescheduler
```

**Production Build:**

```sh
go build -ldflags="-s -w -extldflags=-static" -o scenescheduler
```

Flags: `-s -w` strips debug information, `-extldflags=-static` links all libraries statically, producing single self-contained executable.

**Cross-Platform Builds:**

```sh
# Windows
GOOS=windows GOARCH=amd64 go build -o scenescheduler.exe

# Linux
GOOS=linux GOARCH=amd64 go build -o scenescheduler

# macOS
GOOS=darwin GOARCH=amd64 go build -o scenescheduler
```

### 11.2. Executable Compression

The binary MAY be compressed using UPX:

```sh
# Maximum compression
upx --best --lzma scenescheduler

# Ultra-brute force (slower, better compression)
upx --ultra-brute scenescheduler
```

Typical reduction: 60-70% of original size. Trade-off: Slightly slower startup time.

### 11.3. Command Line Utilities

**List Media Devices:**

```sh
./scenescheduler -list-devices
```

Scans system cameras and microphones, displays detailed list with Friendly Name and DeviceID for config.json configuration. Terminates immediately without starting services.

---

**Last Updated:** 2025-10-16
**Document Version:** 1.3

**Changelog v1.3:**
- Added SourcePreview module specification (section 7.4)
- Added SourcePreview event contracts (section 6.7)
- Updated WebSocket protocol with source preview messages
- Updated WebServer event subscriptions and HTTP routes (section 8.6)
- Documented architecture decision criteria for top-level vs internal modules
- Documented all technical implementation decisions:
  - File structure (6 files < 250 lines)
  - hls-generator binary discovery strategy
  - RequestID generation using Unix timestamp
  - Hardcoded timeout constants
  - No FSM, simple session tracking
  - Aggressive filesystem cleanup strategy
  - Polling-based playlist detection
