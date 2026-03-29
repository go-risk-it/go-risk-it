package bus_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	lobbyclx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

// testEvent is a minimal GameEvent for testing.
type testEvent struct {
	gameID    int64
	eventType string
}

func (e *testEvent) EventType() string         { return e.eventType }
func (e *testEvent) GameID() int64             { return e.gameID }
func (e *testEvent) EventTimestamp() time.Time { return time.Now() }
func (e *testEvent) ToRecord() map[string]any  { return map[string]any{"type": e.eventType} }

func newTestEvent(gameID int64, eventType string) *testEvent {
	return &testEvent{gameID: gameID, eventType: eventType}
}

func TestBus_EmitSingleHandler(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()
	received := make(chan eventbus.Event, 1)

	bus.OnAll(func(_ context.Context, event eventbus.Event) {
		received <- event
	})

	evt := newTestEvent(42, "test.event")
	bus.Emit(context.Background(), evt)

	select {
	case got := <-received:
		require.Equal(t, evt, got)
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not receive event within timeout")
	}
}

func TestBus_EmitMultipleHandlers(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	const numHandlers = 5
	var count atomic.Int32
	done := make(chan struct{})

	for range numHandlers {
		bus.OnAll(func(_ context.Context, _ eventbus.Event) {
			if count.Add(1) == numHandlers {
				close(done)
			}
		})
	}

	bus.Emit(context.Background(), newTestEvent(1, "test.event"))

	select {
	case <-done:
		require.Equal(t, int32(numHandlers), count.Load())
	case <-time.After(2 * time.Second):
		t.Fatal("not all handlers were called within timeout")
	}
}

func TestBus_OnTypeFiltering(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	matchCh := make(chan eventbus.Event, 1)
	noMatchCh := make(chan struct{}, 1)

	bus.OnType("move.executed", func(_ context.Context, event eventbus.Event) {
		matchCh <- event
	})

	bus.OnType("game.created", func(_ context.Context, _ eventbus.Event) {
		noMatchCh <- struct{}{}
	})

	evt := newTestEvent(1, "move.executed")
	bus.Emit(context.Background(), evt)

	select {
	case got := <-matchCh:
		require.Equal(t, evt, got)
	case <-time.After(2 * time.Second):
		t.Fatal("matching handler did not receive event")
	}

	// Give non-matching handler time to (incorrectly) fire.
	select {
	case <-noMatchCh:
		t.Fatal("non-matching handler should not have been called")
	case <-time.After(100 * time.Millisecond):
		// Expected: no call.
	}
}

func TestBus_OnAllAndOnTypeCoexistence(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	allCh := make(chan eventbus.Event, 1)
	typedCh := make(chan eventbus.Event, 1)

	bus.OnAll(func(_ context.Context, event eventbus.Event) {
		allCh <- event
	})
	bus.OnType("move.executed", func(_ context.Context, event eventbus.Event) {
		typedCh <- event
	})

	evt := newTestEvent(1, "move.executed")
	bus.Emit(context.Background(), evt)

	select {
	case got := <-allCh:
		require.Equal(t, evt, got)
	case <-time.After(2 * time.Second):
		t.Fatal("OnAll handler did not receive event")
	}

	select {
	case got := <-typedCh:
		require.Equal(t, evt, got)
	case <-time.After(2 * time.Second):
		t.Fatal("OnType handler did not receive event")
	}
}

func TestBus_ConcurrentEmit(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	const numEmits = 100
	var count atomic.Int32
	done := make(chan struct{})

	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		if count.Add(1) == numEmits {
			close(done)
		}
	})

	var waitGroup sync.WaitGroup
	waitGroup.Add(numEmits)

	for i := range numEmits {
		go func() {
			defer waitGroup.Done()
			bus.Emit(context.Background(), newTestEvent(int64(i), "test.event"))
		}()
	}

	waitGroup.Wait()

	select {
	case <-done:
		require.Equal(t, int32(numEmits), count.Load())
	case <-time.After(5 * time.Second):
		t.Fatalf("only %d of %d handlers completed", count.Load(), numEmits)
	}
}

