package events_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/events"
	"github.com/stretchr/testify/require"
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

	bus := events.NewBusForTest()
	received := make(chan events.Event, 1)

	bus.OnAll(func(_ context.Context, event events.Event) {
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

	bus := events.NewBusForTest()

	const numHandlers = 5
	var count atomic.Int32
	done := make(chan struct{})

	for range numHandlers {
		bus.OnAll(func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()

	matchCh := make(chan events.Event, 1)
	noMatchCh := make(chan struct{}, 1)

	bus.OnType("move.executed", func(_ context.Context, event events.Event) {
		matchCh <- event
	})

	bus.OnType("game.created", func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()

	allCh := make(chan events.Event, 1)
	typedCh := make(chan events.Event, 1)

	bus.OnAll(func(_ context.Context, event events.Event) {
		allCh <- event
	})
	bus.OnType("move.executed", func(_ context.Context, event events.Event) {
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

	bus := events.NewBusForTest()

	const numEmits = 100
	var count atomic.Int32
	done := make(chan struct{})

	bus.OnAll(func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()
	survived := make(chan struct{}, 1)

	// First handler panics.
	bus.OnAll(func(_ context.Context, _ events.Event) {
		panic("boom")
	})

	// Second handler should still execute.
	bus.OnAll(func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()
	started := make(chan struct{})
	proceed := make(chan struct{})
	handlerDone := make(chan struct{})

	bus.OnAll(func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()
	started := make(chan struct{})

	bus.OnAll(func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()

	called := make(chan struct{}, 1)
	bus.OnAll(func(_ context.Context, _ events.Event) {
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

	bus := events.NewBusForTest()

	err1 := bus.Close(context.Background())
	require.NoError(t, err1)

	err2 := bus.Close(context.Background())
	require.NoError(t, err2)
}

func TestDetachContext_GameContext(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-123")
	gameCtx := ctx.WithGameID(userCtx, 42)

	detached, cancel := events.DetachContextForTest(gameCtx, 5*time.Second)
	defer cancel()

	// Must preserve GameID through the detached context.
	gameContext, ok := detached.(ctx.GameContext)
	require.True(t, ok, "expected GameContext, got %T", detached)
	require.Equal(t, int64(42), gameContext.GameID())

	// Must have a deadline from the timeout.
	deadline, hasDeadline := detached.Deadline()
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)
}

func TestDetachContext_LobbyContext(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-456")
	lobbyCtx := ctx.WithLobbyID(userCtx, 77)

	detached, cancel := events.DetachContextForTest(lobbyCtx, 5*time.Second)
	defer cancel()

	// Must preserve LobbyID through the detached context.
	lobbyContext, ok := detached.(ctx.LobbyContext)
	require.True(t, ok, "expected LobbyContext, got %T", detached)
	require.Equal(t, int64(77), lobbyContext.LobbyID())

	// Must have a deadline from the timeout.
	deadline, hasDeadline := detached.Deadline()
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)
}

func TestDetachContext_PlainContext(t *testing.T) {
	t.Parallel()

	plain := context.Background()

	detached, cancel := events.DetachContextForTest(plain, 5*time.Second)
	defer cancel()

	// Should have a deadline.
	deadline, hasDeadline := detached.Deadline()
	require.True(t, hasDeadline)
	require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)

	// Should NOT be a GameContext.
	_, isGame := detached.(ctx.GameContext)
	require.False(t, isGame, "plain context should not become GameContext")

	// Should NOT be a LobbyContext.
	_, isLobby := detached.(ctx.LobbyContext)
	require.False(t, isLobby, "plain context should not become LobbyContext")
}

func TestBus_NilEventPanics(t *testing.T) {
	t.Parallel()

	bus := events.NewBusForTest()

	require.Panics(t, func() {
		bus.Emit(context.Background(), nil)
	})
}

func TestBus_NilHandlerPanicsOnAll(t *testing.T) {
	t.Parallel()

	bus := events.NewBusForTest()

	require.Panics(t, func() {
		bus.OnAll(nil)
	})
}

func TestBus_NilHandlerPanicsOnType(t *testing.T) {
	t.Parallel()

	bus := events.NewBusForTest()

	require.Panics(t, func() {
		bus.OnType("test.event", nil)
	})
}

func TestBus_EmptyEventTypePanicsOnType(t *testing.T) {
	t.Parallel()

	bus := events.NewBusForTest()

	require.Panics(t, func() {
		bus.OnType("", func(_ context.Context, _ events.Event) {})
	})
}

func TestNewBus_FxLifecycle(t *testing.T) {
	t.Parallel()

	var bus events.Bus

	app := fx.New(
		fx.Provide(events.NewBus),
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
