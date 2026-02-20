# **Scene Scheduler for OBS - Backend Specification**

**Status:** Prescriptive
**Version:** 1.5
**Date:** 2025-10-17

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
    "hlsPath": "hls",
    "enableTls": false,
    "certFilePath": "string",
    "keyFilePath": "string"
  }
}
```

Defines HTTP server behavior. Basic authentication and TLS are optional.

**hlsPath Security:**
- MUST be a relative path (e.g., "hls", "data/previews")
- Absolute paths rejected ("/etc/hls", "C:\hls")
- Directory traversal forbidden ("../hls")
- Validated at startup via `validateSafeRelativePath()`

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
    "logFile": "string",
    "schedule": "schedule.json"
  }
}
```

Defines filesystem paths for log and schedule files. HLS path moved to webServer section.

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

When a module has an internal/ component AND requires coordination logic, a corresponding integration file SHOULD exist:

```text
mediasource/
├── feed.go      # Bridge: Coordinates FSM + feed lifecycle
└── internal/
    └── feed/
        └── *.go

obsclient/
├── switcher.go  # Bridge: Coordinates FSM + program switching
└── internal/
    └── switcher/
        └── *.go

webserver/
├── sourcepreview.go   # Bridge: Coordinates WebSocket + sourcepreview
└── internal/
    ├── sourcepreview/
    ├── websocket/     # No bridge needed (callbacks are sufficient)
    └── whep/          # No bridge needed (event relay is sufficient)
```

**When to Create Bridge File:**

CREATE bridge file `{internal_name}.go` IF:
- ✅ Parent has coordination logic beyond simple delegation
- ✅ Multiple operations need synchronization with internal module
- ✅ FSM or state machine needs to interact with internal module
- ✅ Complex callback orchestration or error handling required

DO NOT create bridge file IF:
- ❌ Integration is purely callbacks (e.g., websocket callbacks)
- ❌ Integration is simple event relay (e.g., WHEP track forwarding)
- ❌ Parent only creates and stores instance without coordination
- ❌ Would result in redundant pass-through methods with no logic

**Bridge File Contents:**

The integration file SHALL contain:
- Initialization and configuration of internal component
- Coordination logic between parent state and internal module
- Callback implementations that involve parent logic
- Public method wrappers that add parent-level validation/logging
- Cleanup and shutdown coordination

**Examples:**

✅ **GOOD**: `mediasource/feed.go`
- Coordinates FSM transitions with feed lifecycle
- Implements performAcquisition, releaseFeed with state checks
- Handles feed failures and triggers FSM events

✅ **GOOD**: `obsclient/switcher.go`
- Coordinates program switching with FSM
- Implements convergeToState with complex state comparison
- Handles OBS scene setup and program transitions

✅ **GOOD**: `webserver/sourcepreview.go`
- Coordinates WebSocket requests with sourcepreview manager
- Parses payloads and creates callback closures
- Handles unicast responses to specific clients

❌ **NOT NEEDED**: `webserver/websocket.go`
- Callbacks already provide clean API (OnGetSchedule, OnCommitSchedule, etc.)
- No coordination logic beyond callback routing
- Constructor.go integration is sufficient

❌ **NOT NEEDED**: `webserver/whep.go`
- Simple event relay in events.go (SetTracks)
- No coordination or state synchronization
- Direct handler usage is clearer

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

Parent modules SHALL communicate with their internal/ components via callbacks or channels. Internal components MUST NOT use EventBus infrastructure (publish/subscribe). Parent acts as ambassador, translating external events to internal notifications.

**EventBus DTO Import Exception:**

Internal components MAY import shared data types from `backend/eventbus` when these types are pure DTOs (Data Transfer Objects) without behavior. These types represent domain contracts, not infrastructure.

**Allowed:**
- ✅ Importing `eventbus.Program`, `eventbus.Track`, or other pure data structs
- ✅ Using these types in function signatures and return values
- ✅ Passing these types between parent and internal components