func TestBus_PanicInHandlerRecovery(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()
	survived := make(chan struct{}, 1)

	// First handler panics.
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		panic("boom")
	})

	// Second handler should still execute.
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		survived <- struct{}{}
	})

	bus.Emit(context.Background(), newTestEvent(1, "test.event"))

	select {
	case <-survived:
		// Success: panic in first handler did not prevent second handler.
	case <-time.After(2 * time.Second):
		t.Fatal("surviving handler was not called after peer panic")
	}
}

func TestBus_CloseWaitsForInFlight(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()
	started := make(chan struct{})
	proceed := make(chan struct{})
	handlerDone := make(chan struct{})

	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		close(started)
		<-proceed
		close(handlerDone)
	})

	bus.Emit(context.Background(), newTestEvent(1, "test.event"))

	// Wait for handler to start.
	<-started

	// Close in a goroutine — it should block until handler completes.
	closeDone := make(chan error, 1)
	go func() {
		closeDone <- bus.Close(context.Background())
	}()

	// Close should not return yet since handler is blocked.
	select {
	case <-closeDone:
		t.Fatal("Close returned before in-flight handler completed")
	case <-time.After(100 * time.Millisecond):
		// Expected.
	}

	// Unblock handler.
	close(proceed)
	<-handlerDone

	// Now Close should return.
	select {
	case err := <-closeDone:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not return after handlers finished")
	}
}

func TestBus_CloseTimeout(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()
	started := make(chan struct{})

	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		close(started)
		// Block forever — simulates a stuck handler.
		select {}
	})

	bus.Emit(context.Background(), newTestEvent(1, "test.event"))
	<-started

	closeCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := bus.Close(closeCtx)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestBus_EmitAfterClose(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	called := make(chan struct{}, 1)
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		called <- struct{}{}
	})

	err := bus.Close(context.Background())
	require.NoError(t, err)

	// Emit after close should be a no-op.
	bus.Emit(context.Background(), newTestEvent(1, "test.event"))

	select {
	case <-called:
		t.Fatal("handler was called after Close")
	case <-time.After(100 * time.Millisecond):
		// Expected: no dispatch.
	}
}

func TestBus_CloseIdempotent(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	err1 := bus.Close(context.Background())
	require.NoError(t, err1)

	err2 := bus.Close(context.Background())
	require.NoError(t, err2)
}

func TestDetachContext_GameContext(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-123")
	gameCtx := gamectx.WithGameID(userCtx, 42)

	evt := newTestEvent(42, "test.event")
	detached, cancel := eventbus.DetachContextForTest(gameCtx, evt, 5*time.Second)
	defer cancel()

	// Must preserve GameID and UserID through the detached context.
	gameContext, ok := detached.(gamectx.GameContext)
	require.True(t, ok, "expected GameContext, got %T", detached)
	require.Equal(t, int64(42), gameContext.GameID())
	require.Equal(t, "user-123", gameContext.UserID())

	// Must have a deadline from the timeout.
	deadline, hasDeadline := detached.Deadline()
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)
}

func TestDetachContext_LobbyContext(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-456")
	lobbyCtx := lobbyclx.WithLobbyID(userCtx, 77)

	evt := newTestEvent(77, "test.event")
	detached, cancel := eventbus.DetachContextForTest(lobbyCtx, evt, 5*time.Second)
	defer cancel()

	// Must preserve LobbyID and UserID through the detached context.
	lobbyContext, ok := detached.(lobbyclx.LobbyContext)
	require.True(t, ok, "expected LobbyContext, got %T", detached)
	require.Equal(t, int64(77), lobbyContext.LobbyID())
	require.Equal(t, "user-456", lobbyContext.UserID())

	// Must have a deadline from the timeout.
	deadline, hasDeadline := detached.Deadline()
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)
}

