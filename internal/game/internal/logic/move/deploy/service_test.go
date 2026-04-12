package deploy_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/deploy"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) deploy.Service {
	t.Helper()

	service, _ := deploy.NewService(nil)

	return service
}

func input() (string, int64, int64, ctx.GameContext) {
	gameID := int64(1)
	userID := "Giovanni"
	regionReference := "greenland"
	currentTroops := 0
	desiredTroops := 5
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	gameContext := ctx.WithGameID(userContext, gameID)

	return regionReference, int64(
			currentTroops,
		), int64(
			desiredTroops,
		), gameContext
}

func cachedRegions(owner string, troops int64) []snapshot.RegionState {
	return []snapshot.RegionState{
		{
			InternalID: 1,
			ID:         "greenland",
			OwnerID:    owner,
			Troops:     troops,
		},
	}
}

func prevWithDeployableTroops(troops int64) *snapshot.CachedGameState {
	return prevWithDeployableTroopsAndRegions(troops, "Giovanni", 0)
}

func prevWithDeployableTroopsAndRegions(
	deployableTroops int64,
	regionOwner string,
	regionTroops int64,
) *snapshot.CachedGameState {
	return &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{
					DeployableTroops: deployableTroops,
				},
			},
			Regions: cachedRegions(regionOwner, regionTroops),
		},
	}
}

func TestService_DeployShouldFailWhenPlayerDoesntHaveEnoughDeployableTroops(t *testing.T) {
	t.Parallel()

	service := setup(t)
	regionReference, currentTroops, desiredTroops, ctx := input()

	prev := prevWithDeployableTroops(0)

	_, _, err := service.Perform(ctx, deploy.Move{
		RegionID:      regionReference,
		CurrentTroops: currentTroops,
		DesiredTroops: desiredTroops,
	}, prev)

	require.Error(t, err)
	require.EqualError(t, err, "not enough deployable troops")
}

func TestService_DeployShouldFail(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name           string
		regionOwner    string
		declaredTroops int64
		expectedError  string
	}

	tests := []inputType{
		{
			"When region is not owned by player",
			"Gabriele",
			0,
			"deploy region is not owned by player",
		},
		{
			"When amount of troops declared is wrong",
			"Giovanni",
			10,
			"must deploy at least 1 troop",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			regionReference, _, desiredTroops, ctx := input()

			currentTroops := test.declaredTroops

			prev := prevWithDeployableTroopsAndRegions(5, test.regionOwner, 0)

			_, _, err := service.Perform(ctx, deploy.Move{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			}, prev)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestService_DeployShouldFailWhenRegionNotInCache(t *testing.T) {
	t.Parallel()

	service := setup(t)
	_, _, _, ctx := input()

	prev := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{
					DeployableTroops: 10,
				},
			},
			Regions: []snapshot.RegionState{
				{InternalID: 99, ID: "brazil", OwnerID: "Giovanni", Troops: 3},
			},
		},
	}

	_, _, err := service.Perform(ctx, deploy.Move{
		RegionID:      "greenland",
		CurrentTroops: 0,
		DesiredTroops: 5,
	}, prev)

	require.Error(t, err)
	require.EqualError(
		t,
		err,
		"failed to find region in cache: region greenland not found in cached state",
	)
}

func TestService_DeployShouldSucceed(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name             string
		deployableTroops int64
	}

	tests := []inputType{
		{
			"Should succeed without advancing phase",
			15,
		},
		{
			"Should succeed and advance phase",
			5,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := setup(t)
			regionReference, currentTroops, desiredTroops, ctx := input()
			troops := desiredTroops - currentTroops

			prev := prevWithDeployableTroopsAndRegions(
				test.deployableTroops,
				"Giovanni",
				currentTroops,
			)

			_, effect, err := service.Perform(ctx, deploy.Move{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			}, prev)

			require.NoError(t, err)

			// Verify MoveEffect
			require.Len(t, effect.RegionUpdates, 1)
			require.Equal(t, moveservice.RegionUpdate{
				RegionID:  "greenland",
				NewOwner:  "Giovanni",
				NewTroops: currentTroops + troops,
			}, effect.RegionUpdates[0])
			require.Equal(t, snapshot.DeployPhaseState{
				DeployableTroops: test.deployableTroops - troops,
			}, effect.UpdatedPhase)
		})
	}
}
