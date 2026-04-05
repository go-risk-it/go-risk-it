package game_test

import (
	"testing"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance guards.
var (
	_ gameevt.GameEvent = (*gameevt.MoveCompleted)(nil)
	_ gameevt.GameEvent = (*gameevt.GameCreated)(nil)
	_ gameevt.GameEvent = (*gameevt.PlayerConnected)(nil)
	_ gameevt.GameEvent = (*gameevt.TurnEnded)(nil)

	// GameCreationFailed implements bus.Event but NOT GameEvent.
	_ eventbus.Event = (*gameevt.GameCreationFailed)(nil)
)

func TestEventTypes_NilPointerSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		callType func() string
	}{
		{
			name:     "MoveCompleted nil pointer",
			callType: (*gameevt.MoveCompleted)(nil).EventType,
		},
		{
			name:     "GameCreated nil pointer",
			callType: (*gameevt.GameCreated)(nil).EventType,
		},
		{
			name:     "GameCreationFailed nil pointer",
			callType: (*gameevt.GameCreationFailed)(nil).EventType,
		},
		{
			name:     "PlayerConnected nil pointer",
			callType: (*gameevt.PlayerConnected)(nil).EventType,
		},
		{
			name:     "TurnEnded nil pointer",
			callType: (*gameevt.TurnEnded)(nil).EventType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				_ = test.callType()
			})
		})
	}
}

func TestEventTypes_EventType(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		event    eventbus.Event
		expected string
	}{
		{
			name:     "MoveCompleted",
			event:    gameevt.NewMoveCompleted(1, "u", now, "DEPLOY", 1, "DEPLOY", "DEPLOY", false, nil, nil, nil),
			expected: gameevt.TypeMoveCompleted,
		},
		{
			name:     "GameCreated",
			event:    gameevt.NewGameCreated(1, 0, now, 4, nil, nil),
			expected: gameevt.TypeGameCreated,
		},
		{
			name:     "GameCreationFailed",
			event:    gameevt.NewGameCreationFailed(1, now, "boom"),
			expected: gameevt.TypeGameCreationFailed,
		},
		{
			name:     "PlayerConnected",
			event:    gameevt.NewPlayerConnected(1, "u", now),
			expected: gameevt.TypePlayerConnected,
		},
		{
			name:     "TurnEnded",
			event:    gameevt.NewTurnEnded(1, "u", now, 1),
			expected: gameevt.TypeTurnEnded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.expected, test.event.EventType())
		})
	}
}

func TestGameEvents_GameIDAndTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		event     gameevt.GameEvent
		gameID    int64
		timestamp time.Time
	}{
		{
			name: "MoveCompleted",
			event: gameevt.NewMoveCompleted(
				42, "user1", now,
				gameapi.GamePhaseTypeDEPLOY,
				1,
				gameapi.GamePhaseTypeDEPLOY,
				gameapi.GamePhaseTypeDEPLOY,
				false, nil, nil, nil,
			),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "GameCreated",
			event:     gameevt.NewGameCreated(42, 10, now, 4, nil, nil),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "PlayerConnected",
			event:     gameevt.NewPlayerConnected(42, "user1", now),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "TurnEnded",
			event:     gameevt.NewTurnEnded(42, "user1", now, 5),
			gameID:    42,
			timestamp: now,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.gameID, test.event.GameID())
			require.Equal(t, test.timestamp, test.event.EventTimestamp())
		})
	}
}

func TestGameCreated_LobbyID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	event := gameevt.NewGameCreated(42, 99, now, 4, nil, nil)
	require.Equal(t, int64(99), event.LobbyID())
}

func TestGameCreationFailed_LobbyIDAndTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()
	event := gameevt.NewGameCreationFailed(77, now, "out of memory")

	require.Equal(t, int64(77), event.LobbyID())
	require.Equal(t, now, event.EventTimestamp())
}

func TestGameCreationFailed_DoesNotImplementGameEvent(t *testing.T) {
	t.Parallel()

	event := gameevt.NewGameCreationFailed(1, time.Now(), "reason")

	// GameCreationFailed implements bus.Event ...
	var busEvent eventbus.Event = event
	require.NotNil(t, busEvent)

	// ... but NOT GameEvent (no GameID method).
	_, ok := busEvent.(gameevt.GameEvent)
	require.False(t, ok, "GameCreationFailed must not implement GameEvent")
}