func TestDetachContext_PlainContext(t *testing.T) {
	t.Parallel()

	plain := context.Background()

	evt := newTestEvent(1, "test.event")
	detached, cancel := eventbus.DetachContextForTest(plain, evt, 5*time.Second)
	defer cancel()

	// Should have a deadline.
	deadline, hasDeadline := detached.Deadline()
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)

	// Should NOT be a GameContext.
	_, isGame := detached.(gamectx.GameContext)
	require.False(t, isGame, "plain context should not become GameContext")

	// Should NOT be a LobbyContext.
	_, isLobby := detached.(lobbyclx.LobbyContext)
	require.False(t, isLobby, "plain context should not become LobbyContext")
}

//nolint:paralleltest // swaps global TracerProvider
func TestStartLinkedSpan_ValidTrigger(
	t *testing.T,
) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	// Set this test's TracerProvider as the global so startLinkedSpan picks it up.
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	// Create a source span to link from.
	tracer := tracerProvider.Tracer("test")
	sourceCtx, sourceSpan := tracer.Start(context.Background(), "source")
	sourceSpanCtx := sourceSpan.SpanContext()
	sourceSpan.End()

	// Call startLinkedSpan with the source context.
	linkedCtx, linkedSpan := eventbus.StartLinkedSpanForTest(sourceCtx, "bus:test.event")
	linkedSpan.End()

	_ = linkedCtx // linked context should carry the linked span

	// Find the linked span in recorded spans.
	stubs := exporter.GetSpans()
	var linkedStub *tracetest.SpanStub
	for i := range stubs {
		if stubs[i].Name == "bus:test.event" {
			linkedStub = &stubs[i]

			break
		}
	}

	require.NotNil(t, linkedStub, "linked span must be in recorded spans")
	require.Len(t, linkedStub.Links, 1, "linked span must have exactly 1 link")
	require.Equal(t, sourceSpanCtx, linkedStub.Links[0].SpanContext,
		"link must reference the source span")

	// Must be a new root — different TraceID from source.
	require.NotEqual(t, sourceSpanCtx.TraceID(), linkedStub.SpanContext.TraceID(),
		"linked span must have its own trace (WithNewRoot)")
}

func TestStartLinkedSpan_NoopDegradation(t *testing.T) {
	t.Parallel()

	// With a noop provider (default when no real provider is registered), startLinkedSpan
	// must not panic and must return a valid (non-nil) span.
	noopCtx, noopSpan := noop.NewTracerProvider().Tracer("test").Start(
		context.Background(), "source",
	)
	noopSpan.End()

	// This must not panic.
	linkedCtx, linkedSpan := eventbus.StartLinkedSpanForTest(noopCtx, "bus:test.event")
	linkedSpan.End()

	require.NotNil(t, linkedCtx)
	require.NotNil(t, linkedSpan)
	require.False(t, linkedSpan.SpanContext().IsValid(),
		"noop linked span should have invalid SpanContext")
}

//nolint:paralleltest // swaps global TracerProvider
func TestDetachContext_ComposedCancel(
	t *testing.T,
) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	// Set this test's TracerProvider as the global so startLinkedSpan picks it up.
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	// Create a source span context so startLinkedSpan has something to link.
	tracer := tracerProvider.Tracer("test")
	sourceCtx, sourceSpan := tracer.Start(context.Background(), "source")
	sourceSpan.End()

	evt := newTestEvent(1, "test.event")
	detached, cancel := eventbus.DetachContextForTest(sourceCtx, evt, 5*time.Second)

	// The detached context should carry a new span from the linked span creation.
	linkedSpan := trace.SpanFromContext(detached)
	require.True(t, linkedSpan.IsRecording(), "linked span should be recording before cancel")

	// Call the composed cancel — it should end the span.
	cancel()

	// Verify the span was ended by checking the exporter.
	stubs := exporter.GetSpans()
	var found bool
	for _, stub := range stubs {
		if stub.Name == "bus:test.event" {
			found = true
			require.False(t, stub.EndTime.IsZero(), "span must have been ended by cancel")
		}
	}
	require.True(t, found, "linked span must be in recorded spans after cancel")
}

