package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	"github.com/stretchr/testify/require"
)

// spyBus captures OnType registration calls for verifying OnEvent behavior
// without async dispatch complexity.
type spyBus struct {
	registeredType    string
	registeredHandler events.Handler
}

func (s *spyBus) OnType(eventType string, handler events.Handler) {
	s.registeredType = eventType
	s.registeredHandler = handler
}

func (s *spyBus) OnAll(events.Handler)               {}
func (s *spyBus) Emit(context.Context, events.Event) {}
func (s *spyBus) Close(context.Context) error        { return nil }

func TestOnEvent_RegistersCorrectType(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}

	events.OnEvent[*gameevt.MoveExecuted](spy, func(_ context.Context, _ *gameevt.MoveExecuted) {})

	require.Equal(t, gameevt.TypeMoveExecuted, spy.registeredType)
}

func TestOnEvent_TypedHandlerCalledWithCorrectEvent(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	var received *gameevt.MoveExecuted

	events.OnEvent[*gameevt.MoveExecuted](spy, func(_ context.Context, evt *gameevt.MoveExecuted) {
		received = evt
	})

	// Invoke the wrapped handler directly with the correct event type.
	event := gameevt.NewMoveExecuted(
		42, "user1", now,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GameMoveLog{},
		sqlc.GamePhaseTypeDEPLOY,
		false, 1, nil, nil,
	)
	spy.registeredHandler(context.Background(), event)

	require.NotNil(t, received)
	require.Equal(t, int64(42), received.GameID())
}

func TestOnEvent_TypeAssertionMismatchDoesNotCallHandler(t *testing.T) {
	t.Parallel()

	spy := &spyBus{}
	called := false

	events.OnEvent[*gameevt.MoveExecuted](spy, func(_ context.Context, _ *gameevt.MoveExecuted) {
		called = true
	})

	// Invoke the wrapped handler with a different event type.
	// The type assertion inside the wrapper should fail silently.
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
			name: "MoveExecuted",
			register: func(spy *spyBus) {
				events.OnEvent[*gameevt.MoveExecuted](
					spy,
					func(_ context.Context, _ *gameevt.MoveExecuted) {},
				)
			},
			expectedType: gameevt.TypeMoveExecuted,
		},
		{
			name: "PhaseTransitioned",
			register: func(spy *spyBus) {
				events.OnEvent[*gameevt.PhaseTransitioned](
					spy,
					func(_ context.Context, _ *gameevt.PhaseTransitioned) {},
				)
			},
			expectedType: gameevt.TypePhaseTransitioned,
		},
		{
			name: "GameCompleted",
			register: func(spy *spyBus) {
				events.OnEvent[*gameevt.GameCompleted](
					spy,
					func(_ context.Context, _ *gameevt.GameCompleted) {},
				)
			},
			expectedType: gameevt.TypeGameCompleted,
		},
		{
			name: "GameCreated",
			register: func(spy *spyBus) {
				events.OnEvent[*gameevt.GameCreated](
					spy,
					func(_ context.Context, _ *gameevt.GameCreated) {},
				)
			},
			expectedType: gameevt.TypeGameCreated,
		},
		{
			name: "PlayerConnected",
			register: func(spy *spyBus) {
				events.OnEvent[*gameevt.PlayerConnected](
					spy,
					func(_ context.Context, _ *gameevt.PlayerConnected) {},
				)
			},
			expectedType: gameevt.TypePlayerConnected,
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
