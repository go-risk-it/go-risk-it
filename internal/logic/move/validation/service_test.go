package validation_test

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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*player.Service,
	*game.Service,
	*deploy.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	gameService := game.NewService(t)
	regionService := region.NewService(t)
	service := deploy.NewService(
		querier,
		zap.NewNop().Sugar(),
		gameService,
		playerService,
		regionService,
	)

	return querier, playerService, gameService, service
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

	querier, playerService, gameService, service := setup(t)
	gameID, userID, regionReference, currentTroops, desiredTroops, ctx := input()

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
	}

	game := &sqlc.Game{
		ID:               gameID,
		Phase:            sqlc.PhaseDEPLOY,
		Turn:             1,
		DeployableTroops: 5,
	}
	gameService.
		EXPECT().
		GetGameStateQ(ctx, querier, gameID).
		Return(game, nil)
	playerService.
		EXPECT().
		GetPlayersQ(ctx, querier, gameID).
		Return(players, nil)

	err := service.PerformQ(
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
		game,
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
			querier, playerService, gameService, service := setup(t)
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
			game := &sqlc.Game{
				ID:    gameID,
				Phase: test.phase,
				Turn:  test.turn,
			}
			gameService.
				EXPECT().
				GetGameStateQ(ctx, querier, gameID).
				Return(game, nil)

			err := service.PerformQ(
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
				game,
			)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}