func TestBus_NilEventPanics(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	require.Panics(t, func() {
		bus.Emit(context.Background(), nil)
	})
}

func TestBus_NilHandlerPanicsOnAll(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	require.Panics(t, func() {
		bus.OnAll(nil)
	})
}

func TestBus_NilHandlerPanicsOnType(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	require.Panics(t, func() {
		bus.OnType("test.event", nil)
	})
}

func TestBus_EmptyEventTypePanicsOnType(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	require.Panics(t, func() {
		bus.OnType("", func(_ context.Context, _ eventbus.Event) {})
	})
}

func TestNewBus_FxLifecycle(t *testing.T) {
	t.Parallel()

	var bus eventbus.Bus

	app := fx.New(
		fx.Provide(eventbus.NewBus),
		fx.Supply((*metrics.InfraMetrics)(nil)),
		fx.Populate(&bus),
		fx.NopLogger,
	)

	require.NoError(t, app.Err())
	require.NotNil(t, bus)

	// Start and stop to exercise the lifecycle hook.
	startCtx, startCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startCancel()

	require.NoError(t, app.Start(startCtx))

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	require.NoError(t, app.Stop(stopCtx))
}

func TestBus_SequentialDispatch(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	var mutex sync.Mutex
	var order []int
	done := make(chan struct{})

	const numHandlers = 3

	for i := range numHandlers {
		idx := i
		bus.OnAll(func(_ context.Context, _ eventbus.Event) {
			mutex.Lock()
			order = append(order, idx)
			if len(order) == numHandlers {
				close(done)
			}
			mutex.Unlock()
		})
	}

	bus.Emit(context.Background(), newTestEvent(1, "test.event"))

	select {
	case <-done:
		mutex.Lock()
		require.Equal(t, []int{0, 1, 2}, order,
			"handlers must execute sequentially in registration order")
		mutex.Unlock()
	case <-time.After(5 * time.Second):
		t.Fatal("not all handlers completed within timeout")
	}
}

func TestBus_PanicIsolation_Sequential(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	var mutex sync.Mutex
	var called []int
	done := make(chan struct{})

	// Handler 0: normal
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		mutex.Lock()
		called = append(called, 0)
		mutex.Unlock()
	})

	// Handler 1: panics
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		panic("boom from handler 1")
	})

	// Handler 2: should still run
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		mutex.Lock()
		called = append(called, 2)
		mutex.Unlock()
	})

	// Handler 3: signals completion
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		mutex.Lock()
		called = append(called, 3)
		close(done)
		mutex.Unlock()
	})

	bus.Emit(context.Background(), newTestEvent(1, "test.event"))

	select {
	case <-done:
		mutex.Lock()
		require.Equal(t, []int{0, 2, 3}, called,
			"handlers 0, 2, 3 must execute despite handler 1 panicking")
		mutex.Unlock()
	case <-time.After(5 * time.Second):
		t.Fatal("surviving handlers did not complete within timeout")
	}
}

