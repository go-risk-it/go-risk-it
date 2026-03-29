package lobby_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// spyBus captures OnType registration calls for verifying OnLobbyEvent behavior
// without async dispatch complexity.
type spyBus struct {
	registeredType    string
	registeredHandler bus.Handler
}

func (s *spyBus) OnType(eventType string, handler bus.Handler) {
	s.registeredType = eventType
	s.registeredHandler = handler
}

func (s *spyBus) OnAll(bus.Handler)               {}
func (s *spyBus) Emit(context.Context, bus.Event) {}
func (s *spyBus) Close(context.Context) error     { return nil }

func TestOnLobbyEvent_RegistersCorrectType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		register     func(spy *spyBus)
		expectedType string
	}{
		{
			name: "LobbyStateChanged",
			register: func(spy *spyBus) {
				lobbyevt.OnLobbyEvent(
					spy,
					func(_ ctx.LobbyContext, _ *lobbyevt.LobbyStateChanged) {},
				)
			},
			expectedType: lobbyevt.TypeLobbyStateChanged,
		},
		{
			name: "LobbyPlayerConnected",
			register: func(spy *spyBus) {
				lobbyevt.OnLobbyEvent(
					spy,
					func(_ ctx.LobbyContext, _ *lobbyevt.LobbyPlayerConnected) {},
				)
			},
			expectedType: lobbyevt.TypeLobbyPlayerConnected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			spy := &spyBus{}
			test.register(spy)

			require.Equal(t, test.expectedType, spy.registeredType)
		})
	}
}

func TestOnLobbyEvent_LobbyContextPassed(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}

	var (
		receivedCtx   ctx.LobbyContext
		receivedEvent *lobbyevt.LobbyStateChanged
	)

	lobbyevt.OnLobbyEvent(spy, func(lc ctx.LobbyContext, evt *lobbyevt.LobbyStateChanged) {
		receivedCtx = lc
		receivedEvent = evt
	})

	// Build a real LobbyContext.
	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "user-42")
	lobbyCtx := ctx.WithLobbyID(userCtx, 99)

	event := lobbyevt.NewLobbyStateChanged(99, "user-42")
	spy.registeredHandler(lobbyCtx, event)

	require.NotNil(t, receivedCtx)
	require.Equal(t, int64(99), receivedCtx.LobbyID())
	require.Equal(t, "user-42", receivedCtx.UserID())

	require.NotNil(t, receivedEvent)
	require.Equal(t, int64(99), receivedEvent.LobbyID())
}

func TestOnLobbyEvent_TypeMismatch_Skips(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	called := false

	// Register for LobbyStateChanged.
	lobbyevt.OnLobbyEvent(spy, func(_ ctx.LobbyContext, _ *lobbyevt.LobbyStateChanged) {
		called = true
	})

	// Dispatch with the correct LobbyContext but a different event type.
	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "user-42")
	lobbyCtx := ctx.WithLobbyID(userCtx, 99)

	wrongEvent := lobbyevt.NewLobbyPlayerConnected(99, "user-42")
	spy.registeredHandler(lobbyCtx, wrongEvent)

	require.False(t, called, "handler should not be called when event type does not match")
}

func TestOnLobbyEvent_NonLobbyContext_Skips(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	called := false

	lobbyevt.OnLobbyEvent(spy, func(_ ctx.LobbyContext, _ *lobbyevt.LobbyStateChanged) {
		called = true
	})

	// Dispatch with a plain context.Background() — not a LobbyContext.
	event := lobbyevt.NewLobbyStateChanged(99, "user-42")
	spy.registeredHandler(context.Background(), event)

	require.False(t, called, "handler should not be called when context is not LobbyContext")
}
