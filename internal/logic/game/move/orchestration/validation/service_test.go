package validation_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*player.Service,
	*validation.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	service := validation.NewService(zap.NewNop().Sugar(), playerService)

	return querier, playerService, service
}

func input() ctx.MoveContext {
	gameID := int64(1)
	userID := "Giovanni"
	userContext := ctx.WithUserID(ctx.WithLog(context.Background(), zap.NewNop().Sugar()), userID)

	gameContext := ctx.WithGameID(userContext, gameID)

	return ctx.NewMoveContext(userContext, gameContext)
}

func TestServiceImpl_ShouldFailWhenPlayerNotInGame(t *testing.T) {
	t.Parallel()

	querier, playerService, service := setup(t)
	ctx := input()

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
	}

	game := &sqlc.Game{
		ID:               ctx.GameID(),
		Phase:            sqlc.PhaseDEPLOY,
		Turn:             1,
		DeployableTroops: 5,
	}

	playerService.
		EXPECT().
		GetPlayersQ(ctx, querier, ctx.GameID()).
		Return(players, nil)

	err := service.Validate(ctx, querier, game)

	require.Error(t, err)
	require.EqualError(t, err, "player is not in game")
}

func TestServiceImpl_ShouldFailOnTurnCheck(t *testing.T) {
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
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			querier, playerService, service := setup(t)
			ctx := input()

			players := []sqlc.Player{
				{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
				{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
				{ID: 42069, TurnIndex: 2, GameID: 1, UserID: "Giovanni"},
			}
			playerService.
				EXPECT().
				GetPlayersQ(ctx, querier, ctx.GameID()).
				Return(players, nil)
			game := &sqlc.Game{
				ID:    ctx.GameID(),
				Phase: test.phase,
				Turn:  test.turn,
			}

			err := service.Validate(ctx, querier, game)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}
