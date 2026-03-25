package orchestration

import (
	"fmt"
	"log/slog"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/timing"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/tracing"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

type Orchestrator[T any] interface {
	OrchestrateMove(ctx gamectx.GameContext, move T) error
}

type orchestrator[T any] struct {
	querier                db.Querier
	service                service.Service[T]
	gameService            state.Service
	loggingService         logging.Service
	missionService         mission.Service
	validationService      validation.Service
	gameStateChangedSignal signals.GameStateChangedSignal
	metrics                *metrics.Metrics
	gameTiming             *timing.GameTiming
}

var _ Orchestrator[any] = (*orchestrator[any])(nil)

func NewOrchestrator[T any](
	querier db.Querier,
	service service.Service[T],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) Orchestrator[T] {
	return &orchestrator[T]{
		querier:                querier,
		service:                service,
		gameService:            gameService,
		loggingService:         loggingService,
		missionService:         missionService,
		validationService:      validationService,
		gameStateChangedSignal: gameStateChangedSignal,
		metrics:                metrics,
		gameTiming:             gameTiming,
	}
}

func (s *orchestrator[T]) OrchestrateMove(
	ctx gamectx.GameContext,
	move T,
) error {
	ctx, span := tracing.StartGameSpan(ctx, "game.orchestrate_move",
		attribute.String("phase", string(s.service.PhaseType())),
	)
	defer span.End()

	start := time.Now()

	targetPhase, err := s.executeTransaction(ctx, move)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.recordMetrics(ctx, start)
	s.emitSignal(ctx, targetPhase)

	return nil
}

func (s *orchestrator[T]) executeTransaction(
	ctx gamectx.GameContext,
	move T,
) (sqlc.GamePhaseType, error) {
	return dbutil.InTransactionWithIsolation(
		s.querier, ctx, s.metrics, pgx.RepeatableRead,
		func(querier db.Querier) (sqlc.GamePhaseType, error) {
			phase := s.service.PhaseType()

			gameState, err := s.gameService.GetGameStateWithQuerier(
				ctx, querier,
			)
			if err != nil {
				return "", fmt.Errorf(
					"unable to get game state: %w", err,
				)
			}

			if gameState.Phase != phase {
				return "", domainerrors.NewConflictErrorf(
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

func (s *orchestrator[T]) recordMetrics(
	ctx gamectx.GameContext,
	start time.Time,
) {
	phase := string(s.service.PhaseType())
	phaseAttr := metric.WithAttributes(
		attribute.String("phase", phase),
	)

	s.metrics.MovesTotal.Add(ctx, 1, phaseAttr)
	s.metrics.PhaseDuration.Record(
		ctx, time.Since(start).Seconds(), phaseAttr,
	)
}

func (s *orchestrator[T]) emitSignal(
	ctx gamectx.GameContext,
	targetPhase sqlc.GamePhaseType,
) {
	_, signalSpan := tracing.StartGameSpan(
		ctx, "game.move.signal",
		attribute.String("phase", string(s.service.PhaseType())),
		attribute.String("target_phase", string(targetPhase)),
	)

	s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		FromPhase: s.service.PhaseType(),
		ToPhase:   targetPhase,
	})
	signalSpan.End()
}

func (s *orchestrator[T]) orchestrateMoveWithQuerier(
	ctx gamectx.GameContext,
	querier db.Querier,
	move T,
	gameState *state.Game,
) (sqlc.GamePhaseType, error) {
	phase := string(s.service.PhaseType())
	slog.InfoContext(ctx, "orchestrating move", "move", move)

	if err := tracing.SpanStep(
		ctx, "game.move.validate", phase,
		func(spanCtx gamectx.GameContext) error {
			return s.validationService.Validate(
				spanCtx, querier, gameState,
			)
		},
	); err != nil {
		return "", fmt.Errorf("invalid move: %w", err)
	}

	performResult, err := s.performAndLog(
		ctx, querier, move, phase,
	)
	if err != nil {
		return "", err
	}

	return s.resolveTargetPhase(
		ctx, querier, phase, performResult,
	)
}

func (s *orchestrator[T]) performAndLog(
	ctx gamectx.GameContext,
	querier db.Querier,
	move T,
	phase string,
) (any, error) {
	var performResult any

	if err := tracing.SpanStep(
		ctx, "game.move.perform", phase,
		func(spanCtx gamectx.GameContext) error {
			var performErr error
			performResult, performErr = s.service.Perform(
				spanCtx, querier, move,
			)

			return performErr
		},
	); err != nil {
		return nil, fmt.Errorf("unable to perform move: %w", err)
	}

	if err := tracing.SpanStep(
		ctx, "game.move.log", phase,
		func(spanCtx gamectx.GameContext) error {
			return s.loggingService.LogMove(
				spanCtx, querier, move, performResult,
			)
		},
	); err != nil {
		return nil, fmt.Errorf("unable to log move: %w", err)
	}

	return performResult, nil
}

func (s *orchestrator[T]) resolveTargetPhase(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
	performResult any,
) (sqlc.GamePhaseType, error) {
	if accomplished, err := s.checkMission(
		ctx, querier, phase,
	); err != nil {
		return "", err
	} else if accomplished {
		return s.service.PhaseType(), nil
	}

	return s.walkAndAdvance(
		ctx, querier, phase, performResult,
	)
}

func (s *orchestrator[T]) checkMission(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
) (bool, error) {
	var isMissionAccomplished bool

	if err := tracing.SpanStep(
		ctx, "game.move.check_mission", phase,
		func(spanCtx gamectx.GameContext) error {
			var missionErr error
			isMissionAccomplished, missionErr = s.missionService.IsMissionAccomplished(
				spanCtx, querier,
			)

			return missionErr
		},
	); err != nil {
		return false, fmt.Errorf(
			"unable to check if mission is accomplished: %w", err,
		)
	}

	if isMissionAccomplished {
		slog.InfoContext(ctx, "game is over")
		s.recordGameFinished(ctx)
	}

	return isMissionAccomplished, nil
}

func (s *orchestrator[T]) walkAndAdvance(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
	performResult any,
) (sqlc.GamePhaseType, error) {
	var targetPhase sqlc.GamePhaseType

	if err := tracing.SpanStep(
		ctx, "game.move.walk", phase,
		func(spanCtx gamectx.GameContext) error {
			var walkErr error
			targetPhase, walkErr = s.service.Walk(
				spanCtx, querier, false,
			)

			return walkErr
		},
	); err != nil {
		return "", fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		slog.InfoContext(ctx, "no need to advance")

		return targetPhase, nil
	}

	slog.InfoContext(ctx, "advancing phase", "target", targetPhase)

	if err := tracing.SpanStep(
		ctx, "game.move.advance", phase,
		func(spanCtx gamectx.GameContext) error {
			return s.service.Advance(
				spanCtx, querier, targetPhase, performResult,
			)
		},
	); err != nil {
		return "", fmt.Errorf("unable to advance move: %w", err)
	}

	slog.InfoContext(ctx, "successfully advanced phase",
		"target", targetPhase,
	)

	return targetPhase, nil
}

func (s *orchestrator[T]) recordGameFinished(
	ctx gamectx.GameContext,
) {
	s.metrics.GamesFinished.Add(ctx, 1)
	s.metrics.ActiveGames.Add(ctx, -1)

	if elapsed, ok := s.gameTiming.ElapsedAndClear(ctx.GameID()); ok {
		s.metrics.GameDuration.Record(ctx, elapsed.Seconds())
	}
}
