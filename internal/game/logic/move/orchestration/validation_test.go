package orchestration_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/game/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/player"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func validationSetup(t *testing.T) (
	*db.Querier,
	*player.Service,
	orchestration.ValidationService,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	service := orchestration.NewValidationService(playerService)

	return querier, playerService, service
}

func validationInput() ctx.GameContext {
	gameID := int64(1)
	userID := "Giovanni"
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

func TestValidationService_ShouldFailWhenPlayerNotInGame(t *testing.T) {
	t.Parallel()

	querier, playerService, service := validationSetup(t)
	ctx := validationInput()

	players := []sqlc.GamePlayer{
		{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
		{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
	}

	game := &state.Game{
		ID:    ctx.GameID(),
		Phase: sqlc.GamePhaseTypeDEPLOY,
		Turn:  1,
	}

	playerService.
		EXPECT().
		GetPlayers(mock.Anything, querier).
		Return(players, nil)

	err := service.Validate(ctx, querier, game)

	require.Error(t, err)
	require.EqualError(t, err, "player is not in game")
}

func TestValidationService_ShouldFailOnTurnCheck(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name        string
		phase       sqlc.GamePhaseType
		turn        int64
		expectedErr string
	}

	tests := []inputType{
		{
			"When not player's turn",
			sqlc.GamePhaseTypeDEPLOY,
			1,
			"turn check failed: it is not the player's turn",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			querier, playerService, service := validationSetup(t)
			ctx := validationInput()

			players := []sqlc.GamePlayer{
				{ID: 420, TurnIndex: 0, GameID: 1, UserID: "Gabriele"},
				{ID: 69, TurnIndex: 1, GameID: 1, UserID: "Francesco"},
				{ID: 42069, TurnIndex: 2, GameID: 1, UserID: "Giovanni"},
			}
			playerService.
				EXPECT().
				GetPlayers(mock.Anything, querier).
				Return(players, nil)

			game := &state.Game{
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
