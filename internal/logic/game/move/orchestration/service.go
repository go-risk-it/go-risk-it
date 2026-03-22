package orchestration

import (
	"fmt"
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
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Orchestrator[T any] interface {
	OrchestrateMove(ctx ctx.GameContext, move T) error
}

type OrchestratorImpl[T any] struct {
	querier                db.Querier
	service                service.Service[T]
	gameService            state.Service
	loggingService         logging.Service
	missionService         mission.Service
	validationService      validation.Service
	gameStateChangedSignal signals.GameStateChangedSignal
	metrics                *metrics.Metrics
}

func NewOrchestrator[T any](
	querier db.Querier,
	service service.Service[T],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
) *OrchestratorImpl[T] {
	return &OrchestratorImpl[T]{
		querier:                querier,
		service:                service,
		gameService:            gameService,
		loggingService:         loggingService,
		missionService:         missionService,
		validationService:      validationService,
		gameStateChangedSignal: gameStateChangedSignal,
		metrics:                metrics,
	}
}

func (s *OrchestratorImpl[T]) OrchestrateMove(ctx ctx.GameContext, move T) error {
	start := time.Now()

	targetPhase, err := dbutil.InTransactionWithIsolation(
		s.querier,
		ctx,
		pgx.RepeatableRead,
		func(querier db.Querier) (sqlc.GamePhaseType, error) {
			phase := s.service.PhaseType()
			ctx.SetLog(ctx.Log().With("phase", phase))

			gameState, err := s.gameService.GetGameStateQ(ctx, querier)
			if err != nil {
				return "", fmt.Errorf("unable to get game state: %w", err)
			}

			if gameState.Phase != phase {
				return "", domainerrors.NewConflictErrorf(
					"game is in phase %s, expected %s", gameState.Phase, phase,
				)
			}

			resultPhase, err := s.OrchestrateMoveQ(ctx, querier, move, gameState)
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

func (s *OrchestratorImpl[T]) OrchestrateMoveQ(
	ctx ctx.GameContext,
	querier db.Querier,
	move T,
	gameState *state.Game,
) (sqlc.GamePhaseType, error) {
	ctx.Log().Infow("orchestrating move", "move", move)

	if err := s.validationService.ValidateQ(ctx, querier, gameState); err != nil {
		return "", fmt.Errorf("invalid move: %w", err)
	}

	performResult, err := s.service.PerformQ(ctx, querier, move)
	if err != nil {
		return "", fmt.Errorf("unable to perform move: %w", err)
	}

	if err := s.loggingService.LogMoveQ(ctx, querier, move, performResult); err != nil {
		return "", fmt.Errorf("unable to log move: %w", err)
	}

	isMissionAccomplished, err := s.missionService.IsMissionAccomplishedQ(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("unable to check if mission is accomplished: %w", err)
	}

	if isMissionAccomplished {
		ctx.Log().Infow("game is over")

		s.metrics.GamesFinished.Add(ctx, 1)
		s.metrics.ActiveGames.Add(ctx, -1)

		return s.service.PhaseType(), nil
	}

	targetPhase, err := s.service.WalkQ(ctx, querier, false)
	if err != nil {
		return "", fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		ctx.Log().Infow("no need to advance")

		return targetPhase, nil
	}

	ctx.Log().Infow("advancing phase", "target", targetPhase)

	if err := s.service.AdvanceQ(ctx, querier, targetPhase, performResult); err != nil {
		return "", fmt.Errorf("unable to advance move: %w", err)
	}

	ctx.Log().Infow("successfully advanced phase", "target", targetPhase)

	return targetPhase, nil
}