**Forbidden:**
- ❌ Calling `eventbus.Publish()` from internal components
- ❌ Calling `eventbus.Subscribe()` from internal components
- ❌ Storing `*eventbus.EventBus` reference in internal component structs

**Rationale:** DTOs are domain contracts for data exchange, not infrastructure. Requiring duplication of identical DTOs creates unnecessary conversions and violates DRY principle without architectural benefit.

**Example:**

```go
// ✅ ALLOWED: Internal component using EventBus DTOs
package switcher

import "scenescheduler/backend/eventbus"

type Switcher struct {
    logger *logger.Logger
}

func (s *Switcher) PerformSwitch(
    current *eventbus.Program,  // ✅ DTO parameter
    target *eventbus.Program,   // ✅ DTO parameter
) (*SwitchResult, error) {
    // Pure business logic, no EventBus calls
}

type SwitchResult struct {
    PreviousProgram *eventbus.Program  // ✅ DTO in return type
    CurrentProgram  *eventbus.Program  // ✅ DTO in return type
}
```

```go
// ❌ FORBIDDEN: Internal component publishing to EventBus
package websocket

import "scenescheduler/backend/eventbus"

type Handler struct {
    bus *eventbus.EventBus  // ❌ Storing EventBus reference
}

func (h *Handler) handleMessage(msg Message) {
    eventbus.Publish(h.bus, event)  // ❌ Publishing from internal
}
```

```go
// ✅ CORRECT: Use callbacks instead
package websocket

type Handler struct {
    onMessage func(msg Message)  // ✅ Callback to parent
}

func (h *Handler) handleMessage(msg Message) {
    if h.onMessage != nil {
        h.onMessage(msg)  // ✅ Parent handles EventBus
    }
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

**Location:** `backend/webserver/internal/sourcepreview/`

**Responsibility:** Generate on-demand HLS preview streams for program sources before scheduling.

**Classification:** Internal module (direct communication, no EventBus).

**Communication:** WebSocket Handler → SourcePreview Manager (direct method calls).

**Core Behavior:**

The module SHALL spawn temporary hls-generator processes that transcode source URIs into HLS streams, allowing users to verify content before scheduling.

**Why Internal (Not Top-Level)?**

1. **Local scope:** Only WebSocket Handler needs preview capability
2. **Request-scoped:** Previews tied to specific WebSocket client lifecycle  
3. **Direct responses:** Preview ready/error sent only to requesting client via callbacks
4. **No broadcast:** Other modules don't need preview events
5. **No Run() loop:** Stateless goroutines handle each request independently
6. **Small codebase:** ~700 lines total across 5 files

**File Structure:**

```
sourcepreview/
├── types.go      (125 lines) - Structs, constants, errors
├── manager.go    (292 lines) - Constructor, public API  
├── preview.go    (124 lines) - processPreview goroutine
├── process.go    (126 lines) - Binary discovery, spawn, kill
└── session.go    (26 lines)  - Session tracking helpers
```

**Total:** 693 lines (avg 139 lines/file)

**Design Decisions:**

**1. Incremental PreviewID Pattern:**

Session tracking uses incremental PreviewID (1, 2, 3...) like EventBus subscription IDs:
- Thread-safe generation with `atomic.Uint64`
- Filesystem: `hls/preview-1/`, `hls/preview-2/`, `hls/preview-3/` (clean, ordered, debuggable)
- HLS URLs: `/hls/preview-{previewID}/playlist.m3u8`
- Logs: `previewID=1`, `previewID=2` (easy to correlate)

**Why not UUID/hash?** Simplicity over cleverness. Incremental IDs provide:
- Natural temporal ordering
- Clear debugging ("Preview 7 failed")
- Consistency with EventBus pattern
- Zero dependencies (no UUID library)

**Simplified Directory Structure:**
- No PID-based instance isolation (single instance assumption)
- Direct path: `{hlsPath}/preview-{id}/` instead of `{hlsPath}/{pid}/preview-{id}/`
- Simpler URLs and filesystem layout

**2. RemoteAddr for Client Identification:**

Each WebSocket connection has unique IP:port (RemoteAddr):
- Used for lookup in `StopPreview(remoteAddr)`
- Logged for debugging purposes (which client requested preview)
- PreviewID used for resource management (filesystem, maps)
- Separation of concerns: PreviewID = resources, RemoteAddr = client

**3. Session Pool (Not FSM):**

```go
type Manager struct {
    activePreviews map[uint64]*Session  // Key: PreviewID
    addrToPreview  map[string]uint64    // Key: remoteAddr → PreviewID
    nextPreviewID  atomic.Uint64        // Thread-safe counter
}
```

- One preview per client enforced automatically
- Auto-cleanup on WebSocket disconnect
- Three implicit states: spawning, ready, stopped
- State transitions managed by goroutines, no FSM tracking

**4. Hardcoded Timeouts:**

```go
playlistWaitTimeout = 30 * time.Second   // Wait for playlist.m3u8
processKillTimeout  = 5 * time.Second    // SIGTERM grace period
pollInterval        = 500 * time.Millisecond
stderrBufferSize    = 1024  // Circular buffer
```

No configuration - reasonable defaults that work for all scenarios.

**5. Simplified Directory Structure:**

```
hls/
├── preview-1/      ← First preview
│   ├── playlist.m3u8
│   └── segment*.ts
├── preview-2/      ← Second preview
└── preview-3/      ← Third preview
```

Direct structure without PID isolation. Running multiple instances simultaneously is not a supported use case (single OBS control).

**6. Filesystem Cleanup Strategy:**

- **Startup:** Remove entire HLS base directory and recreate empty
- **Runtime:** Immediate cleanup when preview stopped (delete preview-{id}/)
- **Shutdown:** Remove entire HLS base directory
- **Aggressive strategy:** Safe for temporary preview files

**7. Polling-based Playlist Detection:**

- Check `playlist.m3u8` existence every 500ms
- Timeout after 30 seconds if not created
- NO fsnotify - polling sufficient for this use case

**Public API:**

```go
// Constructor
func New(logger *logger.Logger, hlsBasePath string) (*Manager, error)

