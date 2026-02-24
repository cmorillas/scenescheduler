// backend/eventbus/eventbus.go
package eventbus

import (
    "errors"
    "fmt"
    "os"
    "sync"
    "sync/atomic"
    "time"
)

// This file provides a thread-safe, generic publish-subscribe event system.
// It allows different parts of the application to communicate in a decoupled manner.
//
// Order of sections:
// 1) Core Types & Constructor
// 2) Public API (Subscribe, Publish)
// 3) Lifecycle Management
// 4) Internal Methods
// 5) Introspection & Helpers
// 6) Logging Methods

// ============================================================================
// CORE TYPES & CONSTRUCTOR
// ============================================================================

// Event is the interface that all events must implement.
type Event interface {
    GetTopic() string
}

// UnsubscribeFunc is a function that cancels a subscription when called.
type UnsubscribeFunc func()

// subscription holds the handler and the name of a subscriber for logging.
type subscription struct {
    handler func(any)
    name    string
}

// EventBus provides a thread-safe publish-subscribe system.
type EventBus struct {
    mu          sync.RWMutex
    subscribers map[string]map[uint64]subscription // map[topic]map[subID]subscription
    nextID      atomic.Uint64
    closed      bool
}

// New creates and returns a new EventBus instance.
func New() *EventBus {
    bus := &EventBus{
        subscribers: make(map[string]map[uint64]subscription),
    }
    bus.logInfo("new EventBus created")
    return bus
}

// ============================================================================
// PUBLIC API
// ============================================================================

// Subscribe adds a typed handler function for events of type T
// and returns an UnsubscribeFunc to cancel the subscription.
// The topic is automatically determined from the event type using GetTopic().
// The subscriberName is used for logging purposes.
func Subscribe[T Event](bus *EventBus, subscriberName string, handler func(T)) (UnsubscribeFunc, error) {
    if handler == nil {
        bus.logError("subscribe failed: handler cannot be nil (subscriber: %s)", subscriberName)
        return nil, errors.New("handler cannot be nil")
    }
    if subscriberName == "" {
        subscriberName = "unknown" // Prevent empty log messages
    }

    var zero T
    topic := zero.GetTopic()

    bus.mu.Lock()
    defer bus.mu.Unlock()

    if bus.closed {
        bus.logWarn("subscribe failed: bus is closed for topic '%s' (subscriber: %s)", topic, subscriberName)
        return nil, errors.New("bus is closed")
    }

    id := bus.nextID.Add(1)

    // Wrapper that performs the type assertion internally.
    wrapper := func(event any) {
        if typedEvent, ok := event.(T); ok {
            handler(typedEvent)
        } else {
            bus.logWarn("event type mismatch for topic '%s': expected %T, got %T",
                topic, zero, event)
        }
    }

    if bus.subscribers[topic] == nil {
        bus.subscribers[topic] = make(map[uint64]subscription)
    }

    bus.subscribers[topic][id] = subscription{
        handler: wrapper,
        name:    subscriberName,
    }

    bus.logInfo("'%s' subscribed to topic '%s' with id %d (total handlers: %d)",
        subscriberName, topic, id, len(bus.subscribers[topic]))

    // Create and return the cancellation function.
    unsubscribe := func() {
        bus.unsubscribe(topic, id)
    }

    return unsubscribe, nil
}

// Publish sends an event to all subscribers of the event's topic.
// The topic is automatically determined from the event using GetTopic().
func Publish[T Event](bus *EventBus, event T) {
    topic := event.GetTopic()

    bus.mu.RLock()
    closed := bus.closed
    topicHandlers := bus.subscribers[topic]
    handlersCopy := make([]subscription, 0, len(topicHandlers))
    for _, sub := range topicHandlers {
        handlersCopy = append(handlersCopy, sub)
    }
    bus.mu.RUnlock()

    if closed {
        bus.logWarn("publish failed: bus is closed for topic '%s'", topic)
        return
    }
    if len(handlersCopy) == 0 {
        return // Publishing to a topic with no subscribers is normal, so no log needed.
    }

    for _, sub := range handlersCopy {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    bus.logError("handler from '%s' panicked for topic '%s': %v", sub.name, topic, r)
                }
            }()
            sub.handler(event)
        }()
    }
}

// ============================================================================
// LIFECYCLE MANAGEMENT
// ============================================================================

// Close shuts down the EventBus and removes all subscribers.
func (bus *EventBus) Close() {
    bus.mu.Lock()
    defer bus.mu.Unlock()

    if bus.closed {
        bus.logWarn("close called on an already closed bus")
        return
    }

    totalSubscribers := 0
    totalTopics := len(bus.subscribers)
    for _, handlers := range bus.subscribers {
        totalSubscribers += len(handlers)
    }

    bus.closed = true
    bus.subscribers = nil // Release memory
    bus.logInfo("bus closed, removed %d subscribers across %d topics", totalSubscribers, totalTopics)
}

// ============================================================================
// INTERNAL METHODS
// ============================================================================

// unsubscribe is a private, unexported method that performs the actual removal.
func (bus *EventBus) unsubscribe(topic string, id uint64) {
    bus.mu.Lock()
    defer bus.mu.Unlock()

    if bus.closed {
        return
    }

    topicHandlers, ok := bus.subscribers[topic]
    if !ok {
        return // Topic does not exist, nothing to do.
    }

    // Check if the specific subscription still exists.
    if sub, ok := topicHandlers[id]; ok {
        delete(topicHandlers, id)
        bus.logInfo("'%s' (id %d) unsubscribed from topic '%s'", sub.name, id, topic)

        // If the topic becomes empty, remove it to clean up.
        if len(topicHandlers) == 0 {
            delete(bus.subscribers, topic)
            bus.logInfo("removed empty topic '%s'", topic)
        }
    }
}

// ============================================================================
// INTROSPECTION & HELPERS
// ============================================================================

// IsClosed returns whether the EventBus is closed.
func (bus *EventBus) IsClosed() bool {
    bus.mu.RLock()
    defer bus.mu.RUnlock()
    return bus.closed
}

// ============================================================================
// LOGGING METHODS
// ============================================================================

func (bus *EventBus) logInfo(msg string, args ...any) {
    fmt.Fprintf(os.Stdout, "[%s] INFO [eventbus]: %s\n",
        time.Now().Format("15:04:05"), fmt.Sprintf(msg, args...))
}

func (bus *EventBus) logWarn(msg string, args ...any) {
    fmt.Fprintf(os.Stderr, "[%s] WARN [eventbus]: %s\n",
        time.Now().Format("15:04:05"), fmt.Sprintf(msg, args...))
}

func (bus *EventBus) logError(msg string, args ...any) {
    fmt.Fprintf(os.Stderr, "[%s] ERROR [eventbus]: %s\n",
        time.Now().Format("15:04:05"), fmt.Sprintf(msg, args...))
}

