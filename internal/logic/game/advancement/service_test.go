package advancement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/advancement"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	mockdb "github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	mockvalidation "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/orchestration/validation"
	mockmoveservice "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/service"
	mockstate "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/state"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// Using concrete types for the generic Service: string for move type T, any for result type R.
type (
	testMove   = string
	testResult = any
)

func setup(t *testing.T) (
	*mockdb.Querier,
	*mockstate.Service,
	*mockmoveservice.Service[testMove, testResult],
	*mockvalidation.Service,
	*events.TestBus,
	advancement.Service[testMove, testResult],
) {
	t.Helper()

	querier := mockdb.NewQuerier(t)
	gameState := mockstate.NewService(t)
	moveService := mockmoveservice.NewService[testMove, testResult](t)
	validationService := mockvalidation.NewService(t)
	bus := events.NewTestBus()

	service := advancement.NewService[testMove, testResult](
		gameState,
		querier,
		moveService,
		validationService,
		bus,
		nil,
	)

	return querier, gameState, moveService, validationService, bus, service
}

func gameContext() ctx.GameContext {
	userID := "Giovanni"
	gameID := int64(1)

	userContext := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

// matchGameCtx returns a mock.MatchedBy matcher that validates a GameContext
// carries the expected gameID and userID. SpanStep wraps contexts via WithBase,
// so we can't match on exact instance identity.
func matchGameCtx(expected ctx.GameContext) any {
	return mock.MatchedBy(func(actual ctx.GameContext) bool {
		return actual.GameID() == expected.GameID() &&
			actual.UserID() == expected.UserID()
	})
}

func TestAdvanceWithQuerier_HappyPath(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, _, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().Validate(matchGameCtx(gameCtx), querier, game).Return(nil)
	moveService.EXPECT().
		Walk(matchGameCtx(gameCtx), querier, true).
		Return(sqlc.GamePhaseTypeREINFORCE, nil)
	moveService.EXPECT().
		Advance(matchGameCtx(gameCtx), querier, sqlc.GamePhaseTypeREINFORCE, nil).
		Return(nil)

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.NoError(t, err)
	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, targetPhase)
}

func TestAdvanceWithQuerier_GetGameStateFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, _, _, service := setup(t)
	gameCtx := gameContext()

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().
		GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).
		Return(nil, errors.New("db connection lost"))

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "unable to get game state")
	require.ErrorContains(t, err, "db connection lost")
}

func TestAdvanceWithQuerier_ValidationFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, _, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().
		Validate(matchGameCtx(gameCtx), querier, game).
		Return(domainerrors.NewConflictError("it is not the player's turn"))

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "validation failed")

	var conflictErr *domainerrors.DomainError
	require.ErrorAs(t, err, &conflictErr)
}

func TestAdvanceWithQuerier_PhaseMismatchReturnsConflictError(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, _, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeDEPLOY, // actual phase
		Turn:  2,
	}

	// Service expects ATTACK phase but game is in DEPLOY
	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().Validate(matchGameCtx(gameCtx), querier, game).Return(nil)

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)

	var conflictErr *domainerrors.DomainError
	require.ErrorAs(t, err, &conflictErr)
	require.Contains(t, conflictErr.Error(), "DEPLOY")
	require.Contains(t, conflictErr.Error(), "ATTACK")
}

func TestAdvanceWithQuerier_WalkFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, _, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().Validate(matchGameCtx(gameCtx), querier, game).Return(nil)
	moveService.EXPECT().
		Walk(matchGameCtx(gameCtx), querier, true).
		Return(sqlc.GamePhaseType(""), errors.New("walk computation failed"))

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "unable to walk to target phase")
	require.ErrorContains(t, err, "walk computation failed")
}

func TestAdvanceWithQuerier_AdvanceFails(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, _, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().Validate(matchGameCtx(gameCtx), querier, game).Return(nil)
	moveService.EXPECT().
		Walk(matchGameCtx(gameCtx), querier, true).
		Return(sqlc.GamePhaseTypeREINFORCE, nil)
	moveService.EXPECT().
		Advance(matchGameCtx(gameCtx), querier, sqlc.GamePhaseTypeREINFORCE, nil).
		Return(errors.New("advance db write failed"))

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)
	require.ErrorContains(t, err, "unable to perform move")
	require.ErrorContains(t, err, "advance db write failed")
}

func TestAdvanceWithQuerier_DifferentPhaseTransitions(t *testing.T) {
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

			querier, gameState, moveService, validationService, _, service := setup(t)
			gameCtx := gameContext()

			game := &state.Game{
				ID:    gameCtx.GameID(),
				Phase: testCase.fromPhase,
				Turn:  2,
			}

			moveService.EXPECT().PhaseType().Return(testCase.fromPhase)
			gameState.EXPECT().
				GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).
				Return(game, nil)
			validationService.EXPECT().Validate(matchGameCtx(gameCtx), querier, game).Return(nil)
			moveService.EXPECT().
				Walk(matchGameCtx(gameCtx), querier, true).
				Return(testCase.targetPhase, nil)
			moveService.EXPECT().
				Advance(matchGameCtx(gameCtx), querier, testCase.targetPhase, nil).
				Return(nil)

			resultPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

			require.NoError(t, err)
			require.Equal(t, testCase.targetPhase, resultPhase)
		})
	}
}

func TestAdvanceWithQuerier_ForbiddenErrorPropagated(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, _, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().
		Validate(matchGameCtx(gameCtx), querier, game).
		Return(domainerrors.NewForbiddenError("player is not in game"))

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, targetPhase)

	var forbiddenErr *domainerrors.DomainError
	require.ErrorAs(t, err, &forbiddenErr)
}

func TestAdvanceWithQuerier_NoBusEmission(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, validationService, bus, service := setup(t)
	gameCtx := gameContext()

	game := &state.Game{
		ID:    gameCtx.GameID(),
		Phase: sqlc.GamePhaseTypeATTACK,
		Turn:  2,
	}

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).Return(game, nil)
	validationService.EXPECT().Validate(matchGameCtx(gameCtx), querier, game).Return(nil)
	moveService.EXPECT().
		Walk(matchGameCtx(gameCtx), querier, true).
		Return(sqlc.GamePhaseTypeREINFORCE, nil)
	moveService.EXPECT().
		Advance(matchGameCtx(gameCtx), querier, sqlc.GamePhaseTypeREINFORCE, nil).
		Return(nil)

	targetPhase, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.NoError(t, err)
	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, targetPhase)

	// AdvanceWithQuerier does NOT emit bus events — only Advance() does (post-commit)
	require.Empty(t, bus.Events(), "AdvanceWithQuerier must not emit bus events")
}

func TestAdvanceWithQuerier_NoBusEmissionOnError(t *testing.T) {
	t.Parallel()

	querier, gameState, moveService, _, bus, service := setup(t)
	gameCtx := gameContext()

	moveService.EXPECT().PhaseType().Return(sqlc.GamePhaseTypeATTACK)
	gameState.EXPECT().
		GetGameStateWithQuerier(matchGameCtx(gameCtx), querier).
		Return(nil, errors.New("db error"))

	_, err := service.AdvanceWithQuerier(gameCtx, querier)

	require.Error(t, err)
	require.Empty(t, bus.Events(), "no bus events on error path")
}
