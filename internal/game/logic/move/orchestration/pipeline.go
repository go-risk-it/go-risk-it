package orchestration

import (
	"fmt"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
)

//nolint:nonamedreturns // named returns needed for defer-based error recording
func (s *orchestrator[T, R]) OrchestrateMove(
	ctx gamectx.GameContext,
	move T,
) (err error) {
	ctx, done := observe.TypedSpan(ctx, "game.orchestrate_move",
		attribute.String("phase", string(s.service.PhaseType())),
	)
	defer func() { done(err) }()

	outcome, err := s.executeTransaction(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.emitEvents(ctx, outcome)

	return nil
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

			return s.orchestrateMoveWithQuerier(
				ctx, querier, move, gameState,
			)
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

	if err := s.validationService.Validate(
		ctx, querier, gameState,
	); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("invalid move: %w", err)
	}

	performResult, moveLog, err := s.performAndLog(
		ctx, querier, move,
	)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	targetPhase, gameOver, err := s.resolveTargetPhase(
		ctx, querier, performResult,
	)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	return moveOutcome[R]{
		targetPhase: targetPhase,
		gameOver:    gameOver,
		result:      performResult,
		moveLog:     moveLog,
		turn:        turn,
	}, nil
}

func (s *orchestrator[T, R]) performAndLog(
	ctx gamectx.GameContext,
	querier db.Querier,
	move T,
) (R, sqlc.GameMoveLog, error) {
	performResult, err := s.service.Perform(ctx, querier, move)
	if err != nil {
		var zero R

		return zero, sqlc.GameMoveLog{}, fmt.Errorf("unable to perform move: %w", err)
	}

	moveLog, err := s.loggingService.LogMove(ctx, querier, move, performResult)
	if err != nil {
		var zero R

		return zero, sqlc.GameMoveLog{}, fmt.Errorf("unable to log move: %w", err)
	}

	return performResult, moveLog, nil
}

func (s *orchestrator[T, R]) resolveTargetPhase(
	ctx gamectx.GameContext,
	querier db.Querier,
	performResult R,
) (sqlc.GamePhaseType, bool, error) {
	if accomplished, err := s.checkMission(
		ctx, querier,
	); err != nil {
		return "", false, err
	} else if accomplished {
		return s.service.PhaseType(), true, nil
	}

	targetPhase, err := s.walkAndAdvance(
		ctx, querier, performResult,
	)
	if err != nil {
		return "", false, err
	}

	return targetPhase, false, nil
}

func (s *orchestrator[T, R]) walkAndAdvance(
	ctx gamectx.GameContext,
	querier db.Querier,
	performResult R,
) (sqlc.GamePhaseType, error) {
	targetPhase, err := s.service.Walk(ctx, querier, false)
	if err != nil {
		return "", fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		return targetPhase, nil
	}

	if err := s.service.Advance(
		ctx, querier, targetPhase, performResult,
	); err != nil {
		return "", fmt.Errorf("unable to advance move: %w", err)
	}

	return targetPhase, nil
}
