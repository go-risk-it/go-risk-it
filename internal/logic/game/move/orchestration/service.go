package orchestration

import (
	"fmt"
	"log/slog"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/timing"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/tracing"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

type Orchestrator[T, R any] interface {
	OrchestrateMove(ctx gamectx.GameContext, move T) error
}

// moveOutcome carries the committed transaction result for post-commit event emission.
type moveOutcome[R any] struct {
	targetPhase sqlc.GamePhaseType
	gameOver    bool
	result      R
	moveLog     sqlc.GameMoveLog
	turn        int64
}

type orchestrator[T, R any] struct {
	querier           db.Querier
	service           service.Service[T, R]
	gameService       state.Service
	loggingService    logging.Service
	missionService    mission.Service
	validationService validation.Service
	bus               events.Bus
	metrics           *metrics.Metrics
	gameTiming        *timing.GameTiming
}

var _ Orchestrator[any, any] = (*orchestrator[any, any])(nil)

func NewOrchestrator[T, R any](
	querier db.Querier,
	service service.Service[T, R],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	bus events.Bus,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) Orchestrator[T, R] {
	return &orchestrator[T, R]{
		querier:           querier,
		service:           service,
		gameService:       gameService,
		loggingService:    loggingService,
		missionService:    missionService,
		validationService: validationService,
		bus:               bus,
		metrics:           metrics,
		gameTiming:        gameTiming,
	}
}

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
		s.querier, ctx, s.metrics, pgx.RepeatableRead,
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

func (s *orchestrator[T, R]) recordMetrics(
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

func (s *orchestrator[T, R]) orchestrateMoveWithQuerier(
	ctx gamectx.GameContext,
	querier db.Querier,
	move T,
	gameState *state.Game,
) (moveOutcome[R], error) {
	phase := string(s.service.PhaseType())
	turn := gameState.Turn
	slog.InfoContext(ctx, "orchestrating move", "move", move)

	if err := tracing.SpanStep(
		ctx, "game.move.validate", phase,
		func(spanCtx gamectx.GameContext) error {
			return s.validationService.Validate(
				spanCtx, querier, gameState,
			)
		},
	); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("invalid move: %w", err)
	}

	performResult, moveLog, err := s.performAndLog(
		ctx, querier, move, phase,
	)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	targetPhase, gameOver, err := s.resolveTargetPhase(
		ctx, querier, phase, performResult,
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
	phase string,
) (R, sqlc.GameMoveLog, error) {
	var performResult R

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
		var zero R

		return zero, sqlc.GameMoveLog{}, fmt.Errorf("unable to perform move: %w", err)
	}

	var moveLog sqlc.GameMoveLog

	if err := tracing.SpanStep(
		ctx, "game.move.log", phase,
		func(spanCtx gamectx.GameContext) error {
			var logErr error
			moveLog, logErr = s.loggingService.LogMove(
				spanCtx, querier, move, performResult,
			)

			return logErr
		},
	); err != nil {
		var zero R

		return zero, sqlc.GameMoveLog{}, fmt.Errorf("unable to log move: %w", err)
	}

	return performResult, moveLog, nil
}

func (s *orchestrator[T, R]) resolveTargetPhase(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
	performResult R,
) (sqlc.GamePhaseType, bool, error) {
	if accomplished, err := s.checkMission(
		ctx, querier, phase,
	); err != nil {
		return "", false, err
	} else if accomplished {
		return s.service.PhaseType(), true, nil
	}

	targetPhase, err := s.walkAndAdvance(
		ctx, querier, phase, performResult,
	)
	if err != nil {
		return "", false, err
	}

	return targetPhase, false, nil
}

func (s *orchestrator[T, R]) checkMission(
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

func (s *orchestrator[T, R]) walkAndAdvance(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
	performResult R,
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

func (s *orchestrator[T, R]) recordGameFinished(
	ctx gamectx.GameContext,
) {
	s.metrics.GamesFinished.Add(ctx, 1)
	s.metrics.ActiveGames.Add(ctx, -1)

	if elapsed, ok := s.gameTiming.ElapsedAndClear(ctx.GameID()); ok {
		s.metrics.GameDuration.Record(ctx, elapsed.Seconds())
	}
}
