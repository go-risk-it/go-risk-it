package snapshot_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/stretchr/testify/require"
)

func TestPlayerMission_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mission snapshot.PlayerMission
	}{
		{
			name: "two continents",
			mission: snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinents,
				Detail: snapshot.TwoContinentsMission{
					Continent1: "europe",
					Continent2: "asia",
				},
			},
		},
		{
			name: "two continents plus one",
			mission: snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinentsPlusOne,
				Detail: snapshot.TwoContinentsPlusOneMission{
					Continent1: "north_america",
					Continent2: "south_america",
				},
			},
		},
		{
			name: "eighteen territories two troops",
			mission: snapshot.PlayerMission{
				Type:   snapshot.MissionEighteenTerritoriesTwoTroops,
				Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
			},
		},
		{
			name: "twenty four territories",
			mission: snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		},
		{
			name: "eliminate player",
			mission: snapshot.PlayerMission{
				Type: snapshot.MissionEliminatePlayer,
				Detail: snapshot.EliminatePlayerMission{
					TargetUserID: "target-user-42",
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(testCase.mission)
			require.NoError(t, err)

			var got snapshot.PlayerMission
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			require.Equal(t, testCase.mission.Type, got.Type)
			require.Equal(t, testCase.mission.Detail, got.Detail)
		})
	}
}

func TestPlayerMission_WireCompatibility(t *testing.T) {
	t.Parallel()

	// The sealed PlayerMission wrapper must produce identical JSON to a
	// generic MissionState[T]{Type, Details} struct. This is the pattern
	// used in messaging/mission_state.go today.
	tests := []struct {
		name          string
		anonymousJSON any
		sealedMission snapshot.PlayerMission
	}{
		{
			name: "two continents",
			anonymousJSON: struct {
				Type    snapshot.MissionType `json:"type"`
				Details any                  `json:"details"`
			}{
				Type: snapshot.MissionTwoContinents,
				Details: snapshot.TwoContinentsMission{
					Continent1: "europe",
					Continent2: "asia",
				},
			},
			sealedMission: snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinents,
				Detail: snapshot.TwoContinentsMission{
					Continent1: "europe",
					Continent2: "asia",
				},
			},
		},
		{
			name: "two continents plus one",
			anonymousJSON: struct {
				Type    snapshot.MissionType `json:"type"`
				Details any                  `json:"details"`
			}{
				Type: snapshot.MissionTwoContinentsPlusOne,
				Details: snapshot.TwoContinentsPlusOneMission{
					Continent1: "north_america",
					Continent2: "south_america",
				},
			},
			sealedMission: snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinentsPlusOne,
				Detail: snapshot.TwoContinentsPlusOneMission{
					Continent1: "north_america",
					Continent2: "south_america",
				},
			},
		},
		{
			name: "eighteen territories (empty details)",
			anonymousJSON: struct {
				Type    snapshot.MissionType `json:"type"`
				Details any                  `json:"details"`
			}{
				Type:    snapshot.MissionEighteenTerritoriesTwoTroops,
				Details: snapshot.EighteenTerritoriesTwoTroopsMission{},
			},
			sealedMission: snapshot.PlayerMission{
				Type:   snapshot.MissionEighteenTerritoriesTwoTroops,
				Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
			},
		},
		{
			name: "twenty four territories (empty details)",
			anonymousJSON: struct {
				Type    snapshot.MissionType `json:"type"`
				Details any                  `json:"details"`
			}{
				Type:    snapshot.MissionTwentyFourTerritories,
				Details: snapshot.TwentyFourTerritoriesMission{},
			},
			sealedMission: snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		},
		{
			name: "eliminate player",
			anonymousJSON: struct {
				Type    snapshot.MissionType `json:"type"`
				Details any                  `json:"details"`
			}{
				Type: snapshot.MissionEliminatePlayer,
				Details: snapshot.EliminatePlayerMission{
					TargetUserID: "target-user-42",
				},
			},
			sealedMission: snapshot.PlayerMission{
				Type: snapshot.MissionEliminatePlayer,
				Detail: snapshot.EliminatePlayerMission{
					TargetUserID: "target-user-42",
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			existingJSON, err := json.Marshal(testCase.anonymousJSON)
			require.NoError(t, err)

			sealedJSON, err := json.Marshal(testCase.sealedMission)
			require.NoError(t, err)

			require.JSONEq(t, string(existingJSON), string(sealedJSON),
				"wire format mismatch: existing=%s sealed=%s", existingJSON, sealedJSON)
		})
	}
}

func TestPlayerMission_UnknownType(t *testing.T) {
	t.Parallel()

	data := []byte(`{"type":"unknown_mission","details":{}}`)

	var got snapshot.PlayerMission
	err := json.Unmarshal(data, &got)
	require.Error(t, err)
	require.ErrorContains(t, err, `unknown mission type: "unknown_mission"`)
}

func TestPlayerMission_MalformedJSON(t *testing.T) {
	t.Parallel()

	t.Run("truncated JSON object", func(t *testing.T) {
		t.Parallel()

		var got snapshot.PlayerMission
		err := json.Unmarshal([]byte(`{"type":"twoContinents","details":`), &got)
		require.Error(t, err)
	})

	t.Run("invalid field type in two continents", func(t *testing.T) {
		t.Parallel()

		var got snapshot.PlayerMission
		err := json.Unmarshal(
			[]byte(`{"type":"twoContinents","details":{"continent1":123}}`),
			&got,
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "unmarshaling two continents mission")
	})

	t.Run("invalid field type in eliminate player", func(t *testing.T) {
		t.Parallel()

		var got snapshot.PlayerMission
		err := json.Unmarshal(
			[]byte(`{"type":"eliminatePlayer","details":{"targetUserId":999}}`),
			&got,
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "unmarshaling eliminate player mission")
	})
}
