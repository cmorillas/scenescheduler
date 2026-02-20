// backend/eventbus/types_shared.go
package eventbus

import "time"

// =============================================================================
// Shared Domain Types
// =============================================================================

// Program represents a program in the schedule with all its properties.
// This type is used across multiple events (scheduler, OBS client, etc.)
// to ensure consistency and avoid duplication.
type Program struct {
    ID            string      `json:"id"`
    Title         string      `json:"title"`
    SourceName    string      `json:"sourceName,omitempty"`
    SceneName     string      `json:"sceneName,omitempty"`
    InputKind     string      `json:"inputKind,omitempty"`
    URI           string      `json:"uri,omitempty"`
    InputSettings interface{} `json:"inputSettings,omitempty"`
    Transform     interface{} `json:"transform,omitempty"`
    Start         time.Time   `json:"start,omitempty"`
    End           time.Time   `json:"end,omitempty"`
}