// Operations
func (m *Manager) StartPreview(req StartPreviewRequest) error
func (m *Manager) StopPreview(remoteAddr string) error
func (m *Manager) Shutdown() error
```

**StartPreviewRequest:**

```go
type StartPreviewRequest struct {
    RemoteAddr    string      // WebSocket remote address (IP:port)
    SourceURI     string      // Source URI to preview
    InputKind     string      // OBS input kind
    InputSettings interface{} // OBS input settings (optional)
    
    // Async callbacks (called from goroutine)
    OnReady func(hlsURL string)
    OnError func(errorMsg string)
}
```

**Process Lifecycle:**

1. **StartPreview:** Validates request, generates PreviewID, creates temp directory
2. **processPreview goroutine:** Spawns hls-generator, polls for playlist
3. **OnReady callback:** Invoked with HLS URL when playlist detected
4. **OnError callback:** Invoked on failure (spawn error, timeout, etc)
5. **StopPreview:** Kills process, cleans filesystem, removes from maps
6. **Auto-cleanup:** On WebSocket disconnect

**Error Handling:**

Immediate validation errors (synchronous):
- Browser source rejection (not supported)
- Binary not found (installation issue)

Async errors (via callback):
- Process spawn failure (invalid source URI)
- Timeout (30s, no playlist created)
- Internal panic (goroutine crash with recovery)

**Graceful Shutdown:**

Module shutdown uses WaitGroup pattern (consistent with WHEP handler) for parallel cleanup:

1. **Collect active sessions:** Copy all PreviewIDs while holding read lock
2. **Parallel termination:** Launch goroutine per preview with WaitGroup
   - Send SIGTERM (graceful shutdown)
   - Wait 5 seconds grace period
   - Send SIGKILL if still running
3. **Wait for completion:** WaitGroup.Wait() ensures all processes terminated
4. **Cleanup filesystem:** Remove entire instance directory
5. **Clear state:** Nil out maps (idempotent via sync.Once)

Performance: With N active previews, shutdown completes in ~5s (parallel) vs N×5s (sequential).

**Safety Features:**

- **Auto-cancel:** Starting new preview auto-stops old one (1 per client rule)
- **Idempotent:** StopPreview safe to call multiple times
- **Panic recovery:** Preview goroutines recover panics, invoke error callback
- **Circular buffer:** Stderr capture limited to 1KB (prevents memory growth)
- **Instance isolation:** Multiple backend instances use separate directories

**Integration with hls-generator:**

Binary discovery order:
1. Same directory as scenescheduler executable
2. System PATH

Configuration in `config.json`:

```json
{
  "webServer": {
    "hlsPath": "hls"
  }
}
```

The module SHALL attempt binary discovery during initialization:
- If found: Log info with path
- If not found: Log error, continue startup, fail preview requests gracefully

No panic or fatal error if binary not found - allows system to run without preview capability.

**Browser Source Handling:**

The SourcePreview module SHALL NOT handle browser_source types. Browser sources are rendered directly in the frontend. The frontend detects inputKind and handles browser sources locally by setting video element src.

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
| startPreview | Request HLS preview for a program source (payload: inputKind, uri, inputSettings) |
| stopPreview | Stop currently active source preview (no payload - tracked by remoteAddr) |

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
| previewReady | Source preview HLS stream ready (payload: hlsUrl) |
| previewError | Source preview generation failed (payload: error) |

### 8.4. Initial State Synchronization

**Problem:** Events fire before client connects. When frontend client connects, backend may have already established connections and activated features. These events fired before client was listening.

**Solution:** On WebSocket connection establishment, WebServer SHALL automatically:

1. Create response channel
2. Publish GetStatusRequested event with channel
3. Wait for StatusResponse from OBSClient
4. Send currentStatus message to client with obsConnected, obsVersion, virtualCamActive
5. Client updates status indicators immediately

This ensures accurate state display without race conditions or missed events.

### 8.5. Event Subscriptions and Message Routing

**EventBus Subscriptions**

WebServer SHALL subscribe to EventBus events for broadcasting:

- obs.system.connected → Broadcast obsConnected
- obs.system.disconnected → Broadcast obsDisconnected
- obs.virtualcam.started → Broadcast virtualCamStarted
- obs.virtualcam.stopped → Broadcast virtualCamStopped
- obs.program.changed → Broadcast programChanged
- Schedule file change → Broadcast scheduleChanged

**WebSocket Message Routing**

WebSocket handler SHALL route incoming messages via callbacks to WebServer:

- `startPreview` → `handleStartPreview(clientID, remoteAddr, payload)`
  - Parses payload (inputKind, uri, inputSettings)
  - Delegates to SourcePreview Manager with callbacks
  - Sends `previewReady` or `previewError` to requesting client only
- `stopPreview` → `handleStopPreview(clientID, remoteAddr)`
  - Delegates to SourcePreview Manager
  - Automatic cleanup (no response message)

**Client Tracking**

Preview requests are tracked by client IP (remoteAddr) to:
- Enforce one active preview per client
- Enable automatic cleanup on WebSocket disconnect
- Prevent resource leaks from abandoned connections

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
GET /hls/preview-{previewID}/playlist.m3u8  → HLS playlist
GET /hls/preview-{previewID}/*.ts           → HLS video segments
```

