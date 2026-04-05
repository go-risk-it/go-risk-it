package snapshot_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/stretchr/testify/require"
)

func TestBuildPlayerView_MergesPublicAndPrivate(t *testing.T) {
	t.Parallel()

	public := &snapshot.GameSnapshot{
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
			{ID: "brazil", OwnerID: "user-2", Troops: 3},
		},
		Players: []snapshot.PlayerState{
			{
				UserID:    "user-1",
				Name:      "Alice",
				Index:     0,
				CardCount: 2,
				Status:    snapshot.PlayerAlive,
			},
			{UserID: "user-2", Name: "Bob", Index: 1, CardCount: 1, Status: snapshot.PlayerAlive},
		},
	}

	private := &snapshot.PlayerPrivate{
		Cards: []snapshot.CardState{
			{ID: 1, Type: snapshot.CardInfantry, Region: "alaska"},
			{ID: 2, Type: snapshot.CardCavalry, Region: "brazil"},
		},
		Mission: snapshot.PlayerMission{
			Type: snapshot.MissionTwoContinents,
			Detail: snapshot.TwoContinentsMission{
				Continent1: "europe",
				Continent2: "asia",
			},
		},
	}

	view := snapshot.BuildPlayerView(public, private)

	// Public fields come from GameSnapshot
	require.Equal(t, public.Game, view.Game)
	require.Equal(t, public.Phase, view.Phase)
	require.Equal(t, public.Regions, view.Regions)
	require.Equal(t, public.Players, view.Players)

	// Private fields come from PlayerPrivate
	require.Equal(t, private.Cards, view.Cards)
	require.Equal(t, private.Mission, view.Mission)
}

func TestPlayerView_JSON_RoundTrip(t *testing.T) {
	t.Parallel()

	original := snapshot.PlayerView{
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
		Cards: []snapshot.CardState{
			{ID: 5, Type: snapshot.CardArtillery, Region: "congo"},
			{ID: 6, Type: snapshot.CardJolly, Region: ""},
		},
		Mission: snapshot.PlayerMission{
			Type: snapshot.MissionEliminatePlayer,
			Detail: snapshot.EliminatePlayerMission{
				TargetUserID: "user-2",
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got snapshot.PlayerView
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Equal(t, original.Game, got.Game)
	require.Equal(t, original.Phase.Type, got.Phase.Type)
	require.Equal(t, original.Phase.State, got.Phase.State)
	require.Equal(t, original.Regions, got.Regions)
	require.Equal(t, original.Players, got.Players)
	require.Equal(t, original.Cards, got.Cards)
	require.Equal(t, original.Mission.Type, got.Mission.Type)
	require.Equal(t, original.Mission.Detail, got.Mission.Detail)
}

func TestPlayerView_AllFieldsPresent(t *testing.T) {
	t.Parallel()

	public := &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{
			ID:   1,
			Turn: 1,
		},
		Phase: snapshot.Phase{
			Type:  snapshot.PhaseAttack,
			State: snapshot.EmptyPhaseState{},
		},
		Regions: []snapshot.RegionState{
			{ID: "alaska", OwnerID: "user-1", Troops: 1},
		},
		Players: []snapshot.PlayerState{
			{
				UserID:    "user-1",
				Name:      "Alice",
				Index:     0,
				CardCount: 0,
				Status:    snapshot.PlayerAlive,
			},
		},
	}

	private := &snapshot.PlayerPrivate{
		Cards: []snapshot.CardState{
			{ID: 10, Type: snapshot.CardInfantry, Region: "alaska"},
		},
		Mission: snapshot.PlayerMission{
			Type:   snapshot.MissionTwentyFourTerritories,
			Detail: snapshot.TwentyFourTerritoriesMission{},
		},
	}

	view := snapshot.BuildPlayerView(public, private)

	// All 6 fields must be non-zero
	require.NotEmpty(t, view.Game)
	require.NotEmpty(t, view.Phase.Type)
	require.NotNil(t, view.Phase.State)
	require.NotEmpty(t, view.Regions)
	require.NotEmpty(t, view.Players)
	require.NotEmpty(t, view.Cards)
	require.NotEmpty(t, view.Mission.Type)

	// Marshal and check no omitempty fields are dropped
	data, err := json.Marshal(view)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	expectedFields := []string{"game", "phase", "regions", "players", "cards", "mission"}
	for _, field := range expectedFields {
		require.Contains(t, raw, field, "PlayerView JSON must always include %q", field)
	}

	require.Len(t, raw, len(expectedFields), "PlayerView JSON must have exactly 6 fields")
}

func TestPlayerView_JSON_NoOmitEmpty(t *testing.T) {
	t.Parallel()

	// Build a PlayerView with empty slices — they should serialize as [] not be omitted
	view := snapshot.PlayerView{
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
		Cards:   []snapshot.CardState{},
		Mission: snapshot.PlayerMission{
			Type:   snapshot.MissionTwentyFourTerritories,
			Detail: snapshot.TwentyFourTerritoriesMission{},
		},
	}

	data, err := json.Marshal(view)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	// All 6 fields must be present even when slices are empty
	expectedFields := []string{"game", "phase", "regions", "players", "cards", "mission"}
	for _, field := range expectedFields {
		require.Contains(
			t,
			raw,
			field,
			"PlayerView JSON must always include %q even when empty",
			field,
		)
	}
}
