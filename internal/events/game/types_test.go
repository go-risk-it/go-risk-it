package game_test

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance guards.
var (
	_ gameevt.GameEvent = (*gameevt.MoveExecuted)(nil)
	_ gameevt.GameEvent = (*gameevt.PhaseTransitioned)(nil)
	_ gameevt.GameEvent = (*gameevt.GameCompleted)(nil)
	_ gameevt.GameEvent = (*gameevt.GameCreated)(nil)
	_ gameevt.GameEvent = (*gameevt.PlayerConnected)(nil)
)

func TestEventTypes_NilPointerSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		callType func() string
	}{
		{
			name:     "MoveExecuted nil pointer",
			callType: (*gameevt.MoveExecuted)(nil).EventType,
		},
		{
			name:     "PhaseTransitioned nil pointer",
			callType: (*gameevt.PhaseTransitioned)(nil).EventType,
		},
		{
			name:     "GameCompleted nil pointer",
			callType: (*gameevt.GameCompleted)(nil).EventType,
		},
		{
			name:     "GameCreated nil pointer",
			callType: (*gameevt.GameCreated)(nil).EventType,
		},
		{
			name:     "PlayerConnected nil pointer",
			callType: (*gameevt.PlayerConnected)(nil).EventType,
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

	tests := []struct {
		name     string
		event    gameevt.GameEvent
		expected string
	}{
		{
			name:     "MoveExecuted",
			event:    &gameevt.MoveExecuted{},
			expected: gameevt.TypeMoveExecuted,
		},
		{
			name:     "PhaseTransitioned",
			event:    &gameevt.PhaseTransitioned{},
			expected: gameevt.TypePhaseTransitioned,
		},
		{
			name:     "GameCompleted",
			event:    &gameevt.GameCompleted{},
			expected: gameevt.TypeGameCompleted,
		},
		{
			name:     "GameCreated",
			event:    &gameevt.GameCreated{},
			expected: gameevt.TypeGameCreated,
		},
		{
			name:     "PlayerConnected",
			event:    &gameevt.PlayerConnected{},
			expected: gameevt.TypePlayerConnected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.expected, test.event.EventType())
		})
	}
}

func TestEventTypes_GameIDAndTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		event     gameevt.GameEvent
		gameID    int64
		timestamp time.Time
	}{
		{
			name: "MoveExecuted",
			event: gameevt.NewMoveExecuted(
				42, "user1", now,
				sqlc.GamePhaseTypeDEPLOY,
				sqlc.GameMoveLog{},
				sqlc.GamePhaseTypeDEPLOY,
				false, 1, nil, nil,
			),
			gameID:    42,
			timestamp: now,
		},
		{
			name: "PhaseTransitioned",
			event: gameevt.NewPhaseTransitioned(
				42,
				"user1",
				now,
				sqlc.GamePhaseTypeDEPLOY,
				sqlc.GamePhaseTypeATTACK,
				1,
			),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "GameCompleted",
			event:     gameevt.NewGameCompleted(42, "winner", now, 10),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "GameCreated",
			event:     gameevt.NewGameCreated(42, now, 4),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "PlayerConnected",
			event:     gameevt.NewPlayerConnected(42, "user1", now),
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

func TestMoveExecuted_ToRecord_Attack(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	moveLog := sqlc.GameMoveLog{ID: 99, GameID: 42, PlayerID: 7}
	attackResult := &attack.MoveResult{
		AttackingRegionID: "brazil",
		DefendingRegionID: "argentina",
		ConqueringTroops:  3,
	}

	event := gameevt.NewMoveExecuted(
		42, "attacker", now,
		sqlc.GamePhaseTypeATTACK,
		moveLog,
		sqlc.GamePhaseTypeATTACK,
		false, 5,
		attackResult, nil,
	)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeMoveExecuted, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "attacker", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, sqlc.GamePhaseTypeATTACK, record["action_type"])
	require.Equal(t, sqlc.GamePhaseTypeATTACK, record["target_phase"])
	require.Equal(t, false, record["game_over"])
	require.Equal(t, int64(5), record["turn"])
	require.Equal(t, int64(99), record["move_log_id"])

	// Attack-specific fields
	require.Equal(t, "brazil", record["attacking_region_id"])
	require.Equal(t, "argentina", record["defending_region_id"])
	require.Equal(t, int64(3), record["conquering_troops"])

	// Cards-specific fields must be absent
	require.NotContains(t, record, "extra_deployable_troops")
	require.NotContains(t, record, "region_troop_grants")
}

func TestMoveExecuted_ToRecord_Cards(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	moveLog := sqlc.GameMoveLog{ID: 100, GameID: 42, PlayerID: 7}
	cardsResult := &cards.MoveResult{
		ExtraDeployableTroops: 6,
		RegionTroopGrants: []cards.RegionTroopGrant{
			{RegionID: 1, RegionExternalReference: "brazil"},
			{RegionID: 2, RegionExternalReference: "argentina"},
		},
	}

	event := gameevt.NewMoveExecuted(
		42, "player1", now,
		sqlc.GamePhaseTypeCARDS,
		moveLog,
		sqlc.GamePhaseTypeDEPLOY,
		false, 3,
		nil, cardsResult,
	)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeMoveExecuted, record["event_type"])
	require.Equal(t, sqlc.GamePhaseTypeCARDS, record["action_type"])
	require.Equal(t, int64(6), record["extra_deployable_troops"])
	require.Equal(t, 2, record["region_troop_grants"])

	// Attack-specific fields must be absent
	require.NotContains(t, record, "attacking_region_id")
	require.NotContains(t, record, "defending_region_id")
	require.NotContains(t, record, "conquering_troops")
}

