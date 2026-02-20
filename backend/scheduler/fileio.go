// backend/scheduler/fileio.go
//
// Schedule file I/O operations and client communication.
//
// Contents:
// - Schedule File Loading
// - Schedule File Writing
// - Client Communication

package scheduler

import (
	"encoding/json"
	"fmt"
	"os"

	"scenescheduler/backend/eventbus"
)

// ============================================================================
// SCHEDULE FILE LOADING
// ============================================================================

// loadScheduleFromFile reads and parses the schedule file from disk.
// Returns the parsed schedule on success, or an error if reading/parsing fails.
func (s *Scheduler) loadScheduleFromFile() (*Schedule, error) {
	filePath := s.paths.Schedule
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schedule file '%s': %w", filePath, err)
	}

	var schedule Schedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, fmt.Errorf("failed to parse schedule JSON from '%s': %w", filePath, err)
	}

	if schedule.Programs == nil {
		schedule.Programs = make([]ScheduledProgram, 0)
	}

	return &schedule, nil
}

// ============================================================================
// SCHEDULE FILE WRITING
// ============================================================================

// commitSchedule saves a new schedule received from the frontend.
// It validates, writes to disk, and lets the FileWatcher trigger the reload.
//
// The FileWatcher will detect the file change and call reloadSchedule(),
// which will trigger evaluation. This prevents double-evaluation.
func (s *Scheduler) commitSchedule(clientID string, payload json.RawMessage) {
	s.logger.Info("Committing new schedule to file", "clientID", clientID)

	// Parse payload to validate it's valid JSON
	var scheduleData interface{}
	if err := json.Unmarshal(payload, &scheduleData); err != nil {
		s.logger.Error("Failed to parse schedule payload", "error", err, "clientID", clientID)
		s.sendCommitError(clientID, "Invalid JSON format")
		return
	}

	// Pretty-print JSON for human readability
	prettyJSON, err := json.MarshalIndent(scheduleData, "", "  ")
	if err != nil {
		s.logger.Error("Failed to marshal schedule for saving", "error", err, "clientID", clientID)
		s.sendCommitError(clientID, "Failed to format schedule")
		return
	}

	// Write to disk
	err = os.WriteFile(s.paths.Schedule, prettyJSON, 0644)
	if err != nil {
		s.logger.Error("Failed to write schedule.json file", "error", err, "clientID", clientID)
		s.sendCommitError(clientID, "Failed to write file")
		return
	}

	s.logger.Info("Successfully wrote new schedule to file", "path", s.paths.Schedule)

	// Send success response to client
	s.sendCommitSuccess(clientID)

	// NOTE: Do NOT call evaluateAndSwitch() here.
	// The FileWatcher will detect the change and trigger reloadSchedule(),
	// which will then call evaluateAndSwitch().
}

// ============================================================================
// CLIENT COMMUNICATION
// ============================================================================

// getSchedule sends the currently loaded schedule to a client via WebSocket.
func (s *Scheduler) getSchedule(clientID string) {
	s.logger.Debug("Sending current schedule to client", "clientID", clientID)

	s.mu.RLock()
	currentSchedule := s.schedule
	s.mu.RUnlock()

	if currentSchedule == nil {
		s.logger.Warn("No schedule loaded in memory to send", "clientID", clientID)
		return
	}

	eventbus.Publish(s.bus, eventbus.WebSocketSendMessageToClient{
		ClientID:    clientID,
		MessageType: "currentSchedule",
		Payload:     currentSchedule,
	})
}

// sendCommitSuccess sends a success response to the client after committing schedule.
func (s *Scheduler) sendCommitSuccess(clientID string) {
	eventbus.Publish(s.bus, eventbus.WebSocketSendMessageToClient{
		ClientID:    clientID,
		MessageType: "commitSuccess",
		Payload:     map[string]interface{}{},
	})
}

// sendCommitError sends an error response to the client if commit fails.
func (s *Scheduler) sendCommitError(clientID string, message string) {
	eventbus.Publish(s.bus, eventbus.WebSocketSendMessageToClient{
		ClientID:    clientID,
		MessageType: "commitError",
		Payload: map[string]interface{}{
			"message": message,
		},
	})
}