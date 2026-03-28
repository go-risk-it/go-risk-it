package events

import (
	"context"
	"time"
)

// Event is the base interface for all domain events emitted through the EventBus.
// Each event type carries domain-specific data and supports structured serialization
// via ToRecord() for observability consumers (Loki, metrics, headline detection).
// Domain-specific interfaces (e.g. game.GameEvent) embed Event and add scope IDs.
type Event interface {
	EventType() string
	EventTimestamp() time.Time
	ToRecord() map[string]any
}

// Handler processes a single event. Handlers are invoked by the bus in separate
// goroutines with detached contexts. Panics are recovered by safego.Go.
type Handler func(ctx context.Context, event Event)
