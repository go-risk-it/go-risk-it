package snapshot_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/stretchr/testify/require"
)

func TestGameSnapshot_JSON(t *testing.T) {
	t.Parallel()

	original := snapshot.GameSnapshot{
		Game: snapshot.GameMeta{
			ID:           42,
			Turn:         3,
			WinnerUserID: "",
		},
		Phase: snapshot.Phase{
			Type:  snapshot.PhaseDeploy,
			State: snapshot.DeployPhaseState{DeployableTroops: 7},
		},
		Regions: []snapshot.RegionState{
			{ID: "alaska", OwnerID: "user-1", Troops: 5},
			{ID: "kamchatka", OwnerID: "user-2", Troops: 3},
		},
		Players: []snapshot.PlayerState{
			{
				UserID:    "user-1",
				Name:      "Alice",
				Index:     0,
				CardCount: 2,
				Status:    snapshot.PlayerAlive,
			},
			{
				UserID:    "user-2",
				Name:      "Bob",
				Index:     1,
				CardCount: 0,
				Status:    snapshot.PlayerDead,
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got snapshot.GameSnapshot
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	// Verify top-level structure
	require.Equal(t, original.Game, got.Game)
	require.Equal(t, original.Phase.Type, got.Phase.Type)
	require.Equal(t, original.Phase.State, got.Phase.State)
	require.Equal(t, original.Regions, got.Regions)
	require.Equal(t, original.Players, got.Players)
}

func TestGameSnapshot_JSON_ConquerPhase(t *testing.T) {
	t.Parallel()

	original := snapshot.GameSnapshot{
		Game: snapshot.GameMeta{
			ID:           99,
			Turn:         7,
			WinnerUserID: "user-1",
		},
		Phase: snapshot.Phase{
			Type: snapshot.PhaseConquer,
			State: snapshot.ConquerPhaseState{
				AttackingRegionID: "alaska",
				DefendingRegionID: "kamchatka",
				MinTroopsToMove:   2,
			},
		},
		Regions: []snapshot.RegionState{
			{ID: "alaska", OwnerID: "user-1", Troops: 10},
		},
		Players: []snapshot.PlayerState{
			{
				UserID:    "user-1",
				Name:      "Alice",
				Index:     0,
				CardCount: 3,
				Status:    snapshot.PlayerAlive,
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got snapshot.GameSnapshot
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Equal(t, original, got)
}

func TestGameSnapshot_JSON_EmptyPhase(t *testing.T) {
	t.Parallel()

	original := snapshot.GameSnapshot{
		Game: snapshot.GameMeta{
			ID:   1,
			Turn: 1,
		},
		Phase: snapshot.Phase{
			Type:  snapshot.PhaseCards,
			State: snapshot.EmptyPhaseState{},
		},
		Regions: []snapshot.RegionState{},
		Players: []snapshot.PlayerState{},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got snapshot.GameSnapshot
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Equal(t, original.Game, got.Game)
	require.Equal(t, original.Phase.Type, got.Phase.Type)
	require.Equal(t, original.Phase.State, got.Phase.State)
}

func TestPlayerPrivate_JSON(t *testing.T) {
	t.Parallel()

	original := snapshot.PlayerPrivate{
		Cards: []snapshot.CardState{
			{ID: 1, Type: snapshot.CardInfantry, Region: "alaska"},
			{ID: 2, Type: snapshot.CardCavalry, Region: "brazil"},
			{ID: 3, Type: snapshot.CardJolly, Region: ""},
		},
		Mission: snapshot.PlayerMission{
			Type: snapshot.MissionTwoContinents,
			Detail: snapshot.TwoContinentsMission{
				Continent1: "europe",
				Continent2: "asia",
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got snapshot.PlayerPrivate
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Equal(t, original.Cards, got.Cards)
	require.Equal(t, original.Mission.Type, got.Mission.Type)
	require.Equal(t, original.Mission.Detail, got.Mission.Detail)
}

func TestPlayerPrivate_JSON_EmptyMission(t *testing.T) {
	t.Parallel()

	original := snapshot.PlayerPrivate{
		Cards: []snapshot.CardState{
			{ID: 5, Type: snapshot.CardArtillery, Region: "congo"},
		},
		Mission: snapshot.PlayerMission{
			Type:   snapshot.MissionTwentyFourTerritories,
			Detail: snapshot.TwentyFourTerritoriesMission{},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got snapshot.PlayerPrivate
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Equal(t, original, got)
}