Implementation:
```go
hlsPath := webServerConfig.HlsPath // "hls" (relative path, validated)
hlsHandler := http.StripPrefix("/hls/", http.FileServer(http.Dir(hlsPath)))
http.Handle("/hls/", auth(hlsHandler))
```

Security:
- Handler serves files from hlsPath (validated as safe relative path)
- Protected by authentication middleware
- No directory traversal allowed (validated at config load)

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

**Last Updated:** 2025-10-17
**Document Version:** 1.5

**Changelog v1.5:**
- **WebSocket Preview Protocol (Section 8.3):**
  - Changed from `requestSourcePreview`/`stopSourcePreview` to `startPreview`/`stopPreview`
  - Removed `requestID` - tracking now by client remoteAddr (IP)
  - Changed from `sourcePreviewReady`/`sourcePreviewError` to `previewReady`/`previewError`
  - Simplified payload structure: `{inputKind, uri, inputSettings}` → `{hlsUrl}` or `{error}`
  - Added automatic cleanup on WebSocket disconnect
- **WebServer Preview Integration (Section 8.5):**
  - Preview routing via callbacks instead of EventBus
  - Callbacks: `OnStartPreview`, `OnStopPreview`, `OnClientDisconnected`
  - Client-specific responses (unicast) instead of broadcast
  - Implemented in `backend/webserver/sourcepreview.go`
