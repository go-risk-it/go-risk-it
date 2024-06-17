package deploy_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/region"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/signals"
	"github.com/stretchr/testify/require"
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

func input() (int64, string, string, int64, int64, context.Context) {
	gameID := int64(1)
	userID := "Giovanni"
	regionReference := "greenland"
	currentTroops := 0
	desiredTroops := 5
	ctx := context.Background()

	return gameID, userID, regionReference, int64(currentTroops), int64(desiredTroops), ctx
}

func TestServiceImpl_DeployShouldFailWhenPlayerNotInGame(t *testing.T) {
	t.Parallel()

	querier, playerService, gameService, _, service := setup(t)
	gameID, userID, regionReference, currentTroops, desiredTroops, ctx := input()

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
	}

	gameService.
		EXPECT().
		GetGameStateQ(ctx, querier, gameID).
		Return(&sqlc.Game{
			ID:               gameID,
			Phase:            sqlc.PhaseDEPLOY,
			Turn:             1,
			DeployableTroops: 5,
		}, nil)
	playerService.
		EXPECT().
		GetPlayersQ(ctx, querier, gameID).
		Return(players, nil)

	err := service.PerformDeployMoveQ(
		ctx,
		querier,
		move.Move[deploy.MoveData]{
			UserID: userID,
			GameID: gameID,
			Payload: deploy.MoveData{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			},
		},
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
			"turn check failed: game is not in DEPLOY phase",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			querier, playerService, gameService, _, service := setup(t)
			gameID, userID, regionReference, currentTroops, desiredTroops, ctx := input()

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
				move.Move[deploy.MoveData]{
					UserID: userID,
					GameID: gameID,
					Payload: deploy.MoveData{
						RegionID:      regionReference,
						CurrentTroops: currentTroops,
						DesiredTroops: desiredTroops,
					},
				},
			)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestServiceImpl_DeployShouldFailWhenPlayerDoesntHaveEnoughDeployableTroops(t *testing.T) {
	t.Parallel()

	querier, playerService, gameService, _, service := setup(t)
	gameID, userID, regionReference, currentTroops, desiredTroops, ctx := input()

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
			Phase: sqlc.PhaseDEPLOY,
			Turn:  2,
		}, nil)

	err := service.PerformDeployMoveQ(
		ctx,
		querier,
		move.Move[deploy.MoveData]{
			UserID: userID,
			GameID: gameID,
			Payload: deploy.MoveData{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			},
		},
	)

	require.Error(t, err)
	require.EqualError(t, err, "not enough deployable troops")
}

func TestServiceImpl_DeployShouldFail(t *testing.T) {
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
			"failed to get region: region is not owned by player",
		},
		{
			"When amount of troops declared is wrong",
			"Giovanni",
			10,
			"region has different number of troops than declared",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, playerService, gameService, regionService, service := setup(t)
			gameID, userID, regionReference, _, desiredTroops, ctx := input()

			currentTroops := test.declaredTroops

			players := []sqlc.Player{
				{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
				{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
				{ID: 42069, TurnIndex: 2, GameID: 1, UserID: "Giovanni"},
			}
			gameService.EXPECT().
				GetGameStateQ(ctx, querier, gameID).
				Return(&sqlc.Game{
					ID:               gameID,
					Phase:            sqlc.PhaseDEPLOY,
					Turn:             2,
					DeployableTroops: 5,
				}, nil)
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
					UserID:            test.regionOwner,
					Troops:            0,
				}, nil)

			err := service.PerformDeployMoveQ(
				ctx,
				querier,
				move.Move[deploy.MoveData]{
					UserID: userID,
					GameID: gameID,
					Payload: deploy.MoveData{
						RegionID:      regionReference,
						CurrentTroops: currentTroops,
						DesiredTroops: desiredTroops,
					},
				},
			)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
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
			gameID, userID, regionReference, currentTroops, desiredTroops, ctx := input()
			troops := desiredTroops - currentTroops

			gabriele := sqlc.Player{
				ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele",
			}
			francesco := sqlc.Player{
				ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco",
			}
			giovanni := sqlc.Player{
				ID:        42069,
				TurnIndex: 2,
				GameID:    1,
				UserID:    "Giovanni",
			}
			players := []sqlc.Player{
				gabriele,
				francesco,
				giovanni,
			}

			game := &sqlc.Game{
				ID:               gameID,
				Phase:            sqlc.PhaseDEPLOY,
				Turn:             2,
				DeployableTroops: test.deployableTroops,
			}
			gameService.EXPECT().
				GetGameStateQ(ctx, querier, gameID).
				Return(game, nil)
			playerService.
				EXPECT().
				GetPlayersQ(ctx, querier, gameID).
				Return(players, nil)
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, gameID, regionReference).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "Giovanni",
					Troops:            0,
				}, nil)
			gameService.
				EXPECT().
				DecreaseDeployableTroopsQ(ctx, querier, game, troops).
				Return(nil)
			regionService.
				EXPECT().
				IncreaseTroopsInRegion(ctx, querier, int64(1), troops).
				Return(nil)
			if test.deployableTroops == troops {
				gameService.
					EXPECT().
					SetGamePhaseQ(ctx, querier, gameID, sqlc.PhaseATTACK).
					Return(nil)
			}

			err := service.PerformDeployMoveQ(
				ctx,
				querier,
				move.Move[deploy.MoveData]{
					UserID: userID,
					GameID: gameID,
					Payload: deploy.MoveData{
						RegionID:      regionReference,
						CurrentTroops: currentTroops,
						DesiredTroops: desiredTroops,
					},
				},
			)

			require.NoError(t, err)
			gameService.AssertExpectations(t)
			playerService.AssertExpectations(t)
			regionService.AssertExpectations(t)
		})
	}
}
