package bus_test

import (
	"context"
	"testing"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/stretchr/testify/require"
)

func TestTestBus_CapturesEvents(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	evt1 := gameevt.NewGameCreated(1, 0, now, 4, nil, nil)
	evt2 := gameevt.NewPlayerConnected(1, "user1", now)
	evt3 := gameevt.NewMoveCompleted(1, "user1", now,
		gameapi.GamePhaseTypeDEPLOY, 10,
		gameapi.GamePhaseTypeDEPLOY, gameapi.GamePhaseTypeDEPLOY,
		true, nil, nil, nil)

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

	move1 := gameevt.NewMoveCompleted(
		1, "user1", now,
		gameapi.GamePhaseTypeDEPLOY, 1,
		gameapi.GamePhaseTypeDEPLOY, gameapi.GamePhaseTypeDEPLOY,
		false, nil, nil, nil,
	)
	created := gameevt.NewGameCreated(1, 0, now, 4, nil, nil)
	move2 := gameevt.NewMoveCompleted(
		1, "user2", now,
		gameapi.GamePhaseTypeATTACK, 2,
		gameapi.GamePhaseTypeATTACK, gameapi.GamePhaseTypeATTACK,
		false, nil, nil, nil,
	)
	connected := gameevt.NewPlayerConnected(1, "user1", now)

	bus.Emit(context.Background(), move1)
	bus.Emit(context.Background(), created)
	bus.Emit(context.Background(), move2)
	bus.Emit(context.Background(), connected)

	moves := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, moves, 2)
	require.Equal(t, move1, moves[0])
	require.Equal(t, move2, moves[1])

	connects := eventbus.EventsOfType[*gameevt.PlayerConnected](bus)
	require.Len(t, connects, 1)
	require.Equal(t, connected, connects[0])

	created2 := eventbus.EventsOfType[*gameevt.GameCreated](bus)
	require.Len(t, created2, 1)
}

func TestTestBus_Reset(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	bus.Emit(context.Background(), gameevt.NewGameCreated(1, 0, now, 4, nil, nil))
	bus.Emit(context.Background(), gameevt.NewGameCreated(2, 0, now, 4, nil, nil))

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

	bus.Emit(context.Background(), gameevt.NewGameCreated(1, 0, now, 4, nil, nil))

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

	created := gameevt.NewGameCreated(1, 0, now, 4, nil, nil)
	connected := gameevt.NewPlayerConnected(1, "user1", now)

	bus.Emit(context.Background(), created)
	bus.Emit(context.Background(), connected)

	require.Len(t, allEvents, 2)
	require.Equal(t, created, allEvents[0])
	require.Equal(t, connected, allEvents[1])

	require.Len(t, typedEvents, 1)
	require.Equal(t, created, typedEvents[0])
}

func TestTestBus_Close(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewTestBus()

	err := bus.Close(context.Background())
	require.NoError(t, err)
}
