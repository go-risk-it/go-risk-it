package game_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// spyBus captures OnType registration calls for verifying OnGameEvent behavior
// without async dispatch complexity. Mirrors the pattern from events/typed_test.go.
type spyBus struct {
	registeredType    string
	registeredHandler bus.Handler
}

func (s *spyBus) OnType(eventType string, handler bus.Handler) {
	s.registeredType = eventType
	s.registeredHandler = handler
}

func (s *spyBus) OnAll(bus.Handler) {}

func newTestGameContext(gameID int64) ctx.GameContext {
	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")

	return ctx.WithGameID(userCtx, gameID)
}

func TestOnGameEvent_RegistersCorrectType(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}

	gameevt.OnGameEvent[*gameevt.MoveExecuted](
		spy,
		func(_ ctx.GameContext, _ *gameevt.MoveExecuted) {},
	)

	require.Equal(t, gameevt.TypeMoveExecuted, spy.registeredType)
}

func TestOnGameEvent_GameContextPassed(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)

	var (
		receivedCtx ctx.GameContext
		receivedEvt *gameevt.PlayerConnected
	)

	gameevt.OnGameEvent[*gameevt.PlayerConnected](
		spy,
		func(gc ctx.GameContext, evt *gameevt.PlayerConnected) {
			receivedCtx = gc
			receivedEvt = evt
		},
	)

	gameCtx := newTestGameContext(42)
	event := gameevt.NewPlayerConnected(42, "user1", now)

	spy.registeredHandler(gameCtx, event)

	require.NotNil(t, receivedCtx, "handler should have been called")
	require.Equal(t, int64(42), receivedCtx.GameID())
	require.NotNil(t, receivedEvt)
	require.Equal(t, int64(42), receivedEvt.GameID())
}

func TestOnGameEvent_NonGameContext_Skips(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	called := false

	gameevt.OnGameEvent[*gameevt.PlayerConnected](
		spy,
		func(_ ctx.GameContext, _ *gameevt.PlayerConnected) {
			called = true
		},
	)

	// Dispatch with a plain context.Background() — not a GameContext.
	event := gameevt.NewPlayerConnected(42, "user1", time.Now())
	spy.registeredHandler(context.Background(), event)

	require.False(t, called, "handler should not be called when context is not GameContext")
}

func TestOnGameEvent_TypeMismatch_Skips(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	called := false

	gameevt.OnGameEvent[*gameevt.MoveExecuted](
		spy,
		func(_ ctx.GameContext, _ *gameevt.MoveExecuted) {
			called = true
		},
	)

	// Dispatch with the correct context but a wrong event type.
	// The inner OnEvent type assertion will fail silently.
	gameCtx := newTestGameContext(42)
	wrongEvent := gameevt.NewPlayerConnected(42, "user1", time.Now())

	spy.registeredHandler(gameCtx, wrongEvent)

	require.False(t, called, "handler should not be called when event type does not match")
}
