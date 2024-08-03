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
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/phase"
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
	service := orchestration.NewService(querier, gameService, validationService, nil, nil, nil)

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
		targetPhase     sqlc.PhaseType
		validationError error
		performError    error
		walkError       error
		advanceError    error
		error           string
	}

	tests := []inputType{
		{
			"Should fail when game is not in the correct phase",
			sqlc.PhaseTypeATTACK,
			sqlc.PhaseTypeATTACK,
			nil,
			nil,
			nil,
			nil,
			"game is not in the correct phase to perform move",
		},
		{
			"Should fail when validation state fails",
			sqlc.PhaseTypeDEPLOY,
			sqlc.PhaseTypeDEPLOY,
			fmt.Errorf("validation error"),
			nil,
			nil,
			nil,
			"invalid move: validation error",
		},
		{
			"Should fail when perform function fails",
			sqlc.PhaseTypeDEPLOY,
			sqlc.PhaseTypeDEPLOY,
			nil,
			fmt.Errorf("perform error"),
			nil,
			nil,
			"unable to perform move: perform error",
		},
		{
			"Should fail when unable to walk to target phase",
			sqlc.PhaseTypeDEPLOY,
			sqlc.PhaseTypeDEPLOY,
			nil,
			nil,
			fmt.Errorf("walk error"),
			nil,
			"unable to walk phase: walk error",
		},
		{
			"Should fail when unable to advance phase",
			sqlc.PhaseTypeDEPLOY,
			sqlc.PhaseTypeATTACK,
			nil,
			nil,
			nil,
			fmt.Errorf("advance error"),
			"unable to advance move: advance error",
		},
		{
			"Should succeed and advance phase",
			sqlc.PhaseTypeDEPLOY,
			sqlc.PhaseTypeDEPLOY,
			nil,
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

			querier, gameService, _, validationService, service := setup(t)
			context := input()

			game := &state.Game{
				ID:    context.GameID(),
				Phase: sqlc.PhaseTypeDEPLOY,
				Turn:  2,
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
			}

			performFunc := func(c ctx.MoveContext, querier db.Querier) error {
				return test.performError
			}
			walkerFunc := func(c ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error) {
				return test.targetPhase, test.walkError
			}
			advancerFunc := func(c ctx.MoveContext, querier db.Querier, phase sqlc.PhaseType) error {
				return test.advanceError
			}

			err := service.OrchestrateMoveQ(
				context,
				querier,
				test.phase,
				performFunc,
				walkerFunc,
				advancerFunc,
			)

			if test.phase == sqlc.PhaseTypeDEPLOY &&
				test.validationError == nil &&
				test.performError == nil &&
				test.walkError == nil &&
				test.advanceError == nil {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			require.EqualError(t, err, test.error)
		})
	}
}
