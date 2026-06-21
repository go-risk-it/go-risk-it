package mission_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	boardpkg "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission/checker"
	mockdb "github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	mockrand "github.com/go-risk-it/go-risk-it/internal/game/testmocks/rand"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) mission.Service {
	t.Helper()
	querier := mockdb.NewQuerier(t)
	rng := mockrand.NewRNG(t)

	registry, err := checker.NewRegistry([]checker.MissionChecker{
		checker.NewTwoContinentsChecker(),
		checker.NewTwoContinentsPlusOneChecker(),
		checker.NewEighteenTerritoriesChecker(),
		checker.NewTwentyFourTerritoriesChecker(),
		checker.NewEliminatePlayerChecker(),
	})
	require.NoError(t, err)

	service := mission.New(rng, querier, registry)

	return service
}

func input() ctx.GameContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

func testContinents() boardpkg.Continents {
	boardDto := &boardpkg.BoardDto{
		Regions: []boardpkg.RegionDto{
			{ExternalReference: "na_1", Continent: "north_america"},
			{ExternalReference: "na_2", Continent: "north_america"},
			{ExternalReference: "na_3", Continent: "north_america"},
			{ExternalReference: "af_1", Continent: "africa"},
			{ExternalReference: "af_2", Continent: "africa"},
			{ExternalReference: "sa_1", Continent: "south_america"},
			{ExternalReference: "sa_2", Continent: "south_america"},
		},
		Continents: []boardpkg.ContinentDto{
			{ExternalReference: "north_america", BonusTroops: 5},
			{ExternalReference: "africa", BonusTroops: 3},
			{ExternalReference: "south_america", BonusTroops: 2},
		},
	}

	continents, err := boardpkg.NewContinents(boardDto)
	if err != nil {
		panic(fmt.Sprintf("failed to create test continents: %v", err))
	}

	return continents
}

func makeRegions(count int, troops int64) []snapshot.RegionState {
	regions := make([]snapshot.RegionState, count)
	for i := range regions {
		regions[i] = snapshot.RegionState{
			ID:      fmt.Sprintf("region_%d", i+1),
			OwnerID: "giovanni",
			Troops:  troops,
		}
	}

	return regions
}

