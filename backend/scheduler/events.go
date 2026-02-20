// backend/scheduler/events.go
//
// EventBus subscription management and event handlers.
//
// Contents:
// - Subscription Setup
// - Event Handlers
// - Unsubscribe Cleanup

package scheduler

import (
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// SUBSCRIPTION SETUP
// ============================================================================

// subscribeToEvents sets up all event bus subscriptions for the Scheduler.
// This is called from New() to ensure the module is ready immediately.
func (s *Scheduler) subscribeToEvents() {
	s.logger.Debug("Subscribing to application events")

	unsub1, err1 := eventbus.Subscribe(s.bus, "Scheduler", s.handleGetScheduleRequest)
	s.addUnsubscriber(unsub1, err1, "GetScheduleRequested")

	unsub2, err2 := eventbus.Subscribe(s.bus, "Scheduler", s.handleCommitScheduleRequest)
	s.addUnsubscriber(unsub2, err2, "CommitScheduleRequested")
}

// addUnsubscriber is a helper to reduce boilerplate in the subscription process.
func (s *Scheduler) addUnsubscriber(unsub eventbus.UnsubscribeFunc, err error, topic string) {
	if err != nil {
		s.logger.Error("Failed to subscribe to event", "topic", topic, "error", err)
	} else {
		s.unsubscribeFuncs = append(s.unsubscribeFuncs, unsub)
	}
}

// unsubscribeAllEvents cleans up all event subscriptions.
// Called during shutdown to prevent memory leaks.
func (s *Scheduler) unsubscribeAllEvents() {
	s.logger.Debug("Unsubscribing from all events")
	for _, unsub := range s.unsubscribeFuncs {
		if unsub != nil {
			unsub()
		}
	}
	s.unsubscribeFuncs = nil
}

// ============================================================================
// EVENT HANDLERS
// ============================================================================

// handleGetScheduleRequest receives the event and calls the corresponding method.
//
// Topic: websocket.command.getSchedule
func (s *Scheduler) handleGetScheduleRequest(event eventbus.GetScheduleRequested) {
	s.logger.Debug("Handling GetScheduleRequested event", "clientID", event.ClientID)
	s.getSchedule(event.ClientID)
}

// handleCommitScheduleRequest receives the event and calls the corresponding method.
//
// Topic: websocket.command.commitSchedule
func (s *Scheduler) handleCommitScheduleRequest(event eventbus.CommitScheduleRequested) {
	s.logger.Debug("Handling CommitScheduleRequested event", "clientID", event.ClientID)
	s.commitSchedule(event.ClientID, event.Payload)
}
