package events

import (
	"context"
	"sync"
)

// TestBus is a synchronous Bus implementation for unit tests. It captures all
// emitted events in order and dispatches to handlers in the calling goroutine.
// No linked spans, no goroutines, no detachContext. Dispatch order: OnAll before
// OnType (same as production bus). Exported because test code in other packages
// needs to construct it.
type TestBus struct {
	mu     sync.Mutex
	events []Event
	allH   []Handler
	typedH map[string][]Handler
}

var _ Bus = (*TestBus)(nil)

// NewTestBus creates a new TestBus ready for use in tests.
func NewTestBus() *TestBus {
	return &TestBus{
		typedH: make(map[string][]Handler),
	}
}

// Emit dispatches event to all matching handlers synchronously in the calling
// goroutine, then appends the event to the capture slice. Panics if event is nil.
func (b *TestBus) Emit(ctx context.Context, event Event) {
	if event == nil {
		panic("events: Emit called with nil event")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, h := range b.allH {
		h(ctx, event)
	}

	if typed, ok := b.typedH[event.EventType()]; ok {
		for _, h := range typed {
			h(ctx, event)
		}
	}

	b.events = append(b.events, event)
}

// OnAll registers a handler that receives every emitted event.
// Panics if handler is nil.
func (b *TestBus) OnAll(handler Handler) {
	if handler == nil {
		panic("events: OnAll called with nil handler")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.allH = append(b.allH, handler)
}

// OnType registers a handler that receives only events matching the given type.
// Panics if handler is nil or eventType is empty.
func (b *TestBus) OnType(eventType string, handler Handler) {
	if eventType == "" {
		panic("events: OnType called with empty eventType")
	}

	if handler == nil {
		panic("events: OnType called with nil handler")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.typedH[eventType] = append(b.typedH[eventType], handler)
}

// Close is a no-op for TestBus. Always returns nil.
func (b *TestBus) Close(_ context.Context) error {
	return nil
}

// Events returns all captured events in emission order.
func (b *TestBus) Events() []Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]Event, len(b.events))
	copy(result, b.events)

	return result
}

// Reset clears all captured events.
func (b *TestBus) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = nil
}

// EventsOfType is a generic helper that returns all captured events matching
// the concrete type E. It is a package-level function because Go does not allow
// type parameters on methods.
func EventsOfType[E Event](bus *TestBus) []E {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	var result []E

	for _, event := range bus.events {
		if typed, ok := event.(E); ok {
			result = append(result, typed)
		}
	}

	return result
}
