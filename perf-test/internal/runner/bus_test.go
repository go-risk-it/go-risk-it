package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEvent is a minimal Event implementation for tests.
type testEvent struct {
	eventType EventType
	label     string // distinguishes instances
}

func (e testEvent) Type() EventType { return e.eventType }

func TestBus_Emit_CallsRegisteredHandler(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var called int
	var received Event

	bus.On(EventType("x"), func(b *Bus, e Event) {
		called++
		received = e
	})

	evt := testEvent{eventType: "x", label: "first"}
	bus.Emit(evt)

	assert.Equal(t, 1, called)
	assert.Equal(t, evt, received)
}

func TestBus_Emit_MultipleHandlers(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var order []string

	bus.On(EventType("x"), func(b *Bus, e Event) { order = append(order, "h1") })
	bus.On(EventType("x"), func(b *Bus, e Event) { order = append(order, "h2") })

	bus.Emit(testEvent{eventType: "x"})

	assert.Equal(t, []string{"h1", "h2"}, order)
}

func TestBus_Emit_NoHandler(t *testing.T) {
	t.Parallel()

	bus := NewBus()
	// Should not panic.
	bus.Emit(testEvent{eventType: "x"})
}

func TestBus_Emit_ReentrantEmit(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var log []string

	bus.On(EventType("x"), func(b *Bus, e Event) {
		log = append(log, "x-start")
		b.Emit(testEvent{eventType: "y"})
		log = append(log, "x-end")
	})
	bus.On(EventType("y"), func(b *Bus, e Event) {
		log = append(log, "y")
	})

	bus.Emit(testEvent{eventType: "x"})

	// Depth-first: y handler runs between x-start and x-end.
	assert.Equal(t, []string{"x-start", "y", "x-end"}, log)
}

func TestBus_Stop_PreventsSubsequentEmit(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var called bool

	bus.On(EventType("x"), func(b *Bus, e Event) { called = true })
	bus.Stop()
	bus.Emit(testEvent{eventType: "x"})

	assert.False(t, called)
	assert.True(t, bus.Stopped())
}

func TestBus_Stop_DuringHandler(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var log []string

	bus.On(EventType("x"), func(b *Bus, e Event) {
		log = append(log, "h1")
		b.Stop()
	})
	bus.On(EventType("x"), func(b *Bus, e Event) {
		log = append(log, "h2")
	})

	bus.Emit(testEvent{eventType: "x"})

	// h1 runs and stops; h2 should NOT run.
	assert.Equal(t, []string{"h1"}, log)
}

func TestBus_TestMode_CapturesEmitted(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	bus.On(EventType("x"), func(b *Bus, e Event) {})

	evtX := testEvent{eventType: "x", label: "a"}
	evtY := testEvent{eventType: "y", label: "b"}
	bus.Emit(evtX)
	bus.Emit(evtY)

	assert.Equal(t, []Event{evtX, evtY}, bus.Emitted())
}

func TestBus_TestMode_EmittedOfType(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()

	bus.Emit(testEvent{eventType: "x", label: "1"})
	bus.Emit(testEvent{eventType: "y", label: "2"})
	bus.Emit(testEvent{eventType: "x", label: "3"})

	xEvents := bus.EmittedOfType(EventType("x"))
	assert.Len(t, xEvents, 2)
	assert.Equal(t, "1", xEvents[0].(testEvent).label)
	assert.Equal(t, "3", xEvents[1].(testEvent).label)
}

func TestBus_Emitted_PanicsInNonTestMode(t *testing.T) {
	t.Parallel()

	bus := NewBus()
	require.Panics(t, func() { bus.Emitted() })
	require.Panics(t, func() { bus.EmittedOfType(EventType("x")) })
}

func TestBus_Emit_HandlerPanic_Recovers(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var order []string

	// First handler panics.
	bus.On(EventType("x"), func(b *Bus, e Event) {
		order = append(order, "h1-before-panic")
		panic("handler blew up")
	})
	// Second handler must still fire.
	bus.On(EventType("x"), func(b *Bus, e Event) {
		order = append(order, "h2")
	})

	// Emit must not propagate the panic.
	require.NotPanics(t, func() {
		bus.Emit(testEvent{eventType: "x"})
	})

	assert.Equal(t, []string{"h1-before-panic", "h2"}, order)
}

func TestBus_Emit_HandlerPanic_BusStillFunctional(t *testing.T) {
	t.Parallel()

	bus := NewTestBus()
	var log []string

	// Register a panicking handler on event type "a".
	bus.On(EventType("a"), func(b *Bus, e Event) {
		panic("boom")
	})

	// Register a normal handler on event type "b".
	bus.On(EventType("b"), func(b *Bus, e Event) {
		log = append(log, "b-handled")
	})

	// First emit: the panic is recovered.
	bus.Emit(testEvent{eventType: "a"})

	// Bus must still be functional — not stopped, not broken.
	assert.False(t, bus.Stopped())

	// Second emit: different event type works normally.
	bus.Emit(testEvent{eventType: "b"})
	assert.Equal(t, []string{"b-handled"}, log)

	// Third emit: even the panicking handler's event type can be emitted again.
	require.NotPanics(t, func() {
		bus.Emit(testEvent{eventType: "a"})
	})
}
