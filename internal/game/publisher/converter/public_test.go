package converter_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/publisher/converter"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestConvertPublicSnapshot_DeployPhase(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeDEPLOY,
			Turn:         1,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type:        sqlc.GamePhaseTypeDEPLOY,
			DeployState: &snapshot.DeployState{DeployableTroops: 7},
		},
		Board: []sqlc.GetRegionsByGameRow{
			{ID: 1, ExternalReference: "alaska", Troops: 3, UserID: "user-1"},
			{ID: 2, ExternalReference: "brazil", Troops: 5, UserID: "user-2"},
		},
		Players: []sqlc.GetPlayersStateRow{
			{UserID: "user-1", Name: "Alice", TurnIndex: 0, CardCount: 2, RegionCount: 1},
			{UserID: "user-2", Name: "Bob", TurnIndex: 1, CardCount: 1, RegionCount: 1},
		},
	}

	result, err := converter.ConvertPublicSnapshot(snap, []string{"user-1"})
	require.NoError(t, err)

	// Verify gameState is non-nil (polymorphic any — tested via JSON in publisher tests)
	require.NotNil(t, result.GameState)

	// Verify boardState
	require.Len(t, result.BoardState.Regions, 2)
	require.Equal(t, "alaska", result.BoardState.Regions[0].ID)
	require.Equal(t, "user-1", result.BoardState.Regions[0].OwnerID)
	require.Equal(t, int64(3), result.BoardState.Regions[0].Troops)
	require.Equal(t, "brazil", result.BoardState.Regions[1].ID)
	require.Equal(t, "user-2", result.BoardState.Regions[1].OwnerID)
	require.Equal(t, int64(5), result.BoardState.Regions[1].Troops)

	// Verify playerState
	require.Len(t, result.PlayerState.Players, 2)
	require.Equal(t, "user-1", result.PlayerState.Players[0].UserID)
	require.Equal(t, "Alice", result.PlayerState.Players[0].Name)
	require.Equal(t, int64(0), result.PlayerState.Players[0].Index)
	require.Equal(t, int64(2), result.PlayerState.Players[0].CardCount)
	require.Equal(t, messaging.Alive, result.PlayerState.Players[0].Status)
	require.Equal(t, messaging.Connected, result.PlayerState.Players[0].ConnectionStatus)

	require.Equal(t, "user-2", result.PlayerState.Players[1].UserID)
	require.Equal(t, messaging.Disconnected, result.PlayerState.Players[1].ConnectionStatus)
}

func TestConvertPublicSnapshot_AttackPhase(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeATTACK,
			Turn:         2,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeATTACK,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	result, err := converter.ConvertPublicSnapshot(snap, nil)
	require.NoError(t, err)
	require.NotNil(t, result.GameState)
}

func TestConvertPublicSnapshot_ConquerPhase(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeCONQUER,
			Turn:         3,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeCONQUER,
			ConquerState: &sqlc.GetConquerPhaseStateRow{
				SourceRegion:  "alaska",
				TargetRegion:  "kamchatka",
				MinimumTroops: 2,
			},
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	result, err := converter.ConvertPublicSnapshot(snap, nil)
	require.NoError(t, err)
	require.NotNil(t, result.GameState)
}

func TestConvertPublicSnapshot_ReinforcePhase(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeREINFORCE,
			Turn:         4,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeREINFORCE,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	result, err := converter.ConvertPublicSnapshot(snap, nil)
	require.NoError(t, err)
	require.NotNil(t, result.GameState)
}

func TestConvertPublicSnapshot_CardsPhase(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeCARDS,
			Turn:         1,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeCARDS,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	result, err := converter.ConvertPublicSnapshot(snap, nil)
	require.NoError(t, err)
	require.NotNil(t, result.GameState)
}

func TestConvertPublicSnapshot_WinnerUserID(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeATTACK,
			Turn:         10,
			WinnerUserID: pgtype.Text{String: "winner-user", Valid: true},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeATTACK,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	result, err := converter.ConvertPublicSnapshot(snap, nil)
	require.NoError(t, err)
	require.NotNil(t, result.GameState)
}

func TestConvertPublicSnapshot_NullWinnerUserID(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeATTACK,
			Turn:         1,
			WinnerUserID: pgtype.Text{Valid: false},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeATTACK,
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	result, err := converter.ConvertPublicSnapshot(snap, nil)
	require.NoError(t, err)
	require.NotNil(t, result.GameState)
}

func TestConvertPublicSnapshot_DeadPlayer(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeATTACK,
			Turn:         5,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type: sqlc.GamePhaseTypeATTACK,
		},
		Board: []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{
			{UserID: "dead-user", Name: "DeadGuy", TurnIndex: 0, CardCount: 0, RegionCount: 0},
		},
	}

	result, err := converter.ConvertPublicSnapshot(snap, []string{"dead-user"})
	require.NoError(t, err)

	require.Len(t, result.PlayerState.Players, 1)
	require.Equal(t, messaging.Dead, result.PlayerState.Players[0].Status)
	require.Equal(t, messaging.Connected, result.PlayerState.Players[0].ConnectionStatus)
}

func TestConvertPublicSnapshot_DeployPhaseNilState(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeDEPLOY,
			Turn:         1,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type:        sqlc.GamePhaseTypeDEPLOY,
			DeployState: nil, // invariant violation
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	_, err := converter.ConvertPublicSnapshot(snap, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "deploy phase state is nil")
}

func TestConvertPublicSnapshot_ConquerPhaseNilState(t *testing.T) {
	t.Parallel()

	snap := &snapshot.PublicSnapshot{
		Game: sqlc.GetGameRow{
			ID:           42,
			CurrentPhase: sqlc.GamePhaseTypeCONQUER,
			Turn:         1,
			WinnerUserID: pgtype.Text{},
		},
		Phase: snapshot.PhaseState{
			Type:         sqlc.GamePhaseTypeCONQUER,
			ConquerState: nil, // invariant violation
		},
		Board:   []sqlc.GetRegionsByGameRow{},
		Players: []sqlc.GetPlayersStateRow{},
	}

	_, err := converter.ConvertPublicSnapshot(snap, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "conquer phase state is nil")
}
