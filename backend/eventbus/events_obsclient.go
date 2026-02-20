// backend/eventbus/events_obs.go
package eventbus

import "time"

// =============================================================================
// OBS System & Lifecycle Events
// =============================================================================

// OBSConnected is emitted when a connection to OBS WebSocket is established and identified.
// It carries the discovered version information.
type OBSConnected struct {
    OBSVersion    string    `json:"obsVersion"`
    Timestamp     time.Time `json:"timestamp"`
}

func (e OBSConnected) GetTopic() string { return "obs.system.connected" }

// OBSDisconnected is emitted when the connection to OBS WebSocket is lost.
// It carries the error that caused the disconnection.
type OBSDisconnected struct {
    Error     error     `json:"error"`
    Timestamp time.Time `json:"timestamp"`
}

func (e OBSDisconnected) GetTopic() string { return "obs.system.disconnected" }

// =============================================================================
// Virtual Camera Events
// =============================================================================

// OBSVirtualCamStarted is published when the OBS virtual camera is activated.
type OBSVirtualCamStarted struct {
    Timestamp time.Time
}

func (e OBSVirtualCamStarted) GetTopic() string { return "obs.virtualcam.started" }

// OBSVirtualCamStopped is published when the OBS virtual camera is deactivated.
type OBSVirtualCamStopped struct {
    Timestamp time.Time
}

func (e OBSVirtualCamStopped) GetTopic() string { return "obs.virtualcam.stopped" }

// =============================================================================
// Scene, Streaming & Recording Events
// =============================================================================

// OBSSceneChanged is emitted when the active program scene changes.
type OBSSceneChanged struct {
    SceneName string    `json:"sceneName"`
    Timestamp time.Time `json:"timestamp"`
}

func (e OBSSceneChanged) GetTopic() string { return "obs.scene.changed" }

// OBSStreamStateChanged is emitted when streaming starts or stops.
type OBSStreamStateChanged struct {
    IsStreaming bool      `json:"isStreaming"`
    OutputName  string    `json:"outputName,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
}

func (e OBSStreamStateChanged) GetTopic() string { return "obs.stream.state.changed" }

// OBSRecordStateChanged is emitted when recording starts or stops.
type OBSRecordStateChanged struct {
    IsRecording bool      `json:"isRecording"`
    OutputName  string    `json:"outputName,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
}

func (e OBSRecordStateChanged) GetTopic() string { return "obs.record.state.changed" }

// =============================================================================
// Program Change Events
// =============================================================================

// OBSProgramChanged is emitted when OBSClient successfully completes a program switch.
// This event confirms that the change has been applied in OBS (ON AIR).
// It is NOT emitted if the target state is the same as the current state (idempotent).
type OBSProgramChanged struct {
    Timestamp       time.Time    `json:"timestamp"`
    PreviousProgram *Program     `json:"previousProgram,omitempty"`
    CurrentProgram  *Program     `json:"currentProgram,omitempty"`
    SeekOffsetMs    int64        `json:"seekOffsetMs,omitempty"`
}

func (e OBSProgramChanged) GetTopic() string { return "obs.program.changed" }