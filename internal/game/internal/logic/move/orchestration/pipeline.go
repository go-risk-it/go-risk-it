package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
)

func (s *orchestrator[T, R]) OrchestrateMove(
	ctx gamectx.GameContext,
	move T,
) error {
	return observe.SpanErr(ctx, "game.orchestrate_move", func(ctx gamectx.GameContext) error {
		outcome, err := s.executeTransaction(ctx, move)
		if err != nil {
			return fmt.Errorf("unable to perform move: %w", err)
		}

		s.stateStore.Store(ctx.GameID(), outcome.newState)

		s.emitMoveCompleted(ctx, outcome)

		return nil
	}, attribute.String("phase", string(s.service.PhaseType())))
}

func (s *orchestrator[T, R]) executeTransaction(
	ctx gamectx.GameContext,
	move T,
) (moveOutcome[R], error) {
	return dbutil.InTransactionWithIsolation(
		s.querier, ctx, s.stateMetrics, pgx.RepeatableRead,
		func(querier db.Querier) (moveOutcome[R], error) {
			phase := s.service.PhaseType()

			gameState, err := s.gameService.GetGameStateWithQuerier(
				ctx, querier,
			)
			if err != nil {
				return moveOutcome[R]{}, fmt.Errorf(
					"unable to get game state: %w", err,
				)
			}

			if gameState.Phase != phase {
				return moveOutcome[R]{}, domainerrors.NewConflictErrorf(
					"game is in phase %s, expected %s",
					gameState.Phase, phase,
				)
			}

			outcome, err := s.orchestrateMoveWithQuerier(
				ctx, querier, move, gameState,
			)
			if err != nil {
				return moveOutcome[R]{}, err
			}

			return outcome, nil
		},
	)
}

func (s *orchestrator[T, R]) orchestrateMoveWithQuerier(
	ctx gamectx.GameContext,
	querier db.Querier,
	move T,
	gameState *state.Game,
) (moveOutcome[R], error) {
	turn := gameState.Turn

	observe.SpanEvent(ctx, "orchestrating_move")

	// Get or warm the previous state from the cache (warm-on-miss via DB).
	prevState, err := s.getOrWarmPrevState(ctx, querier, turn)
	if err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to get previous state: %w", err)
	}

	if err := s.validationService.Validate(
		ctx, querier, gameState,
	); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("invalid move: %w", err)
	}

	observe.SpanEvent(ctx, "game.move.validated")

	performResult, effect, moveLog, err := s.performAndLog(
		ctx, querier, move, prevState,
	)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	targetPhase, gameOver, advEffect, err := s.resolveTargetPhase(
		ctx, querier, performResult, prevState, effect,
	)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	return s.buildOutcome(
		ctx, prevState, effect, advEffect,
		targetPhase, gameOver, performResult, moveLog, turn,
	), nil
}

// buildOutcome assembles the moveOutcome by computing the new state from
// enrichment data via BuildNewState. No post-mutation DB reads occur here.
func (s *orchestrator[T, R]) buildOutcome(
	ctx gamectx.GameContext,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
	advEffect moveservice.AdvanceEffect,
	targetPhase sqlc.GamePhaseType,
	gameOver bool,
	performResult R,
	moveLog sqlc.GameMoveLog,
	turn int64,
) moveOutcome[R] {
	winnerUserID := ""
	if gameOver {
		winnerUserID = ctx.UserID()
	}

	var advEffectPtr *moveservice.AdvanceEffect
	if targetPhase != s.service.PhaseType() && !gameOver {
		advEffectPtr = &advEffect
	}

	newState := BuildNewState(
		prevState,
		&effect,
		advEffectPtr,
		targetPhase,
		winnerUserID,
	)

	return moveOutcome[R]{
		targetPhase: targetPhase,
		gameOver:    gameOver,
		result:      performResult,
		moveLog:     moveLog,
		turn:        turn,
		newState:    newState,
		prevRegions: prevState.PublicSnapshot.Regions,
	}
}

func (s *orchestrator[T, R]) performAndLog(
	ctx gamectx.GameContext,
	querier db.Querier,
	move T,
	prevState *snapshot.CachedGameState,
) (R, moveservice.MoveEffect, sqlc.GameMoveLog, error) {
	performResult, effect, err := s.service.Perform(ctx, querier, move, prevState)
	if err != nil {
		var zero R

		return zero, moveservice.MoveEffect{}, sqlc.GameMoveLog{}, fmt.Errorf(
			"unable to perform move: %w", err,
		)
	}

	observe.SpanEvent(ctx, "game.move.performed")

	moveLog, err := s.loggingService.LogMove(ctx, querier, move, performResult)
	if err != nil {
		var zero R

		return zero, moveservice.MoveEffect{}, sqlc.GameMoveLog{}, fmt.Errorf(
			"unable to log move: %w", err,
		)
	}

	observe.SpanEvent(ctx, "game.move.logged")

	return performResult, effect, moveLog, nil
}

func (s *orchestrator[T, R]) resolveTargetPhase(
	ctx gamectx.GameContext,
	querier db.Querier,
	performResult R,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
) (sqlc.GamePhaseType, bool, moveservice.AdvanceEffect, error) {
	if accomplished, err := s.checkMission(
		ctx, querier,
	); err != nil {
		return "", false, moveservice.AdvanceEffect{}, err
	} else if accomplished {
		return s.service.PhaseType(), true, moveservice.AdvanceEffect{}, nil
	}

	// Universal domination: if the player owns all regions after the move,
	// the game is won regardless of their specific mission.
	if IsDomination(prevState, effect, ctx.UserID()) {
		if err := s.assignWinner(ctx, querier); err != nil {
			return "", false, moveservice.AdvanceEffect{}, err
		}

		return s.service.PhaseType(), true, moveservice.AdvanceEffect{}, nil
	}

	targetPhase, advEffect, err := s.walkAndAdvance(
		ctx, querier, performResult, prevState, effect,
	)
	if err != nil {
		return "", false, moveservice.AdvanceEffect{}, err
	}

	return targetPhase, false, advEffect, nil
}

func (s *orchestrator[T, R]) walkAndAdvance(
	ctx gamectx.GameContext,
	querier db.Querier,
	performResult R,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
) (sqlc.GamePhaseType, moveservice.AdvanceEffect, error) {
	walkCtx := buildWalkContext(prevState, effect, false, ctx.UserID())

	targetPhase, err := s.service.Walk(walkCtx)
	if err != nil {
		return "", moveservice.AdvanceEffect{}, fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		return targetPhase, moveservice.AdvanceEffect{}, nil
	}

	advCtx, err := s.buildAdvanceContext(ctx, prevState, effect, ctx.UserID())
	if err != nil {
		return "", moveservice.AdvanceEffect{}, fmt.Errorf(
			"unable to build advance context: %w",
			err,
		)
	}

	advEffect, err := s.service.Advance(
		ctx, querier, targetPhase, performResult, advCtx,
	)
	if err != nil {
		return "", moveservice.AdvanceEffect{}, fmt.Errorf("unable to advance move: %w", err)
	}

	return targetPhase, advEffect, nil
}