func TestMoveExecuted_ToRecord_Deploy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewMoveExecuted(
		42, "player1", now,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GameMoveLog{ID: 101},
		sqlc.GamePhaseTypeDEPLOY,
		false, 2,
		nil, nil,
	)

	record := event.ToRecord()

	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, record["action_type"])

	// No action-specific keys
	require.NotContains(t, record, "attacking_region_id")
	require.NotContains(t, record, "defending_region_id")
	require.NotContains(t, record, "conquering_troops")
	require.NotContains(t, record, "extra_deployable_troops")
	require.NotContains(t, record, "region_troop_grants")
}

func TestMoveExecuted_ToRecord_Conquer(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewMoveExecuted(
		42, "player1", now,
		sqlc.GamePhaseTypeCONQUER,
		sqlc.GameMoveLog{ID: 102},
		sqlc.GamePhaseTypeCONQUER,
		false, 2,
		nil, nil,
	)

	record := event.ToRecord()

	require.Equal(t, sqlc.GamePhaseTypeCONQUER, record["action_type"])

	require.NotContains(t, record, "attacking_region_id")
	require.NotContains(t, record, "defending_region_id")
	require.NotContains(t, record, "conquering_troops")
	require.NotContains(t, record, "extra_deployable_troops")
	require.NotContains(t, record, "region_troop_grants")
}

func TestMoveExecuted_ToRecord_Reinforce(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewMoveExecuted(
		42, "player1", now,
		sqlc.GamePhaseTypeREINFORCE,
		sqlc.GameMoveLog{ID: 103},
		sqlc.GamePhaseTypeREINFORCE,
		false, 2,
		nil, nil,
	)

	record := event.ToRecord()

	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, record["action_type"])

	require.NotContains(t, record, "attacking_region_id")
	require.NotContains(t, record, "defending_region_id")
	require.NotContains(t, record, "conquering_troops")
	require.NotContains(t, record, "extra_deployable_troops")
	require.NotContains(t, record, "region_troop_grants")
}

func TestMoveExecuted_ToRecord_CommonFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	moveLog := sqlc.GameMoveLog{ID: 77, GameID: 42, PlayerID: 5}

	phases := []sqlc.GamePhaseType{
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GamePhaseTypeATTACK,
		sqlc.GamePhaseTypeCONQUER,
		sqlc.GamePhaseTypeREINFORCE,
		sqlc.GamePhaseTypeCARDS,
	}

	commonKeys := []string{
		"event_type", "game_id", "user_id", "timestamp",
		"action_type", "target_phase", "game_over", "turn", "move_log_id",
	}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			t.Parallel()

			event := gameevt.NewMoveExecuted(
				42, "player1", now,
				phase, moveLog, phase, false, 1,
				nil, nil,
			)

			record := event.ToRecord()

			for _, key := range commonKeys {
				require.Contains(
					t,
					record,
					key,
					"missing common key: %s for phase: %s",
					key,
					phase,
				)
			}
		})
	}
}

func TestPhaseTransitioned_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewPhaseTransitioned(
		42, "player1", now,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GamePhaseTypeATTACK,
		5,
	)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypePhaseTransitioned, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, record["from_phase"])
	require.Equal(t, sqlc.GamePhaseTypeATTACK, record["to_phase"])
	require.Equal(t, int64(5), record["turn"])
}

func TestGameCompleted_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewGameCompleted(42, "winner", now, 10)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeGameCompleted, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "winner", record["winner_user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, int64(10), record["turn"])
}

func TestGameCreated_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewGameCreated(42, now, 4)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeGameCreated, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, 4, record["num_players"])
}

func TestPlayerConnected_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewPlayerConnected(42, "player1", now)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypePlayerConnected, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
}

func TestMoveExecuted_ToRecord_NilFieldsOmitted(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	// Explicitly pass nil for both result pointers
	event := gameevt.NewMoveExecuted(
		42, "player1", now,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GameMoveLog{ID: 200},
		sqlc.GamePhaseTypeDEPLOY,
		false, 1,
		nil, nil,
	)

	record := event.ToRecord()

	// Attack-specific keys must not exist
	_, hasAttacking := record["attacking_region_id"]
	require.False(t, hasAttacking, "attacking_region_id should be absent when AttackResult is nil")

	_, hasDefending := record["defending_region_id"]
	require.False(t, hasDefending, "defending_region_id should be absent when AttackResult is nil")

	_, hasConquering := record["conquering_troops"]
	require.False(t, hasConquering, "conquering_troops should be absent when AttackResult is nil")

	// Cards-specific keys must not exist
	_, hasExtraTroops := record["extra_deployable_troops"]
	require.False(
		t,
		hasExtraTroops,
		"extra_deployable_troops should be absent when CardsResult is nil",
	)

	_, hasGrants := record["region_troop_grants"]
	require.False(t, hasGrants, "region_troop_grants should be absent when CardsResult is nil")
}