func regionsForContinent(
	continents boardpkg.Continents,
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

func TestServiceImpl_IsTwoContinentsMissionAccomplished(t *testing.T) {
	t.Parallel()

	continents := testContinents()

	type inputType struct {
		name       string
		regions    []snapshot.RegionState
		continent1 string
		continent2 string
		expected   bool
	}

	tests := []inputType{
		{
			"player does not control any continent",
			[]snapshot.RegionState{},
			"north_america",
			"africa",
			false,
		},
		{
			"controls only first continent",
			regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "opponent", nil)),
			"north_america",
			"africa",
			false,
		},
		{
			"controls both continents",
			regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "giovanni", nil)),
			"north_america",
			"africa",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			gameCtx := input()

			privateSnapshots := map[string]*snapshot.PlayerPrivate{
				"giovanni": {
					Mission: snapshot.PlayerMission{
						Type: snapshot.MissionTwoContinents,
						Detail: snapshot.TwoContinentsMission{
							Continent1: test.continent1,
							Continent2: test.continent2,
						},
					},
				},
			}

			result, err := service.IsMissionAccomplished(
				gameCtx, test.regions, privateSnapshots, continents,
			)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestServiceImpl_IsTwoContinentsPlusOneMissionAccomplished(t *testing.T) {
	t.Parallel()

	continents := testContinents()

	type inputType struct {
		name       string
		regions    []snapshot.RegionState
		continent1 string
		continent2 string
		expected   bool
	}

	tests := []inputType{
		{
			"player does not control any continent",
			[]snapshot.RegionState{},
			"north_america",
			"africa",
			false,
		},
		{
			"controls both mandatory but no third",
			regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "giovanni", nil)),
			"north_america",
			"africa",
			false,
		},
		{
			"controls both mandatory plus a third",
			regionsForContinent(continents, "north_america", "giovanni",
				regionsForContinent(continents, "africa", "giovanni",
					regionsForContinent(continents, "south_america", "giovanni", nil))),
			"north_america",
			"africa",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			gameCtx := input()

			privateSnapshots := map[string]*snapshot.PlayerPrivate{
				"giovanni": {
					Mission: snapshot.PlayerMission{
						Type: snapshot.MissionTwoContinentsPlusOne,
						Detail: snapshot.TwoContinentsPlusOneMission{
							Continent1: test.continent1,
							Continent2: test.continent2,
						},
					},
				},
			}

			result, err := service.IsMissionAccomplished(
				gameCtx, test.regions, privateSnapshots, continents,
			)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestServiceImpl_IsEighteenTerritoriesTwoTroopsMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name     string
		regions  []snapshot.RegionState
		expected bool
	}

	tests := []inputType{
		{
			"17 regions with 2 troops is not enough",
			makeRegions(17, 2),
			false,
		},
		{
			"18 regions with 2 troops is enough",
			makeRegions(18, 2),
			true,
		},
		{
			"18 regions but only 17 with 2 troops is not enough",
			func() []snapshot.RegionState {
				r := makeRegions(18, 2)
				r[0].Troops = 1

				return r
			}(),
			false,
		},
		{
			"19 regions but only 18 with 2 troops is enough",
			func() []snapshot.RegionState {
				r := makeRegions(19, 2)
				r[0].Troops = 1

				return r
			}(),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			gameCtx := input()

			privateSnapshots := map[string]*snapshot.PlayerPrivate{
				"giovanni": {
					Mission: snapshot.PlayerMission{
						Type:   snapshot.MissionEighteenTerritoriesTwoTroops,
						Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
					},
				},
			}

			result, err := service.IsMissionAccomplished(
				gameCtx, test.regions, privateSnapshots, nil,
			)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestServiceImpl_IsTwentyFourTerritoriesMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name     string
		regions  []snapshot.RegionState
		expected bool
	}

	tests := []inputType{
		{
			"23 regions is not enough",
			makeRegions(23, 1),
			false,
		},
		{
			"24 regions is enough",
			makeRegions(24, 1),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			gameCtx := input()

			privateSnapshots := map[string]*snapshot.PlayerPrivate{
				"giovanni": {
					Mission: snapshot.PlayerMission{
						Type:   snapshot.MissionTwentyFourTerritories,
						Detail: snapshot.TwentyFourTerritoriesMission{},
					},
				},
			}

			result, err := service.IsMissionAccomplished(
				gameCtx, test.regions, privateSnapshots, nil,
			)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestServiceImpl_IsEliminatePlayerMissionAccomplished(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name     string
		regions  []snapshot.RegionState
		expected bool
	}

	tests := []inputType{
		{
			"target controls zero regions",
			makeRegions(10, 1),
			true,
		},
		{
			"target controls one region",
			append(
				makeRegions(10, 1),
				snapshot.RegionState{ID: "target_r1", OwnerID: "opponent", Troops: 1},
			),
			false,
		},
		{
			"target controls two regions",
			append(
				makeRegions(10, 1),
				snapshot.RegionState{ID: "target_r1", OwnerID: "opponent", Troops: 1},
				snapshot.RegionState{ID: "target_r2", OwnerID: "opponent", Troops: 2},
			),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			gameCtx := input()

			privateSnapshots := map[string]*snapshot.PlayerPrivate{
				"giovanni": {
					Mission: snapshot.PlayerMission{
						Type: snapshot.MissionEliminatePlayer,
						Detail: snapshot.EliminatePlayerMission{
							TargetUserID: "opponent",
						},
					},
				},
			}

			result, err := service.IsMissionAccomplished(
				gameCtx, test.regions, privateSnapshots, nil,
			)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestServiceImpl_IsMissionAccomplished_MissingPlayer(t *testing.T) {
	t.Parallel()

	service := setup(t)
	gameCtx := input()

	privateSnapshots := map[string]*snapshot.PlayerPrivate{
		"other_player": {
			Mission: snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		},
	}

	_, err := service.IsMissionAccomplished(
		gameCtx, []snapshot.RegionState{}, privateSnapshots, nil,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no private snapshot for player giovanni")
}

func TestServiceImpl_IsMissionAccomplished_UnknownMissionType(t *testing.T) {
	t.Parallel()

	service := setup(t)
	gameCtx := input()

	privateSnapshots := map[string]*snapshot.PlayerPrivate{
		"giovanni": {
			Mission: snapshot.PlayerMission{
				Type:   "nonexistent_type",
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		},
	}

	_, err := service.IsMissionAccomplished(
		gameCtx, []snapshot.RegionState{}, privateSnapshots, nil,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown mission type")
}
