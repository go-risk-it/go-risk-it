package orchestration

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/jackc/pgx/v5"
)

type Orchestrator[T, R any] interface {
	OrchestrateMove(ctx ctx.GameContext, move T) error
}

type OrchestratorImpl[T, R any] struct {
	querier                db.Querier
	service                service.Service[T, R]
	gameService            state.Service
	loggingService         logging.Service
	missionService         mission.Service
	validationService      validation.Service
	gameStateChangedSignal signals.GameStateChangedSignal
}

func NewOrchestrator[T, R any](
	querier db.Querier,
	service service.Service[T, R],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *OrchestratorImpl[T, R] {
	return &OrchestratorImpl[T, R]{
		querier:                querier,
		service:                service,
		gameService:            gameService,
		loggingService:         loggingService,
		missionService:         missionService,
		validationService:      validationService,
		gameStateChangedSignal: gameStateChangedSignal,
	}
}

func (s *OrchestratorImpl[T, R]) OrchestrateMove(ctx ctx.GameContext, move T) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(querier db.Querier) (interface{}, error) {
			phase := s.service.PhaseType()
			ctx.SetLog(ctx.Log().With("phase", phase))

			gameState, err := s.gameService.GetGameStateQ(ctx, querier)
			if err != nil {
				return struct{}{}, fmt.Errorf("unable to get game state: %w", err)
			}

			if gameState.Phase != phase {
				return struct{}{}, errors.New("game is not in the correct phase to perform move")
			}

			if err := s.OrchestrateMoveQ(ctx, querier, move, gameState); err != nil {
				return struct{}{}, fmt.Errorf("unable to orchestrate move: %w", err)
			}

			return struct{}{}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	go s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{})

	return nil
}

func (s *OrchestratorImpl[T, R]) OrchestrateMoveQ(
	ctx ctx.GameContext,
	querier db.Querier,
	move T,
	gameState *state.Game,
) error {
	ctx.Log().Infow("orchestrating move", "move", move)

	if err := s.validationService.ValidateQ(ctx, querier, gameState); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	performResult, err := s.service.PerformQ(ctx, querier, move)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	if err := s.loggingService.LogMoveQ(ctx, querier, move, performResult); err != nil {
		return fmt.Errorf("unable to log move: %w", err)
	}

	isMissionAccomplished, err := s.missionService.IsMissionAccomplishedQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to check if mission is accomplished: %w", err)
	}

	if isMissionAccomplished {
		ctx.Log().Infow("game is over")

		return nil
	}

	targetPhase, err := s.service.WalkQ(ctx, querier, false)
	if err != nil {
		return fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		ctx.Log().Infow("no need to advance")

		return nil
	}

	ctx.Log().Infow("advancing phase", "target", targetPhase)

	if err := s.service.AdvanceQ(ctx, querier, targetPhase, performResult); err != nil {
		return fmt.Errorf("unable to advance move: %w", err)
	}

	ctx.Log().Infow("successfully advanced phase", "target", targetPhase)

	return nil
}