func TestMoveCompleted_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	event := gameevt.NewMoveCompleted(
		42, "attacker", now,
		gameapi.GamePhaseTypeATTACK,
		5,
		gameapi.GamePhaseTypeDEPLOY,
		gameapi.GamePhaseTypeATTACK,
		false, nil, nil, nil,
	)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeMoveCompleted, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "attacker", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, "ATTACK", record["action_type"])
	require.Equal(t, int64(5), record["turn"])
	require.Equal(t, "DEPLOY", record["from_phase"])
	require.Equal(t, "ATTACK", record["target_phase"])
	require.Equal(t, false, record["game_over"])
}

func TestMoveCompleted_ToRecord_GameOver(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	event := gameevt.NewMoveCompleted(
		42, "winner", now,
		gameapi.GamePhaseTypeATTACK,
		10,
		gameapi.GamePhaseTypeATTACK,
		gameapi.GamePhaseTypeATTACK,
		true, nil, nil, nil,
	)

	record := event.ToRecord()
	require.Equal(t, true, record["game_over"])
}

func TestGameCreated_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewGameCreated(42, 99, now, 4, nil, nil)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeGameCreated, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, 4, record["num_players"])
	require.Equal(t, int64(99), record["lobby_id"])
}

func TestGameCreated_ToRecord_ZeroLobbyIDOmitted(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewGameCreated(42, 0, now, 4, nil, nil)

	record := event.ToRecord()

	require.NotContains(t, record, "lobby_id", "lobby_id should be omitted when zero")
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, 4, record["num_players"])
}

func TestGameCreationFailed_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewGameCreationFailed(77, now, "out of memory")

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeGameCreationFailed, record["event_type"])
	require.Equal(t, int64(77), record["lobby_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, "out of memory", record["reason"])

	// Must not have game_id (not a GameEvent).
	require.NotContains(t, record, "game_id")
}

func TestPlayerConnected_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewPlayerConnected(42, "player1", now)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypePlayerConnected, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
}

func TestMoveCompleted_ToRecord_ExcludesSnapshots(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	event := gameevt.NewMoveCompleted(
		42, "attacker", now,
		gameapi.GamePhaseTypeATTACK,
		5,
		gameapi.GamePhaseTypeDEPLOY,
		gameapi.GamePhaseTypeATTACK,
		false,
		&snapshot.GameSnapshot{
			Game:    snapshot.GameMeta{ID: 42, Turn: 5},
			Regions: []snapshot.RegionState{{ID: "r1", OwnerID: "attacker", Troops: 3}},
		},
		map[string]*snapshot.PlayerPrivate{
			"attacker": {Cards: []snapshot.CardState{{ID: 1}}},
		},
		[]snapshot.RegionState{{ID: "r1", OwnerID: "defender", Troops: 5}},
	)

	record := event.ToRecord()

	require.NotContains(t, record, "public_snapshot",
		"ToRecord must exclude public_snapshot to avoid logging large payloads")
	require.NotContains(t, record, "private_snapshots",
		"ToRecord must exclude private_snapshots to avoid logging large payloads")
	require.NotContains(t, record, "previous_regions",
		"ToRecord must exclude previous_regions to avoid logging large payloads")

	// Sanity: scalar fields are still present.
	require.Equal(t, gameevt.TypeMoveCompleted, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
}

func TestGameCreated_ToRecord_ExcludesSnapshots(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	event := gameevt.NewGameCreated(
		42, 99, now, 4,
		&snapshot.GameSnapshot{
			Game:    snapshot.GameMeta{ID: 42, Turn: 1},
			Regions: []snapshot.RegionState{{ID: "r1", OwnerID: "p1", Troops: 1}},
		},
		map[string]*snapshot.PlayerPrivate{
			"p1": {Cards: []snapshot.CardState{{ID: 1}}},
		},
	)

	record := event.ToRecord()

	require.NotContains(t, record, "public_snapshot",
		"ToRecord must exclude public_snapshot to avoid logging large payloads")
	require.NotContains(t, record, "private_snapshots",
		"ToRecord must exclude private_snapshots to avoid logging large payloads")

	// Sanity: scalar fields are still present.
	require.Equal(t, gameevt.TypeGameCreated, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, 4, record["num_players"])
}

func TestTurnEnded_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewTurnEnded(42, "player1", now, 7)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeTurnEnded, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, int64(7), record["turn"])
	require.Equal(t, "WAITING", record["action_type"])
}
