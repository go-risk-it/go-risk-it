package orchestration

import (
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
)

// emitMoveCompleted emits the ECST-enriched MoveCompleted event that carries
// the full snapshot payload. Snapshots come from the computed newState
// (BuildNewState output), and previousRegions come from the immutable prevState
// captured before the move. This is the single post-commit event for the
// orchestration pipeline — all downstream handlers consume MoveCompleted.
//
// When the game is over, a GameCompleted event is emitted after MoveCompleted
// to produce the bus:game_completed span for dashboard metrics.
func (s *orchestrator[T, R]) emitMoveCompleted(
	ctx gamectx.GameContext,
	outcome moveOutcome[R],
) {
	s.bus.Emit(ctx, gameevt.NewMoveCompleted(
		ctx.GameID(),
		ctx.UserID(),
		time.Now(),
		toAPIPhase(s.service.PhaseType()),
		outcome.turn,
		toAPIPhase(s.service.PhaseType()),
		toAPIPhase(outcome.targetPhase),
		outcome.gameOver,
		outcome.newState.PublicSnapshot,
		outcome.newState.PrivateSnapshots,
		outcome.prevRegions,
	))

	if outcome.gameOver {
		s.bus.Emit(ctx, gameevt.NewGameCompleted(
			ctx.GameID(),
			ctx.UserID(),
			time.Now(),
			outcome.turn,
		))
	}
}

// toAPIPhase converts a sqlc.GamePhaseType to the api GamePhaseType.
// Both are string-based types with identical constant values.
func toAPIPhase(p sqlc.GamePhaseType) gameapi.GamePhaseType {
	return gameapi.GamePhaseType(p)
}
