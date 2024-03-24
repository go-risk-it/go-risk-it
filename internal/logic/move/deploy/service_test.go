package deploy_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/move/deploy"
	"github.com/tomfran/go-risk-it/mocks/internal_/data/db"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/game"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/player"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/region"
	"github.com/tomfran/go-risk-it/mocks/internal_/signals"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*player.Service,
	*game.Service,
	*region.Service,
	*deploy.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	gameService := game.NewService(t)
	regionService := region.NewService(t)
	boardStateChangedSignal := signals.NewBoardStateChangedSignal(t)
	playerStateChangedSignal := signals.NewPlayerStateChangedSignal(t)
	gameStateChangedSignal := signals.NewGameStateChangedSignal(t)
	service := deploy.NewService(
		querier,
		zap.NewNop().Sugar(),
		gameService,
		playerService,
		regionService,
		boardStateChangedSignal,
		playerStateChangedSignal,
		gameStateChangedSignal,
	)

	return querier, playerService, gameService, regionService, service
}

func input() (int64, string, string, int, context.Context) {
	gameID := int64(1)
	userID := "Giovanni"
	regionReference := "greenland"
	troops := 5
	ctx := context.Background()

	return gameID, userID, regionReference, troops, ctx
}

func TestServiceImpl_DeployShouldFailWhenPlayerNotInGame(t *testing.T) {
	t.Parallel()

	querier, playerService, _, _, service := setup(t)
	gameID, userID, regionReference, troops, ctx := input()

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
	}
	playerService.
		EXPECT().
		GetPlayersQ(ctx, querier, gameID).
		Return(players, nil)

	err := service.PerformDeployMoveQ(
		ctx,
		querier,
		gameID,
		userID,
		regionReference,
		troops,
	)

	require.Error(t, err)
	require.EqualError(t, err, "player is not in game")
}

func TestServiceImpl_DeployShouldFailOnTurnCheck(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name        string
		phase       sqlc.Phase
		turn        int64
		expectedErr string
	}

	tests := []inputType{
		{
			"When not player's turn",
			sqlc.PhaseDEPLOY,
			1,
			"turn check failed: it is not the player's turn",
		},
		{
			"When incorrect phase",
			sqlc.PhaseATTACK,
			2,
			"turn check failed: game is not in deploy phase",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			querier, playerService, gameService, _, service := setup(t)
			gameID, userID, regionReference, troops, ctx := input()

			players := []sqlc.Player{
				{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
				{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
				{ID: 42069, TurnIndex: 2, GameID: 1, UserID: "Giovanni"},
			}
			playerService.
				EXPECT().
				GetPlayersQ(ctx, querier, gameID).
				Return(players, nil)
			gameService.
				EXPECT().
				GetGameStateQ(ctx, querier, gameID).
				Return(&sqlc.Game{
					ID:    gameID,
					Phase: test.phase,
					Turn:  test.turn,
				}, nil)

			err := service.PerformDeployMoveQ(
				ctx,
				querier,
				gameID,
				userID,
				regionReference,
				troops,
			)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestServiceImpl_DeployShouldFailWhenPlayerDoesntHaveEnoughDeployableTroops(t *testing.T) {
	t.Parallel()

	querier, playerService, gameService, _, service := setup(t)
	gameID, userID, regionReference, troops, ctx := input()

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele", DeployableTroops: 10},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco", DeployableTroops: 10},
		{ID: 42069, TurnIndex: 2, GameID: 1, UserID: "Giovanni", DeployableTroops: 4},
	}
	playerService.
		EXPECT().
		GetPlayersQ(ctx, querier, gameID).
		Return(players, nil)
	gameService.
		EXPECT().
		GetGameStateQ(ctx, querier, gameID).
		Return(&sqlc.Game{
			ID:    gameID,
			Phase: sqlc.PhaseDEPLOY,
			Turn:  2,
		}, nil)

	err := service.PerformDeployMoveQ(
		ctx,
		querier,
		gameID,
		userID,
		regionReference,
		troops,
	)

	require.Error(t, err)
	require.EqualError(t, err, "not enough deployable troops")
}

func TestServiceImpl_DeployShouldFailWhenRegionNotOwnedByPlayer(t *testing.T) {
	t.Parallel()

	querier, playerService, gameService, regionService, service := setup(t)
	gameID, userID, regionReference, troops, ctx := input()

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele", DeployableTroops: 10},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco", DeployableTroops: 10},
		{ID: 42069, TurnIndex: 2, GameID: 1, UserID: "Giovanni", DeployableTroops: 5},
	}
	playerService.
		EXPECT().
		GetPlayersQ(ctx, querier, gameID).
		Return(players, nil)
	gameService.
		EXPECT().
		GetGameStateQ(ctx, querier, gameID).
		Return(&sqlc.Game{
			ID:    gameID,
			Phase: sqlc.PhaseDEPLOY,
			Turn:  2,
		}, nil)
	regionService.
		EXPECT().
		GetRegionQ(ctx, querier, gameID, regionReference).
		Return(&sqlc.GetRegionsByGameRow{
			ID:                1,
			ExternalReference: "greenland",
			PlayerName:        "Gabriele",
			Troops:            10,
		}, nil)

	err := service.PerformDeployMoveQ(
		ctx,
		querier,
		gameID,
		userID,
		regionReference,
		troops,
	)

	require.Error(t, err)
	require.EqualError(t, err, "failed to get region: region is not owned by player")
}

func TestServiceImpl_DeployShouldSucceed(t *testing.T) {
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
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, playerService, gameService, regionService, service := setup(
				t,
			)
			gameID, userID, regionReference, troops, ctx := input()

			gabriele := sqlc.Player{
				ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele", DeployableTroops: 15,
			}
			francesco := sqlc.Player{
				ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco", DeployableTroops: 15,
			}
			giovanni := sqlc.Player{
				ID:               42069,
				TurnIndex:        2,
				GameID:           1,
				UserID:           "Giovanni",
				DeployableTroops: test.deployableTroops,
			}
			players := []sqlc.Player{
				gabriele,
				francesco,
				giovanni,
			}

			playerService.
				EXPECT().
				GetPlayersQ(ctx, querier, gameID).
				Return(players, nil)
			gameService.
				EXPECT().
				GetGameStateQ(ctx, querier, gameID).
				Return(&sqlc.Game{
					ID:    gameID,
					Phase: sqlc.PhaseDEPLOY,
					Turn:  2,
				}, nil)
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, gameID, regionReference).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: "greenland",
					PlayerName:        "Giovanni",
					Troops:            10,
				}, nil)
			playerService.
				EXPECT().
				DecreaseDeployableTroopsQ(ctx, querier, &giovanni, int64(troops)).
				Return(nil)
			regionService.
				EXPECT().
				IncreaseTroopsInRegion(ctx, querier, int64(1), int64(troops)).
				Return(nil)
			if test.deployableTroops == int64(troops) {
				gameService.
					EXPECT().
					SetGamePhaseQ(ctx, querier, gameID, sqlc.PhaseATTACK).
					Return(nil)
			}

			err := service.PerformDeployMoveQ(
				ctx,
				querier,
				gameID,
				userID,
				regionReference,
				troops,
			)

			require.NoError(t, err)
			gameService.AssertExpectations(t)
			playerService.AssertExpectations(t)
			regionService.AssertExpectations(t)
		})
	}
}
