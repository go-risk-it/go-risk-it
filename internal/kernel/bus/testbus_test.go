package bus_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/stretchr/testify/require"
)

func TestTestBus_CapturesEvents(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	evt1 := gameevt.NewGameCreated(1, now, 4)
	evt2 := gameevt.NewPlayerConnected(1, "user1", now)
	evt3 := gameevt.NewGameCompleted(1, "user1", now, 10)

	bus.Emit(context.Background(), evt1)
	bus.Emit(context.Background(), evt2)
	bus.Emit(context.Background(), evt3)

	captured := bus.Events()
	require.Len(t, captured, 3)
	require.Equal(t, evt1, captured[0])
	require.Equal(t, evt2, captured[1])
	require.Equal(t, evt3, captured[2])
}

func TestTestBus_EventsOfType(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	move1 := gameevt.NewMoveExecuted(
		1, "user1", now,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GameMoveLog{ID: 1},
		sqlc.GamePhaseTypeDEPLOY,
		false, 1,
		nil, nil,
	)
	created := gameevt.NewGameCreated(1, now, 4)
	move2 := gameevt.NewMoveExecuted(
		1, "user2", now,
		sqlc.GamePhaseTypeATTACK,
		sqlc.GameMoveLog{ID: 2},
		sqlc.GamePhaseTypeATTACK,
		false, 2,
		nil, nil,
	)
	phase := gameevt.NewPhaseTransitioned(
		1, "user1", now,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GamePhaseTypeATTACK,
		1,
	)

	bus.Emit(context.Background(), move1)
	bus.Emit(context.Background(), created)
	bus.Emit(context.Background(), move2)
	bus.Emit(context.Background(), phase)

	moves := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
	require.Len(t, moves, 2)
	require.Equal(t, move1, moves[0])
	require.Equal(t, move2, moves[1])

	phases := eventbus.EventsOfType[*gameevt.PhaseTransitioned](bus)
	require.Len(t, phases, 1)
	require.Equal(t, phase, phases[0])

	completions := eventbus.EventsOfType[*gameevt.GameCompleted](bus)
	require.Empty(t, completions)
}

func TestTestBus_Reset(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	bus.Emit(context.Background(), gameevt.NewGameCreated(1, now, 4))
	bus.Emit(context.Background(), gameevt.NewGameCreated(2, now, 4))

	require.Len(t, bus.Events(), 2)

	bus.Reset()

	require.Empty(t, bus.Events())
}

func TestTestBus_SynchronousDispatch(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	handlerCalled := false

	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		handlerCalled = true
	})

	bus.Emit(context.Background(), gameevt.NewGameCreated(1, now, 4))

	// If dispatch were async, this assertion could race. Synchronous dispatch
	// guarantees the flag is set by the time Emit returns.
	require.True(t, handlerCalled, "handler must be called synchronously before Emit returns")
}

func TestTestBus_HandlersCalledOnEmit(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	var allEvents []eventbus.Event
	var typedEvents []eventbus.Event

	bus.OnAll(func(_ context.Context, event eventbus.Event) {
		allEvents = append(allEvents, event)
	})
	bus.OnType(gameevt.TypeGameCreated, func(_ context.Context, event eventbus.Event) {
		typedEvents = append(typedEvents, event)
	})

	created := gameevt.NewGameCreated(1, now, 4)
	connected := gameevt.NewPlayerConnected(1, "user1", now)

	bus.Emit(context.Background(), created)
	bus.Emit(context.Background(), connected)

	// OnAll receives both events.
	require.Len(t, allEvents, 2)
	require.Equal(t, created, allEvents[0])
	require.Equal(t, connected, allEvents[1])

	// OnType(TypeGameCreated) receives only the GameCreated event.
	require.Len(t, typedEvents, 1)
	require.Equal(t, created, typedEvents[0])
}

func TestTestBus_Close(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()

	err := bus.Close(context.Background())
	require.NoError(t, err)
}
