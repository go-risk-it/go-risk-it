package orchestration

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
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
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Orchestrator[T any] interface {
	OrchestrateMove(ctx ctx.GameContext, move T) error
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

func (s *orchestrator[T]) OrchestrateMove(ctx ctx.GameContext, move T) error {
	_, span := otel.GetTracerProvider().Tracer("go-risk-it-game").Start(
		ctx, "game.orchestrate_move",
		trace.WithAttributes(
			attribute.String("phase", string(s.service.PhaseType())),
		),
	)
	defer span.End()

	start := time.Now()

	targetPhase, err := dbutil.InTransactionWithIsolation(
		s.querier,
		ctx,
		s.metrics,
		pgx.RepeatableRead,
		func(querier db.Querier) (sqlc.GamePhaseType, error) {
			phase := s.service.PhaseType()

			gameState, err := s.gameService.GetGameStateWithQuerier(ctx, querier)
			if err != nil {
				return "", fmt.Errorf("unable to get game state: %w", err)
			}

			if gameState.Phase != phase {
				return "", domainerrors.NewConflictErrorf(
					"game is in phase %s, expected %s", gameState.Phase, phase,
				)
			}

			resultPhase, err := s.orchestrateMoveWithQuerier(ctx, querier, move, gameState)
			if err != nil {
				return "", fmt.Errorf("unable to orchestrate move: %w", err)
			}

			return resultPhase, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	phase := string(s.service.PhaseType())
	phaseAttr := metric.WithAttributes(attribute.String("phase", phase))

	s.metrics.MovesTotal.Add(ctx, 1, phaseAttr)
	s.metrics.PhaseDuration.Record(ctx, time.Since(start).Seconds(), phaseAttr)

	s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		FromPhase: s.service.PhaseType(),
		ToPhase:   targetPhase,
	})

	return nil
}

func (s *orchestrator[T]) orchestrateMoveWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
	move T,
	gameState *state.Game,
) (sqlc.GamePhaseType, error) {
	slog.InfoContext(ctx, "orchestrating move", "move", move)

	if err := s.validationService.Validate(ctx, querier, gameState); err != nil {
		return "", fmt.Errorf("invalid move: %w", err)
	}

	performResult, err := s.service.Perform(ctx, querier, move)
	if err != nil {
		return "", fmt.Errorf("unable to perform move: %w", err)
	}

	if err := s.loggingService.LogMove(ctx, querier, move, performResult); err != nil {
		return "", fmt.Errorf("unable to log move: %w", err)
	}

	isMissionAccomplished, err := s.missionService.IsMissionAccomplished(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("unable to check if mission is accomplished: %w", err)
	}

	if isMissionAccomplished {
		slog.InfoContext(ctx, "game is over")
		s.recordGameFinished(ctx)

		return s.service.PhaseType(), nil
	}

	targetPhase, err := s.service.Walk(ctx, querier, false)
	if err != nil {
		return "", fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		slog.InfoContext(ctx, "no need to advance")

		return targetPhase, nil
	}

	slog.InfoContext(ctx, "advancing phase", "target", targetPhase)

	if err := s.service.Advance(ctx, querier, targetPhase, performResult); err != nil {
		return "", fmt.Errorf("unable to advance move: %w", err)
	}

	slog.InfoContext(ctx, "successfully advanced phase", "target", targetPhase)

	return targetPhase, nil
}

func (s *orchestrator[T]) recordGameFinished(ctx ctx.GameContext) {
	s.metrics.GamesFinished.Add(ctx, 1)
	s.metrics.ActiveGames.Add(ctx, -1)

	if elapsed, ok := s.gameTiming.ElapsedAndClear(ctx.GameID()); ok {
		s.metrics.GameDuration.Record(ctx, elapsed.Seconds())
	}
}
