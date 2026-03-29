package bus

import (
	"context"
	"log/slog"
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

// ScopedEvent is an optional interface for events that carry a domain scope ID
// (e.g., gameID for game events, lobbyID for lobby events). The event logger uses
// a single ScopedEvent type assertion instead of switching on anonymous interfaces.
type ScopedEvent interface {
	ScopeAttrs() []slog.Attr
}

// Handler processes a single event. The bus invokes handlers sequentially within a
// single goroutine per event (OnAll before OnType), each with a detached context
// carrying a linked span. Panics are recovered per handler — a panicking handler does
// not prevent subsequent handlers from executing.
type Handler func(ctx context.Context, event Event)
