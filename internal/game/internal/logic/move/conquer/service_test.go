package conquer_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/conquer"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/attack"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (*db.Querier, conquer.Service) {
	t.Helper()

	querier := db.NewQuerier(t)
	attackService := attack.NewService(t)

	service, _ := conquer.NewService(querier, attackService)

	return querier, service
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

func prevWithConquerState(
	source, target string,
	minTroops int64,
	regions []snapshot.RegionState,
) *snapshot.CachedGameState {
	return &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: source,
					DefendingRegionID: target,
					MinTroopsToMove:   minTroops,
				},
			},
			Regions: regions,
		},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{},
	}
}

func TestServiceImpl_Perform_ShouldFailValidation(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name          string
		minimumTroops int64
		moveTroops    int64
		sourceTroops  int64
		expectedError string
	}

	tests := []inputType{
		{
			"When moving fewer troops than minimum",
			3,
			2,
			5,
			"must move at least 3 troops",
		},
		{
			"When source region does not have enough troops",
			2,
			5,
			5,
			"source region does not have enough troops",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, service := setup(t)
			ctx := input()

			regions := []snapshot.RegionState{
				{
					InternalID: 1,
					ID:         "greenland",
					OwnerID:    "giovanni",
					Troops:     test.sourceTroops,
				},
				{
					InternalID: 2,
					ID:         "iceland",
					OwnerID:    "gabriele",
					Troops:     0,
				},
			}
			prev := prevWithConquerState("greenland", "iceland", test.minimumTroops, regions)

			_, _, err := service.Perform(ctx, conquer.Move{
				Troops: test.moveTroops,
			}, prev)

			require.Error(t, err)
			require.ErrorContains(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_Perform_ShouldFailWhenSourceRegionNotInCache(t *testing.T) {
	t.Parallel()

	_, service := setup(t)
	ctx := input()

	// The conquer phase state references "greenland" as source, but it is not in the regions list.
	regions := []snapshot.RegionState{
		{InternalID: 2, ID: "iceland", OwnerID: "gabriele", Troops: 0},
	}
	prev := prevWithConquerState("greenland", "iceland", 2, regions)

	_, _, err := service.Perform(ctx, conquer.Move{
		Troops: 3,
	}, prev)

	require.Error(t, err)
	require.ErrorContains(t, err, "unable to get attacking region")
	require.ErrorContains(t, err, "region greenland not found in cached state")
}

func TestServiceImpl_Perform_ShouldTransferTroopsAndOwnership(t *testing.T) {
	t.Parallel()

	_, service := setup(t)
	ctx := input()

	regions := []snapshot.RegionState{
		{InternalID: 1, ID: "greenland", OwnerID: "giovanni", Troops: 5},
		{InternalID: 2, ID: "iceland", OwnerID: "gabriele", Troops: 0},
		// gabriele owns another region — not eliminated
		{InternalID: 3, ID: "scandinavia", OwnerID: "gabriele", Troops: 2},
	}
	prev := prevWithConquerState("greenland", "iceland", 2, regions)

	_, effect, err := service.Perform(ctx, conquer.Move{
		Troops: 3,
	}, prev)

	require.NoError(t, err)

	// Verify MoveEffect: region updates but no elimination
	require.Len(t, effect.RegionUpdates, 2)
	require.Equal(t, moveservice.RegionUpdate{
		RegionID:  "greenland",
		NewOwner:  "giovanni",
		NewTroops: 2,
	}, effect.RegionUpdates[0])
	require.Equal(t, moveservice.RegionUpdate{
		RegionID:  "iceland",
		NewOwner:  "giovanni",
		NewTroops: 3,
	}, effect.RegionUpdates[1])
	require.Empty(t, effect.CardDeltas)
	require.Empty(t, effect.Missions)
	require.Equal(t, snapshot.EmptyPhaseState{}, effect.UpdatedPhase)
}

func TestServiceImpl_Perform_ShouldHandlePlayerElimination(t *testing.T) {
	t.Parallel()

	_, service := setup(t)
	ctx := input()

	// gabriele owns only iceland — will be eliminated
	regions := []snapshot.RegionState{
		{InternalID: 1, ID: "greenland", OwnerID: "giovanni", Troops: 5},
		{InternalID: 2, ID: "iceland", OwnerID: "gabriele", Troops: 0},
		{InternalID: 3, ID: "brazil", OwnerID: "giovanni", Troops: 3},
	}

	prev := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: "greenland",
					DefendingRegionID: "iceland",
					MinTroopsToMove:   2,
				},
			},
			Regions: regions,
		},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{
			"gabriele": {
				Cards: []snapshot.CardState{
					{ID: 100, Type: snapshot.CardInfantry, Region: "alaska"},
					{ID: 101, Type: snapshot.CardCavalry, Region: "brazil"},
				},
				Mission: snapshot.PlayerMission{
					Type:   snapshot.MissionTwentyFourTerritories,
					Detail: snapshot.TwentyFourTerritoriesMission{},
				},
			},
			"giovanni": {
				Cards: []snapshot.CardState{},
				Mission: snapshot.PlayerMission{
					Type: snapshot.MissionTwoContinents,
					Detail: snapshot.TwoContinentsMission{
						Continent1: "europe",
						Continent2: "asia",
					},
				},
			},
			"paolo": {
				Cards: []snapshot.CardState{},
				Mission: snapshot.PlayerMission{
					Type:   snapshot.MissionEliminatePlayer,
					Detail: snapshot.EliminatePlayerMission{TargetUserID: "gabriele"},
				},
			},
		},
	}

	_, effect, err := service.Perform(ctx, conquer.Move{
		Troops: 3,
	}, prev)

	require.NoError(t, err)

	// Verify region updates
	require.Len(t, effect.RegionUpdates, 2)
	require.Equal(t, snapshot.EmptyPhaseState{}, effect.UpdatedPhase)

	// Verify card deltas: eliminated player loses their cards, conqueror gains them
	require.Len(t, effect.CardDeltas, 2)
	require.Equal(t, "gabriele", effect.CardDeltas[0].PlayerUserID)
	require.ElementsMatch(t, []int64{100, 101}, effect.CardDeltas[0].Lost)
	require.Equal(t, "giovanni", effect.CardDeltas[1].PlayerUserID)
	require.Len(t, effect.CardDeltas[1].Gained, 2)

	// Verify mission changes: paolo's eliminate-player mission becomes 24 territories
	require.Len(t, effect.Missions, 1)
	require.Equal(t, "paolo", effect.Missions[0].PlayerUserID)
	require.Equal(t, snapshot.MissionTwentyFourTerritories, effect.Missions[0].NewMission.Type)
	require.Equal(t, snapshot.TwentyFourTerritoriesMission{}, effect.Missions[0].NewMission.Detail)
}

