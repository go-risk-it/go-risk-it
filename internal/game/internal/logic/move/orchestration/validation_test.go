package orchestration_test

import (
	"context"
	"testing"

	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

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

	service := orchestration.NewValidationService()
	gameCtx := validationInput()

	players := []apisnapshot.PlayerState{
		{UserID: "Gabriele", Name: "Gabriele", Index: 0, Status: apisnapshot.PlayerAlive},
		{UserID: "Francesco", Name: "Francesco", Index: 1, Status: apisnapshot.PlayerAlive},
	}

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeDEPLOY,
		Turn:  1,
	}

	err := service.Validate(gameCtx, game, players)

	require.Error(t, err)
	require.EqualError(t, err, "player is not in game")
}

func TestValidationService_ShouldFailOnTurnCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phase       sqlc.GamePhaseType
		turn        int64
		expectedErr string
	}{
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

			service := orchestration.NewValidationService()
			gameCtx := validationInput()

			players := []apisnapshot.PlayerState{
				{UserID: "Gabriele", Name: "Gabriele", Index: 0, Status: apisnapshot.PlayerAlive},
				{
					UserID: "Francesco",
					Name:   "Francesco",
					Index:  1,
					Status: apisnapshot.PlayerAlive,
				},
				{UserID: "Giovanni", Name: "Giovanni", Index: 2, Status: apisnapshot.PlayerAlive},
			}

			game := &state.Game{
				ID:    gameCtx.GameID(),
				Phase: test.phase,
				Turn:  test.turn,
			}

			err := service.Validate(gameCtx, game, players)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestValidationService_ShouldSucceed(t *testing.T) {
	t.Parallel()

	service := orchestration.NewValidationService()
	gameCtx := validationInput()

	players := []apisnapshot.PlayerState{
		{UserID: "Gabriele", Name: "Gabriele", Index: 0, Status: apisnapshot.PlayerAlive},
		{UserID: "Giovanni", Name: "Giovanni", Index: 1, Status: apisnapshot.PlayerAlive},
	}

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeDEPLOY,
		Turn:  1,
	}

	err := service.Validate(gameCtx, game, players)
	require.NoError(t, err)
}

func TestValidationService_ShouldFailWhenGameOver(t *testing.T) {
	t.Parallel()

	service := orchestration.NewValidationService()
	gameCtx := validationInput()

	players := []apisnapshot.PlayerState{
		{UserID: "Giovanni", Name: "Giovanni", Index: 0, Status: apisnapshot.PlayerAlive},
	}

	game := &state.Game{
		ID:           gameCtx.GameID(),
		Phase:        sqlc.GamePhaseTypeDEPLOY,
		Turn:         0,
		WinnerUserID: "someone",
	}

	err := service.Validate(gameCtx, game, players)
	require.Error(t, err)
	require.EqualError(t, err, "game is already over")
}
