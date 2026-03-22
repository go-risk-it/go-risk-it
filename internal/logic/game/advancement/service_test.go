package advancement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/advancement"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	mockdb "github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	mockvalidation "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/orchestration/validation"
	mockmoveservice "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/service"
	mocksignals "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/signals"
	mockstate "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/state"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

// Using a concrete type for the generic Service: string for move type T.
type testMove = string

func setup(t *testing.T) (
	*mockdb.Querier,
	*mockstate.Service,
	*mockmoveservice.Service[testMove],
	*mockvalidation.Service,
	*advancement.ServiceImpl[testMove],
) {
	t.Helper()

	querier := mockdb.NewQuerier(t)
	gameState := mockstate.NewService(t)
	moveService := mockmoveservice.NewService[testMove](t)
	validationService := mockvalidation.NewService(t)
	signal := mocksignals.NewGameStateChangedSignal(t)

	service := advancement.NewService[testMove](
		gameState,
		querier,
		moveService,
		validationService,
		signal,
	)

	return querier, gameState, moveService, validationService, service
}

func gameContext() ctx.GameContext {
	userID := "Giovanni"
	gameID := int64(1)

	userContext := ctx.WithUserID(
		ctx.WithSpan(ctx.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

func TestAdvanceQ_HappyPath(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
	validationService.EXPECT().ValidateQ(gameCtx, querier, game).Return(nil)
	moveService.EXPECT().WalkQ(gameCtx, querier, true).Return(sqlc.GamePhaseTypeREINFORCE, nil)
	moveService.EXPECT().
		AdvanceQ(gameCtx, querier, sqlc.GamePhaseTypeREINFORCE, nil).
		Return(nil)

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.NoError(t, err)
	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, targetPhase)
}

func TestAdvanceQ_GetGameStateFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, _, service := setup(t)
	gameCtx := gameContext()

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().
		GetGameStateQ(gameCtx, querier).
		Return(nil, errors.New("db connection lost"))

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "unable to get game state")
	require.ErrorContains(t, err, "db connection lost")
}

func TestAdvanceQ_ValidationFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
	validationService.EXPECT().
		ValidateQ(gameCtx, querier, game).
		Return(domainerrors.NewConflictError("it is not the player's turn"))

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "validation failed")

	var conflictErr *domainerrors.ConflictError
	require.ErrorAs(t, err, &conflictErr)
}

func TestAdvanceQ_PhaseMismatchReturnsConflictError(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeDEPLOY, // actual phase
		Turn:  2,
	}

	// Service expects ATTACK phase but game is in DEPLOY
	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
	validationService.EXPECT().ValidateQ(gameCtx, querier, game).Return(nil)

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)

	var conflictErr *domainerrors.ConflictError
	require.ErrorAs(t, err, &conflictErr)
	require.Contains(t, conflictErr.Error(), "DEPLOY")
	require.Contains(t, conflictErr.Error(), "ATTACK")
}

func TestAdvanceQ_WalkFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
	validationService.EXPECT().ValidateQ(gameCtx, querier, game).Return(nil)
	moveService.EXPECT().
		WalkQ(gameCtx, querier, true).
		Return(sqlc.GamePhaseType(""), errors.New("walk computation failed"))

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "unable to walk to target phase")
	require.ErrorContains(t, err, "walk computation failed")
}

func TestAdvanceQ_AdvanceFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
	validationService.EXPECT().ValidateQ(gameCtx, querier, game).Return(nil)
	moveService.EXPECT().WalkQ(gameCtx, querier, true).Return(sqlc.GamePhaseTypeREINFORCE, nil)
	moveService.EXPECT().
		AdvanceQ(gameCtx, querier, sqlc.GamePhaseTypeREINFORCE, nil).
		Return(errors.New("advance db write failed"))

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "unable to perform move")
	require.ErrorContains(t, err, "advance db write failed")
}

func TestAdvanceQ_DifferentPhaseTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fromPhase   sqlc.GamePhaseType
		targetPhase sqlc.GamePhaseType
	}{
		{
			name:        "CARDS to DEPLOY",
			fromPhase:   sqlc.GamePhaseTypeCARDS,
			targetPhase: sqlc.GamePhaseTypeDEPLOY,
		},
		{
			name:        "DEPLOY to ATTACK",
			fromPhase:   sqlc.GamePhaseTypeDEPLOY,
			targetPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name:        "ATTACK to REINFORCE",
			fromPhase:   sqlc.GamePhaseTypeATTACK,
			targetPhase: sqlc.GamePhaseTypeREINFORCE,
		},
		{
			name:        "REINFORCE to CARDS",
			fromPhase:   sqlc.GamePhaseTypeREINFORCE,
			targetPhase: sqlc.GamePhaseTypeCARDS,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			querier, gameState, moveService, validationService, service := setup(t)
			gameCtx := gameContext()

			game := &state.Game{
				ID:    gameCtx.GameID(),
				Phase: testCase.fromPhase,
				Turn:  2,
			}

			moveService.EXPECT().PhaseType().Return(testCase.fromPhase)
			gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
			validationService.EXPECT().ValidateQ(gameCtx, querier, game).Return(nil)
			moveService.EXPECT().WalkQ(gameCtx, querier, true).Return(testCase.targetPhase, nil)
			moveService.EXPECT().
				AdvanceQ(gameCtx, querier, testCase.targetPhase, nil).
				Return(nil)

			resultPhase, err := service.AdvanceQ(gameCtx, querier)

			require.NoError(t, err)
			require.Equal(t, testCase.targetPhase, resultPhase)
		})
	}
}

func TestAdvanceQ_ForbiddenErrorPropagated(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateQ(gameCtx, querier).Return(game, nil)
	validationService.EXPECT().
		ValidateQ(gameCtx, querier, game).
		Return(domainerrors.NewForbiddenError("player is not in game"))

	targetPhase, err := service.AdvanceQ(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)

	var forbiddenErr *domainerrors.ForbiddenError
	require.ErrorAs(t, err, &forbiddenErr)
}
