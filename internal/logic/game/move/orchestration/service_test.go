package orchestration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/pool"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/orchestration/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/orchestration/validation"
	gamestate "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/state"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	db.Querier,
	*gamestate.Service,
	*phase.Service,
	*validation.Service,
	*orchestration.ServiceImpl,
) {
	t.Helper()
	mockDB := pool.NewDB(t)
	querier := db.New(mockDB)
	gameService := gamestate.NewService(t)
	phaseService := phase.NewService(t)
	validationService := validation.NewService(t)
	service := orchestration.NewService(
		querier,
		phaseService,
		gameService,
		validationService,
		nil,
		nil,
		nil,
	)

	return querier, gameService, phaseService, validationService, service
}

func input() ctx.MoveContext {
	gameID := int64(1)
	userID := "23011836-df7e-4421-bdbf-b9c07b22eb64"

	userContext := ctx.WithUserID(ctx.WithLog(context.Background(), zap.NewNop().Sugar()), userID)

	gameContext := ctx.WithGameID(userContext, gameID)

	return ctx.NewMoveContext(userContext, gameContext)
}

func TestServiceImpl_PerformMove(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name            string
		phase           sqlc.PhaseType
		validationError error
		performError    error
		advanceError    error
		error           string
	}

	tests := []inputType{
		{
			"Should fail when game is not in the correct phase",
			sqlc.PhaseTypeATTACK,
			nil,
			nil,
			nil,
			"game is not in the correct phase to perform move",
		},
		{
			"Should fail when validation state fails",
			sqlc.PhaseTypeDEPLOY,
			fmt.Errorf("validation error"),
			nil,
			nil,
			"invalid move: validation error",
		},
		{
			"Should fail when perform function fails",
			sqlc.PhaseTypeDEPLOY,
			nil,
			fmt.Errorf("perform error"),
			nil,
			"unable to perform move: perform error",
		},
		{
			"Should fail when unable to advance phase",
			sqlc.PhaseTypeDEPLOY,
			nil,
			nil,
			fmt.Errorf("advance error"),
			"unable to advance phase: advance error",
		},
		{
			"Should succeed and advance phase",
			sqlc.PhaseTypeDEPLOY,
			nil,
			nil,
			nil,
			"",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, gameService, phaseService, validationService, service := setup(t)
			context := input()

			game := &state.Game{
				ID:           context.GameID(),
				CurrentPhase: sqlc.PhaseTypeDEPLOY,
				CurrentTurn:  2,
			}

			gameService.
				EXPECT().
				GetGameStateQ(context, querier).
				Return(game, nil)

			if test.phase == sqlc.PhaseTypeDEPLOY {
				validationService.
					EXPECT().
					Validate(context, querier, game).
					Return(test.validationError)
				if test.validationError == nil && test.performError == nil {
					phaseService.
						EXPECT().
						AdvanceQ(context, querier).
						Return(test.advanceError)
				}
			}

			performFunc := func(c ctx.MoveContext, querier db.Querier) error {
				return test.performError
			}
			err := service.OrchestrateMoveQ(context, querier, test.phase, performFunc)

			if test.phase == sqlc.PhaseTypeDEPLOY &&
				test.validationError == nil &&
				test.performError == nil &&
				test.advanceError == nil {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			require.EqualError(t, err, test.error)
		})
	}
}
