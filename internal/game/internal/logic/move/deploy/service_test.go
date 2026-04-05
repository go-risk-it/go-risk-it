package deploy_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/deploy"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/region"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (
	*db.Querier,
	*region.Service,
	deploy.Service,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	phaseService := phase.NewService(t)
	regionService := region.NewService(t)
	service, _ := deploy.NewService(querier, phaseService, regionService)

	return querier, regionService, service
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

func prevWithDeployableTroops(troops int64) *snapshot.CachedGameState {
	return &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{
					DeployableTroops: troops,
				},
			},
		},
	}
}

func TestService_DeployShouldFailWhenPlayerDoesntHaveEnoughDeployableTroops(t *testing.T) {
	t.Parallel()

	_, _, service := setup(t)
	regionReference, currentTroops, desiredTroops, ctx := input()

	prev := prevWithDeployableTroops(0)

	_, _, err := service.Perform(ctx, nil, deploy.Move{
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

			querier, regionService, service := setup(t)
			regionReference, _, desiredTroops, ctx := input()

			currentTroops := test.declaredTroops

			prev := prevWithDeployableTroops(5)

			regionService.
				EXPECT().
				GetRegion(ctx, querier, regionReference).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            test.regionOwner,
					Troops:            0,
				}, nil)

			_, _, err := service.Perform(ctx, querier, deploy.Move{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			}, prev)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
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

			querier, regionService, service := setup(
				t,
			)
			regionReference, currentTroops, desiredTroops, ctx := input()
			troops := desiredTroops - currentTroops

			prev := prevWithDeployableTroops(test.deployableTroops)

			region := &sqlc.GetRegionsByGameRow{
				ID:                1,
				ExternalReference: "greenland",
				UserID:            "Giovanni",
				Troops:            currentTroops,
			}

			regionService.
				EXPECT().
				GetRegion(ctx, querier, regionReference).
				Return(region, nil)
			regionService.
				EXPECT().
				UpdateTroopsInRegion(ctx, querier, region, troops).
				Return(nil)
			querier.
				EXPECT().
				DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
					ID:               ctx.GameID(),
					DeployableTroops: desiredTroops - currentTroops,
				}).
				Return(nil)

			_, effect, err := service.Perform(ctx, querier, deploy.Move{
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
