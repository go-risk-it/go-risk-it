package orchestration_test

import (
	"testing"

	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/stretchr/testify/require"
)

func TestMapSnapshotPhaseToSqlc_AllPhases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    apisnapshot.PhaseType
		expected sqlc.GamePhaseType
	}{
		{"cards", apisnapshot.PhaseCards, sqlc.GamePhaseTypeCARDS},
		{"deploy", apisnapshot.PhaseDeploy, sqlc.GamePhaseTypeDEPLOY},
		{"attack", apisnapshot.PhaseAttack, sqlc.GamePhaseTypeATTACK},
		{"conquer", apisnapshot.PhaseConquer, sqlc.GamePhaseTypeCONQUER},
		{"reinforce", apisnapshot.PhaseReinforce, sqlc.GamePhaseTypeREINFORCE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := orchestration.MapSnapshotPhaseToSqlc(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestMapSnapshotPhaseToSqlc_UnknownPanics(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		orchestration.MapSnapshotPhaseToSqlc("unknown_phase")
	})
}

func TestGameStateFromCache(t *testing.T) {
	t.Parallel()

	cached := &apisnapshot.CachedGameState{
		Turn: 7,
		PublicSnapshot: &apisnapshot.GameSnapshot{
			Game: apisnapshot.GameMeta{
				ID:           42,
				Turn:         7,
				WinnerUserID: "winner1",
			},
			Phase: apisnapshot.Phase{
				Type:  apisnapshot.PhaseAttack,
				State: apisnapshot.EmptyPhaseState{},
			},
		},
	}

	result := orchestration.GameStateFromCache(cached)

	require.Equal(t, int64(42), result.ID)
	require.Equal(t, int64(7), result.Turn)
	require.Equal(t, sqlc.GamePhaseTypeATTACK, result.Phase)
	require.Equal(t, "winner1", result.WinnerUserID)
}

func TestGameStateFromCache_NoWinner(t *testing.T) {
	t.Parallel()

	cached := &apisnapshot.CachedGameState{
		Turn: 3,
		PublicSnapshot: &apisnapshot.GameSnapshot{
			Game: apisnapshot.GameMeta{
				ID: 10,
			},
			Phase: apisnapshot.Phase{
				Type:  apisnapshot.PhaseDeploy,
				State: apisnapshot.EmptyPhaseState{},
			},
		},
	}

	result := orchestration.GameStateFromCache(cached)

	require.Equal(t, int64(10), result.ID)
	require.Equal(t, int64(3), result.Turn)
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, result.Phase)
	require.Empty(t, result.WinnerUserID)
}
