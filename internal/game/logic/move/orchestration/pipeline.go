package orchestration

import (
	"fmt"
	"log/slog"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s *orchestrator[T, R]) OrchestrateMove(
	ctx gamectx.GameContext,
	move T,
) error {
	ctx, span := tracing.StartGameSpan(ctx, "game.orchestrate_move",
		attribute.String("phase", string(s.service.PhaseType())),
	)
	defer span.End()

	start := time.Now()

	outcome, err := s.executeTransaction(ctx, move)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.recordMetrics(ctx, start)
	s.emitEvents(ctx, outcome)

	return nil
}

func (s *orchestrator[T, R]) executeTransaction(
	ctx gamectx.GameContext,
	move T,
) (moveOutcome[R], error) {
	return dbutil.InTransactionWithIsolation(
		s.querier, ctx, s.infraMetrics, pgx.RepeatableRead,
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
	slog.InfoContext(ctx, "orchestrating move", "move", move)

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
		slog.InfoContext(ctx, "no need to advance")

		return targetPhase, nil
	}

	slog.InfoContext(ctx, "advancing phase", "target", targetPhase)

	if err := s.service.Advance(
		ctx, querier, targetPhase, performResult,
	); err != nil {
		return "", fmt.Errorf("unable to advance move: %w", err)
	}

	slog.InfoContext(ctx, "successfully advanced phase",
		"target", targetPhase,
	)

	return targetPhase, nil
}
