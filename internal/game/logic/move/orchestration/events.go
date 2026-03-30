package orchestration

import (
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
)

func (s *orchestrator[T, R]) emitEvents(
	ctx gamectx.GameContext,
	outcome moveOutcome[R],
) {
	now := time.Now()
	attackResult, cardsResult := extractResults(outcome.result, s.service.PhaseType())

	s.bus.Emit(ctx, gameevt.NewMoveExecuted(
		ctx.GameID(),
		ctx.UserID(),
		now,
		s.service.PhaseType(),
		outcome.moveLog,
		outcome.targetPhase,
		outcome.gameOver,
		outcome.turn,
		attackResult,
		cardsResult,
	))

	if outcome.targetPhase != s.service.PhaseType() {
		s.bus.Emit(ctx, gameevt.NewPhaseTransitioned(
			ctx.GameID(),
			ctx.UserID(),
			now,
			s.service.PhaseType(),
			outcome.targetPhase,
			outcome.turn,
		))
	}

	if outcome.gameOver {
		s.bus.Emit(ctx, gameevt.NewGameCompleted(
			ctx.GameID(),
			ctx.UserID(),
			now,
			outcome.turn,
		))
	}
}

// extractResults bridges generic R to concrete action-specific result types.
// It switches on phaseType (not R) to determine the correct type assertion.
func extractResults[R any](
	result R,
	phaseType sqlc.GamePhaseType,
) (*attack.MoveResult, *cards.MoveResult) {
	switch phaseType {
	case sqlc.GamePhaseTypeATTACK:
		if ar, ok := any(result).(*attack.MoveResult); ok {
			return ar, nil
		}
	case sqlc.GamePhaseTypeCARDS:
		if cr, ok := any(result).(*cards.MoveResult); ok {
			return nil, cr
		}
	default:
		// DEPLOY, CONQUER, REINFORCE produce no typed result.
	}

	return nil, nil
}
