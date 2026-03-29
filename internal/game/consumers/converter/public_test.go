package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/consumers/converter"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
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

	// Verify gameState message
	gameMsg := unmarshalMessage(t, result.GameState)
	require.Equal(t, "gameState", gameMsg.Type)

	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))
	require.InDelta(t, float64(42), gameData["id"], 0)
	require.InDelta(t, float64(1), gameData["turn"], 0)
	require.Empty(t, gameData["winnerUserId"])

	phase, ok := gameData["phase"].(map[string]any) //nolint:varnamelen
	require.True(t, ok)
	require.Equal(t, "deploy", phase["type"])

	phaseState, ok := phase["state"].(map[string]any)
	require.True(t, ok)
	require.InDelta(t, float64(7), phaseState["deployableTroops"], 0)

	// Verify boardState message
	boardMsg := unmarshalMessage(t, result.BoardState)
	require.Equal(t, "boardState", boardMsg.Type)

	var boardData map[string]any
	require.NoError(t, json.Unmarshal(boardMsg.Payload, &boardData))

	regions, ok := boardData["regions"].([]any)
	require.True(t, ok)
	require.Len(t, regions, 2)

	region0, ok := regions[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "alaska", region0["id"])
	require.Equal(t, "user-1", region0["ownerId"])
	require.InDelta(t, float64(3), region0["troops"], 0)

	// Verify playerState message
	playerMsg := unmarshalMessage(t, result.PlayerState)
	require.Equal(t, "playerState", playerMsg.Type)

	var playerData map[string]any
	require.NoError(t, json.Unmarshal(playerMsg.Payload, &playerData))

	players, ok := playerData["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, 2)

	player0, ok := players[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "user-1", player0["userId"])
	require.Equal(t, "Alice", player0["name"])
	require.InDelta(t, float64(0), player0["index"], 0)
	require.InDelta(t, float64(2), player0["cardCount"], 0)
	require.Equal(t, "alive", player0["status"])
	require.Equal(t, "connected", player0["connectionStatus"])

	player1, ok := players[1].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "user-2", player1["userId"])
	require.Equal(t, "disconnected", player1["connectionStatus"])
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

	gameMsg := unmarshalMessage(t, result.GameState)
	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))

	phase, ok := gameData["phase"].(map[string]any) //nolint:varnamelen
	require.True(t, ok)
	require.Equal(t, "attack", phase["type"])

	// EmptyState serializes as {}
	phaseState, ok := phase["state"].(map[string]any)
	require.True(t, ok)
	require.Empty(t, phaseState)
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

	gameMsg := unmarshalMessage(t, result.GameState)
	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))

	phase, ok := gameData["phase"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "conquer", phase["type"])

	phaseState, ok := phase["state"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "alaska", phaseState["attackingRegionId"])
	require.Equal(t, "kamchatka", phaseState["defendingRegionId"])
	require.InDelta(t, float64(2), phaseState["minTroopsToMove"], 0)
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

	gameMsg := unmarshalMessage(t, result.GameState)
	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))

	phase, ok := gameData["phase"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "reinforce", phase["type"])
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

	gameMsg := unmarshalMessage(t, result.GameState)
	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))

	phase, ok := gameData["phase"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "cards", phase["type"])
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

	gameMsg := unmarshalMessage(t, result.GameState)
	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))
	require.Equal(t, "winner-user", gameData["winnerUserId"])
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

	gameMsg := unmarshalMessage(t, result.GameState)
	var gameData map[string]any
	require.NoError(t, json.Unmarshal(gameMsg.Payload, &gameData))
	require.Empty(t, gameData["winnerUserId"])
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

	playerMsg := unmarshalMessage(t, result.PlayerState)
	var playerData map[string]any
	require.NoError(t, json.Unmarshal(playerMsg.Payload, &playerData))

	players, ok := playerData["players"].([]any)
	require.True(t, ok)
	player0, ok := players[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "dead", player0["status"])
	require.Equal(t, "connected", player0["connectionStatus"])
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

// messageEnvelope is used to parse the outer WS message envelope.
type messageEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"data"`
}

func unmarshalMessage(t *testing.T, raw json.RawMessage) messageEnvelope {
	t.Helper()

	var msg messageEnvelope
	require.NoError(t, json.Unmarshal(raw, &msg))

	return msg
}