- **WebSocket Internal Module:**
  - Added `startPreview` and `stopPreview` message routing
  - Extract remoteAddr from connection for preview tracking
  - Integrated in `backend/webserver/internal/websocket/handler.go`
- **File Organization (Section 5.1):**
  - Updated internal/ Integration Pattern from MUST to SHOULD
  - Added criteria for when to create bridge files vs when not to
  - Documented examples: `feed.go`, `switcher.go`, `sourcepreview.go` (needed)
  - Documented counter-examples: `websocket.go`, `whep.go` (not needed)
  - Clarified bridge files only for coordination logic, not pure delegation

**Changelog v1.4:**
- **Configuration Changes:**
  - Moved `hlsPath` from `paths` section to `webServer` section
  - Removed `paths.ffmpeg` (no longer used)
  - Added security validation for `hlsPath` (must be relative, no traversal)
- **SourcePreview Module Updates:**
  - Simplified directory structure (removed PID-based isolation)
  - Updated from `hls/{pid}/preview-{id}/` to `hls/preview-{id}/`
  - Simplified URLs from `/hls/{pid}/preview-{id}/` to `/hls/preview-{id}/`
  - Updated API signature: `New(logger, hlsBasePath string)` instead of `New(logger, *PathsConfig)`
- **WebSocket Preview Protocol (Section 8.3):**
  - Changed from `requestSourcePreview`/`stopSourcePreview` to `startPreview`/`stopPreview`
  - Removed `requestID` - tracking now by client remoteAddr (IP)
  - Changed from `sourcePreviewReady`/`sourcePreviewError` to `previewReady`/`previewError`
  - Simplified payload structure: `{inputKind, uri, inputSettings}` → `{hlsUrl}` or `{error}`
  - Added automatic cleanup on WebSocket disconnect
- **WebServer Preview Integration (Section 8.5):**
  - Preview routing via callbacks instead of EventBus
  - Callbacks: `OnStartPreview`, `OnStopPreview`, `OnClientDisconnected`
  - Client-specific responses (unicast) instead of broadcast
  - Implemented in `backend/webserver/sourcepreview.go`
- **Module Communication Rules (Section 5.5):**
  - Clarified EventBus DTO import exception for internal modules
  - Documented that internal modules MAY import pure DTOs (e.g., `eventbus.Program`)
  - Explicitly FORBID `eventbus.Publish()` and `eventbus.Subscribe()` in internal modules
  - Added comprehensive examples of allowed vs forbidden patterns
  - Documented rationale: DTOs are domain contracts, not infrastructure
- **HTTP Routes (Section 8.6):**
  - Added HLS file server handler with authentication
  - Updated implementation to use `webServerConfig.HlsPath`
  - Documented security measures (path validation, auth middleware)

**Changelog v1.3:**
- Added SourcePreview module specification (section 7.4)
- Added SourcePreview event contracts (section 6.7)
- Updated WebSocket protocol with source preview messages
- Updated WebServer event subscriptions and HTTP routes (section 8.6)
- Documented architecture decision criteria for top-level vs internal modules