//nolint:paralleltest // swaps global TracerProvider
func TestBus_EmitLinkedSpanTopology(
	t *testing.T,
) {
	// This test verifies the full Emit → detachContext → handler span link chain:
	//   1. The bus handler span lives in a different trace than the HTTP parent span.
	//   2. The handler span has exactly one link pointing to the parent span.
	//   3. The handler span name follows the "bus:<eventType>" pattern.
	//   4. The detached GameContext carries the linked span's TraceID, not the parent's.

	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	// Simulate an HTTP handler span (the "parent" / trigger span).
	tracer := tracerProvider.Tracer("test-http")
	httpCtx, httpSpan := tracer.Start(context.Background(), "HTTP POST /move")
	httpSpanCtx := httpSpan.SpanContext()

	// Build a GameContext rooted in the HTTP span context.
	traceCtx := ctx.WithSpan(httpCtx, httpSpan)
	userCtx := ctx.WithUserID(traceCtx, "player-1")
	gameCtx := gamectx.WithGameID(userCtx, 99)

	bus := eventbus.NewBusForTest()

	type handlerCapture struct {
		handlerCtx context.Context //nolint:containedctx // test-only: capturing ctx for post-dispatch assertions
		traceID    trace.TraceID
	}

	captured := make(chan handlerCapture, 1)

	bus.OnAll(func(handlerCtx context.Context, _ eventbus.Event) {
		span := trace.SpanFromContext(handlerCtx)
		captured <- handlerCapture{
			handlerCtx: handlerCtx,
			traceID:    span.SpanContext().TraceID(),
		}
	})

	evt := newTestEvent(99, "move.executed")
	bus.Emit(gameCtx, evt)
	httpSpan.End()

	var got handlerCapture
	select {
	case got = <-captured:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not receive event within timeout")
	}

	// Flush all spans from the exporter.
	stubs := exporter.GetSpans()

	// Find the bus handler span by name.
	var busStub *tracetest.SpanStub
	for i := range stubs {
		if stubs[i].Name == "bus:move.executed" {
			busStub = &stubs[i]

			break
		}
	}

	require.NotNil(t, busStub, "bus handler span must be present in recorded spans")

	// AC1: HTTP parent span and bus handler span have different TraceIDs.
	require.NotEqual(t, httpSpanCtx.TraceID(), busStub.SpanContext.TraceID(),
		"bus handler span must live in a different trace than the HTTP parent")

	// AC2: Bus handler span has exactly 1 link pointing to the HTTP parent span.
	require.Len(t, busStub.Links, 1, "bus handler span must have exactly 1 link")
	require.Equal(t, httpSpanCtx, busStub.Links[0].SpanContext,
		"link must reference the HTTP parent span's SpanContext")

	// AC3: Bus handler span name follows "bus:<eventType>" pattern.
	require.Equal(t, "bus:move.executed", busStub.Name,
		"span name must follow bus:<eventType> pattern")

	// AC4: Detached GameContext carries the linked span's TraceID, not the parent's.
	require.Equal(t, busStub.SpanContext.TraceID(), got.traceID,
		"handler context must carry the linked span's TraceID")
	require.NotEqual(t, httpSpanCtx.TraceID(), got.traceID,
		"handler context must NOT carry the HTTP parent's TraceID")

	// Verify the context is still a GameContext with preserved domain metadata.
	handlerGameCtx, ok := got.handlerCtx.(gamectx.GameContext)
	require.True(t, ok, "detached context must be a GameContext, got %T", got.handlerCtx)
	require.Equal(t, int64(99), handlerGameCtx.GameID())
	require.Equal(t, "player-1", handlerGameCtx.UserID())
}

func TestCollectHandlers_Ordering(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBusForTest()

	var order []string

	// Register OnAll handlers first.
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		order = append(order, "all-0")
	})
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		order = append(order, "all-1")
	})

	// Register OnType handler.
	bus.OnType("test.event", func(_ context.Context, _ eventbus.Event) {
		order = append(order, "typed-0")
	})

	handlers := eventbus.CollectHandlersForTest(bus, "test.event")
	require.Len(t, handlers, 3)

	// Execute them to verify ordering.
	for _, h := range handlers {
		h(context.Background(), newTestEvent(1, "test.event"))
	}

	require.Equal(t, []string{"all-0", "all-1", "typed-0"}, order,
		"OnAll handlers must come before OnType handlers")
}
