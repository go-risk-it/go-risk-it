package runner

// EventType identifies the kind of event flowing through the bus.
type EventType string

// Event is the interface all bus events must satisfy.
type Event interface {
	Type() EventType
}

// HandlerFunc is invoked when a matching event is emitted.
type HandlerFunc func(bus *Bus, event Event)

// Bus is a synchronous, in-process event dispatcher.
// One bus per game, single goroutine — no concurrency.
type Bus struct {
	handlers map[EventType][]HandlerFunc
	emitted  []Event
	capture  bool
	stopped  bool
}

// NewBus creates a production bus (no capture).
func NewBus() *Bus {
	return &Bus{handlers: make(map[EventType][]HandlerFunc)}
}

// NewTestBus creates a bus that captures every emitted event.
func NewTestBus() *Bus {
	return &Bus{
		handlers: make(map[EventType][]HandlerFunc),
		capture:  true,
	}
}

// On registers a handler for the given event type.
// Handlers are called in registration order.
func (b *Bus) On(t EventType, h HandlerFunc) {
	b.handlers[t] = append(b.handlers[t], h)
}

// Emit dispatches an event to all registered handlers.
// No-op if the bus has been stopped.
// Re-entrant: a handler may call Emit, which dispatches immediately (depth-first).
func (b *Bus) Emit(e Event) {
	if b.stopped {
		return
	}

	if b.capture {
		b.emitted = append(b.emitted, e)
	}

	for _, h := range b.handlers[e.Type()] {
		if b.stopped {
			return
		}

		h(b, e)
	}
}

// Stop prevents all subsequent Emit calls from dispatching.
func (b *Bus) Stop() {
	b.stopped = true
}

// Stopped returns whether Stop() has been called.
func (b *Bus) Stopped() bool {
	return b.stopped
}

// Emitted returns all captured events. Panics on non-test bus.
func (b *Bus) Emitted() []Event {
	if !b.capture {
		panic("Emitted() called on non-test bus; use NewTestBus()")
	}

	return b.emitted
}

// EmittedOfType returns captured events matching the given type.
// Panics on non-test bus.
func (b *Bus) EmittedOfType(t EventType) []Event {
	if !b.capture {
		panic("EmittedOfType() called on non-test bus; use NewTestBus()")
	}

	var out []Event

	for _, e := range b.emitted {
		if e.Type() == t {
			out = append(out, e)
		}
	}

	return out
}
