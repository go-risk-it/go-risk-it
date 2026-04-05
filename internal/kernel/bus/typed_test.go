package bus_test

import (
	"context"
	"testing"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/stretchr/testify/require"
)

// spyBus captures OnType registration calls for verifying OnEvent behavior
// without async dispatch complexity.
type spyBus struct {
	registeredType    string
	registeredHandler bus.Handler
}

func (s *spyBus) OnType(eventType string, handler bus.Handler) {
	s.registeredType = eventType
	s.registeredHandler = handler
}

func (s *spyBus) OnAll(bus.Handler) {}

func TestOnEvent_RegistersCorrectType(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}

	bus.OnEvent[*gameevt.MoveCompleted](spy, func(_ context.Context, _ *gameevt.MoveCompleted) {})

	require.Equal(t, gameevt.TypeMoveCompleted, spy.registeredType)
}

func TestOnEvent_TypedHandlerCalledWithCorrectEvent(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	var received *gameevt.MoveCompleted

	bus.OnEvent[*gameevt.MoveCompleted](spy, func(_ context.Context, evt *gameevt.MoveCompleted) {
		received = evt
	})

	event := gameevt.NewMoveCompleted(
		42, "user1", now,
		gameapi.GamePhaseTypeDEPLOY, 1,
		gameapi.GamePhaseTypeDEPLOY, gameapi.GamePhaseTypeDEPLOY,
		false, nil, nil, nil,
	)
	spy.registeredHandler(context.Background(), event)

	require.NotNil(t, received)
	require.Equal(t, int64(42), received.GameID())
}

func TestOnEvent_TypeAssertionMismatchDoesNotCallHandler(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	called := false

	bus.OnEvent[*gameevt.MoveCompleted](spy, func(_ context.Context, _ *gameevt.MoveCompleted) {
		called = true
	})

	wrongEvent := gameevt.NewPlayerConnected(42, "user1", time.Now())
	spy.registeredHandler(context.Background(), wrongEvent)

	require.False(t, called, "typed handler should not be called when event type does not match")
}

func TestOnEvent_MultipleTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		register     func(spy *spyBus)
		expectedType string
	}{
		{
			name: "MoveCompleted",
			register: func(spy *spyBus) {
				bus.OnEvent[*gameevt.MoveCompleted](
					spy,
					func(_ context.Context, _ *gameevt.MoveCompleted) {},
				)
			},
			expectedType: gameevt.TypeMoveCompleted,
		},
		{
			name: "GameCreated",
			register: func(spy *spyBus) {
				bus.OnEvent[*gameevt.GameCreated](
					spy,
					func(_ context.Context, _ *gameevt.GameCreated) {},
				)
			},
			expectedType: gameevt.TypeGameCreated,
		},
		{
			name: "PlayerConnected",
			register: func(spy *spyBus) {
				bus.OnEvent[*gameevt.PlayerConnected](
					spy,
					func(_ context.Context, _ *gameevt.PlayerConnected) {},
				)
			},
			expectedType: gameevt.TypePlayerConnected,
		},
		{
			name: "TurnEnded",
			register: func(spy *spyBus) {
				bus.OnEvent[*gameevt.TurnEnded](
					spy,
					func(_ context.Context, _ *gameevt.TurnEnded) {},
				)
			},
			expectedType: gameevt.TypeTurnEnded,
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