func TestServiceImpl_Perform_ShouldHandleEliminationWithNoCards(t *testing.T) {
	t.Parallel()

	_, service := setup(t)
	ctx := input()

	// gabriele owns only iceland — will be eliminated (no cards)
	regions := []snapshot.RegionState{
		{InternalID: 1, ID: "greenland", OwnerID: "giovanni", Troops: 5},
		{InternalID: 2, ID: "iceland", OwnerID: "gabriele", Troops: 0},
	}

	prev := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: "greenland",
					DefendingRegionID: "iceland",
					MinTroopsToMove:   2,
				},
			},
			Regions: regions,
		},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{
			"gabriele": {
				Cards: []snapshot.CardState{},
				Mission: snapshot.PlayerMission{
					Type:   snapshot.MissionTwentyFourTerritories,
					Detail: snapshot.TwentyFourTerritoriesMission{},
				},
			},
		},
	}

	_, effect, err := service.Perform(ctx, conquer.Move{
		Troops: 3,
	}, prev)

	require.NoError(t, err)
	require.Empty(t, effect.CardDeltas)
	require.Empty(t, effect.Missions)
}

func TestServiceImpl_GetPhaseStateWithQuerier(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	ctx := input()

	expected := sqlc.GetConquerPhaseStateRow{
		SourceRegion:  "greenland",
		TargetRegion:  "iceland",
		MinimumTroops: 3,
	}

	querier.
		EXPECT().
		GetConquerPhaseState(ctx, ctx.GameID()).
		Return(expected, nil)

	result, err := service.GetPhaseStateWithQuerier(ctx, querier)

	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestServiceImpl_PhaseType(t *testing.T) {
	t.Parallel()

	_, service := setup(t)

	require.Equal(t, sqlc.GamePhaseTypeCONQUER, service.PhaseType())
}
