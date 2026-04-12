package checker_test

import (
	"fmt"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission/checker"
	"github.com/stretchr/testify/require"
)

func TestEighteenTerritoriesChecker_Type(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewEighteenTerritoriesChecker()
	require.Equal(t, snapshot.MissionEighteenTerritoriesTwoTroops, missionChecker.Type())
}

func TestEighteenTerritoriesChecker_Check(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		regions  []snapshot.RegionState
		expected bool
	}{
		{
			name:     "zero regions",
			regions:  []snapshot.RegionState{},
			expected: false,
		},
		{
			name:     "17 regions with 2 troops is not enough",
			regions:  makeRegionsForPlayer("giovanni", 17, 2),
			expected: false,
		},
		{
			name:     "18 regions with 2 troops is enough",
			regions:  makeRegionsForPlayer("giovanni", 18, 2),
			expected: true,
		},
		{
			name: "18 regions but only 17 with 2 troops is not enough",
			regions: func() []snapshot.RegionState {
				r := makeRegionsForPlayer("giovanni", 18, 2)
				r[0].Troops = 1

				return r
			}(),
			expected: false,
		},
		{
			name: "19 regions but only 18 with 2 troops is enough",
			regions: func() []snapshot.RegionState {
				r := makeRegionsForPlayer("giovanni", 19, 2)
				r[0].Troops = 1

				return r
			}(),
			expected: true,
		},
		{
			name: "other player regions are not counted",
			regions: append(
				makeRegionsForPlayer("giovanni", 17, 2),
				makeRegionsForPlayer("opponent", 10, 5)...,
			),
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			missionChecker := checker.NewEighteenTerritoriesChecker()

			checkCtx := checker.CheckContext{
				Regions:       testCase.regions,
				CurrentUserID: "giovanni",
			}
			mission := snapshot.PlayerMission{
				Type:   snapshot.MissionEighteenTerritoriesTwoTroops,
				Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
			}

			result, err := missionChecker.Check(checkCtx, mission)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestTwentyFourTerritoriesChecker_Type(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewTwentyFourTerritoriesChecker()
	require.Equal(t, snapshot.MissionTwentyFourTerritories, missionChecker.Type())
}

func TestTwentyFourTerritoriesChecker_Check(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		regions  []snapshot.RegionState
		expected bool
	}{
		{
			name:     "23 regions is not enough",
			regions:  makeRegionsForPlayer("giovanni", 23, 1),
			expected: false,
		},
		{
			name:     "24 regions is enough",
			regions:  makeRegionsForPlayer("giovanni", 24, 1),
			expected: true,
		},
		{
			name:     "25 regions is enough",
			regions:  makeRegionsForPlayer("giovanni", 25, 1),
			expected: true,
		},
		{
			name: "other player regions are not counted",
			regions: append(
				makeRegionsForPlayer("giovanni", 23, 1),
				makeRegionsForPlayer("opponent", 5, 1)...,
			),
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			missionChecker := checker.NewTwentyFourTerritoriesChecker()

			checkCtx := checker.CheckContext{
				Regions:       testCase.regions,
				CurrentUserID: "giovanni",
			}
			mission := snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			}

			result, err := missionChecker.Check(checkCtx, mission)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestEliminatePlayerChecker_Type(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewEliminatePlayerChecker()
	require.Equal(t, snapshot.MissionEliminatePlayer, missionChecker.Type())
}

func TestEliminatePlayerChecker_Check(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		regions  []snapshot.RegionState
		target   string
		expected bool
	}{
		{
			name:     "target has no regions",
			regions:  makeRegionsForPlayer("giovanni", 10, 1),
			target:   "opponent",
			expected: true,
		},
		{
			name: "target has one region",
			regions: append(
				makeRegionsForPlayer("giovanni", 10, 1),
				snapshot.RegionState{ID: "target_r1", OwnerID: "opponent", Troops: 1},
			),
			target:   "opponent",
			expected: false,
		},
		{
			name: "target has multiple regions",
			regions: append(
				makeRegionsForPlayer("giovanni", 10, 1),
				snapshot.RegionState{ID: "target_r1", OwnerID: "opponent", Troops: 1},
				snapshot.RegionState{ID: "target_r2", OwnerID: "opponent", Troops: 2},
			),
			target:   "opponent",
			expected: false,
		},
		{
			name:     "empty board",
			regions:  []snapshot.RegionState{},
			target:   "opponent",
			expected: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			missionChecker := checker.NewEliminatePlayerChecker()

			checkCtx := checker.CheckContext{
				Regions:       testCase.regions,
				CurrentUserID: "giovanni",
			}
			mission := snapshot.PlayerMission{
				Type: snapshot.MissionEliminatePlayer,
				Detail: snapshot.EliminatePlayerMission{
					TargetUserID: testCase.target,
				},
			}

			result, err := missionChecker.Check(checkCtx, mission)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestEliminatePlayerChecker_WrongDetailType(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewEliminatePlayerChecker()

	checkCtx := checker.CheckContext{
		Regions:       []snapshot.RegionState{},
		CurrentUserID: "giovanni",
	}
	mission := snapshot.PlayerMission{
		Type:   snapshot.MissionEliminatePlayer,
		Detail: snapshot.TwoContinentsMission{Continent1: "asia", Continent2: "europe"},
	}

	_, err := missionChecker.Check(checkCtx, mission)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected EliminatePlayerMission detail")
}

func TestTwoContinentsChecker_Type(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewTwoContinentsChecker()
	require.Equal(t, snapshot.MissionTwoContinents, missionChecker.Type())
}

func TestTwoContinentsChecker_Check(t *testing.T) {
	t.Parallel()

	continents := testContinents()

	tests := []struct {
		name       string
		regions    []snapshot.RegionState
		continent1 string
		continent2 string
		expected   bool
	}{
		{
			name:       "controls neither continent",
			regions:    []snapshot.RegionState{},
			continent1: "north_america",
			continent2: "africa",
			expected:   false,
		},
		{
			name: "controls only first continent",
			regions: regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "opponent", nil)),
			continent1: "north_america",
			continent2: "africa",
			expected:   false,
		},
		{
			name: "controls both continents",
			regions: regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "giovanni", nil)),
			continent1: "north_america",
			continent2: "africa",
			expected:   true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			missionChecker := checker.NewTwoContinentsChecker()

			checkCtx := checker.CheckContext{
				Regions:       testCase.regions,
				Continents:    continents,
				CurrentUserID: "giovanni",
			}
			mission := snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinents,
				Detail: snapshot.TwoContinentsMission{
					Continent1: testCase.continent1,
					Continent2: testCase.continent2,
				},
			}

			result, err := missionChecker.Check(checkCtx, mission)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestTwoContinentsChecker_WrongDetailType(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewTwoContinentsChecker()

	checkCtx := checker.CheckContext{
		Regions:       []snapshot.RegionState{},
		Continents:    testContinents(),
		CurrentUserID: "giovanni",
	}
	mission := snapshot.PlayerMission{
		Type:   snapshot.MissionTwoContinents,
		Detail: snapshot.EliminatePlayerMission{TargetUserID: "someone"},
	}

	_, err := missionChecker.Check(checkCtx, mission)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected TwoContinentsMission detail")
}

func TestTwoContinentsPlusOneChecker_Type(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewTwoContinentsPlusOneChecker()
	require.Equal(t, snapshot.MissionTwoContinentsPlusOne, missionChecker.Type())
}

func TestTwoContinentsPlusOneChecker_Check(t *testing.T) {
	t.Parallel()

	continents := testContinents()

	tests := []struct {
		name       string
		regions    []snapshot.RegionState
		continent1 string
		continent2 string
		expected   bool
	}{
		{
			name:       "controls neither continent",
			regions:    []snapshot.RegionState{},
			continent1: "north_america",
			continent2: "africa",
			expected:   false,
		},
		{
			name: "controls both mandatory but no third",
			regions: regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "giovanni", nil)),
			continent1: "north_america",
			continent2: "africa",
			expected:   false,
		},
		{
			name: "controls both mandatory plus a third",
			regions: regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "giovanni",
					regionsForContinent(continents, "south_america", "giovanni", nil))),
			continent1: "north_america",
			continent2: "africa",
			expected:   true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			missionChecker := checker.NewTwoContinentsPlusOneChecker()

			checkCtx := checker.CheckContext{
				Regions:       testCase.regions,
				Continents:    continents,
				CurrentUserID: "giovanni",
			}
			mission := snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinentsPlusOne,
				Detail: snapshot.TwoContinentsPlusOneMission{
					Continent1: testCase.continent1,
					Continent2: testCase.continent2,
				},
			}

			result, err := missionChecker.Check(checkCtx, mission)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestTwoContinentsPlusOneChecker_WrongDetailType(t *testing.T) {
	t.Parallel()
	missionChecker := checker.NewTwoContinentsPlusOneChecker()

	checkCtx := checker.CheckContext{
		Regions:       []snapshot.RegionState{},
		Continents:    testContinents(),
		CurrentUserID: "giovanni",
	}
	mission := snapshot.PlayerMission{
		Type:   snapshot.MissionTwoContinentsPlusOne,
		Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
	}

	_, err := missionChecker.Check(checkCtx, mission)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected TwoContinentsPlusOneMission detail")
}

func TestRegistry_GetChecker_UnknownType(t *testing.T) {
	t.Parallel()

	registry, err := checker.NewRegistry([]checker.MissionChecker{
		checker.NewEighteenTerritoriesChecker(),
	})
	require.NoError(t, err)

	_, err = registry.GetChecker("nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown mission type")
}

func TestRegistry_DuplicateChecker(t *testing.T) {
	t.Parallel()

	_, err := checker.NewRegistry([]checker.MissionChecker{
		checker.NewEighteenTerritoriesChecker(),
		checker.NewEighteenTerritoriesChecker(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate checker registered")
}

// --- Test helpers ---

func makeRegionsForPlayer(
	owner string,
	count int,
	troops int64,
) []snapshot.RegionState {
	regions := make([]snapshot.RegionState, count)
	for i := range regions {
		regions[i] = snapshot.RegionState{
			ID:      fmt.Sprintf("%s_region_%d", owner, i+1),
			OwnerID: owner,
			Troops:  troops,
		}
	}

	return regions
}

// testContinents builds a minimal Continents for testing.
func testContinents() board.Continents {
	boardDto := &board.BoardDto{
		Regions: []board.RegionDto{
			{ExternalReference: "na_1", Continent: "north_america"},
			{ExternalReference: "na_2", Continent: "north_america"},
			{ExternalReference: "na_3", Continent: "north_america"},
			{ExternalReference: "af_1", Continent: "africa"},
			{ExternalReference: "af_2", Continent: "africa"},
			{ExternalReference: "sa_1", Continent: "south_america"},
			{ExternalReference: "sa_2", Continent: "south_america"},
		},
		Continents: []board.ContinentDto{
			{ExternalReference: "north_america", BonusTroops: 5},
			{ExternalReference: "africa", BonusTroops: 3},
			{ExternalReference: "south_america", BonusTroops: 2},
		},
	}

	continents, err := board.NewContinents(boardDto)
	if err != nil {
		panic(fmt.Sprintf("failed to create test continents: %v", err))
	}

	return continents
}

// regionsForContinent creates RegionState entries for all regions in a
// continent, assigning them to the given owner. Appends to existing if
// provided.
func regionsForContinent(
	continents board.Continents,
	continentName string,
	owner string,
	existing []snapshot.RegionState,
) []snapshot.RegionState {
	for _, c := range continents.All() {
		if c.ExternalReference == continentName {
			for _, r := range c.Regions() {
				existing = append(existing, snapshot.RegionState{
					ID:      r,
					OwnerID: owner,
					Troops:  1,
				})
			}
		}
	}

	return existing
}
