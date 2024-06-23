package orchestration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration/orchestration"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/pool"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/move/orchestration/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/move/validation"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	db.Querier,
	*game.Service,
	*phase.Service,
	*validation.Service,
	*orchestration.ServiceImpl,
) {
	t.Helper()
	mockDB := pool.NewDB(t)
	querier := db.New(mockDB, zap.NewNop().Sugar())
	gameService := game.NewService(t)
	phaseService := phase.NewService(t)
	validationService := validation.NewService(t)
	service := orchestration.NewService(
		zap.NewNop().Sugar(),
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

func input() (context.Context, int64, string) {
	gameID := int64(1)
	userID := "23011836-df7e-4421-bdbf-b9c07b22eb64"

	return context.Background(), gameID, userID
}

func TestServiceImpl_PerformMove(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name            string
		phase           sqlc.Phase
		validationError error
		performError    error
		advanceError    error
		error           string
	}

	tests := []inputType{
		{
			"Should fail when game is not in the correct phase",
			sqlc.PhaseATTACK,
			nil,
			nil,
			nil,
			"game is not in the correct phase to perform move",
		},
		{
			"Should fail when validation service fails",
			sqlc.PhaseDEPLOY,
			fmt.Errorf("validation error"),
			nil,
			nil,
			"invalid move: validation error",
		},
		{
			"Should fail when perform function fails",
			sqlc.PhaseDEPLOY,
			nil,
			fmt.Errorf("perform error"),
			nil,
			"unable to perform move: perform error",
		},
		{
			"Should fail when unable to advance phase",
			sqlc.PhaseDEPLOY,
			nil,
			nil,
			fmt.Errorf("advance error"),
			"unable to advance phase: advance error",
		},
		{
			"Should succeed and advance phase",
			sqlc.PhaseDEPLOY,
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
			ctx, gameID, userID := input()

			game := &sqlc.Game{
				ID:               gameID,
				Phase:            sqlc.PhaseDEPLOY,
				Turn:             2,
				DeployableTroops: 5,
			}

			gameService.
				EXPECT().
				GetGameStateQ(ctx, querier, gameID).
				Return(game, nil)

			if test.phase == sqlc.PhaseDEPLOY {
				validationService.
					EXPECT().
					Validate(ctx, querier, game, userID).
					Return(test.validationError)
				if test.validationError == nil && test.performError == nil {
					phaseService.
						EXPECT().
						AdvanceQ(ctx, querier, gameID).
						Return(test.advanceError)
				}
			}

			performFunc := func(ctx context.Context, querier db.Querier, game *sqlc.Game) error {
				return test.performError
			}
			err := service.PerformMoveQ(
				ctx,
				querier,
				gameID,
				test.phase,
				userID,
				performFunc,
			)

			if test.phase == sqlc.PhaseDEPLOY &&
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
