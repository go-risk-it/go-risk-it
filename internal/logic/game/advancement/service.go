package advancement

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/jackc/pgx/v5"
)

type Service[T, R any] interface {
	Advance(ctx ctx.GameContext) error
	AdvanceQ(ctx ctx.GameContext, querier db.Querier) error
}

type ServiceImpl[T, R any] struct {
	querier                db.Querier
	gameState              state.Service
	moveService            service.Service[T, R]
	validationService      validation.Service
	gameStateChangedSignal signals.GameStateChangedSignal
}

func NewService[T, R any](
	gameState state.Service,
	querier db.Querier,
	moveService service.Service[T, R],
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl[T, R] {
	return &ServiceImpl[T, R]{
		gameState:              gameState,
		querier:                querier,
		moveService:            moveService,
		validationService:      validationService,
		gameStateChangedSignal: gameStateChangedSignal,
	}
}

func (s *ServiceImpl[T, R]) Advance(ctx ctx.GameContext) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(q db.Querier) (interface{}, error) {
			err := s.AdvanceQ(ctx, q)
			if err != nil {
				return struct{}{}, err
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

func (s *ServiceImpl[T, R]) AdvanceQ(ctx ctx.GameContext, querier db.Querier) error {
	currentPhase := s.moveService.PhaseType()

	ctx.Log().Infow("processing request to advance phase", "currentPhase", currentPhase)

	game, err := s.gameState.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to get game state: %w", err)
	}

	if err := s.validationService.ValidateQ(ctx, querier, game); err != nil {
		ctx.Log().Errorw("validation failed", "error", err)

		return fmt.Errorf("validation failed: %w", err)
	}

	ctx.Log().Infof("game is in phase %s", game.Phase)

	if game.Phase != currentPhase {
		return fmt.Errorf("game is not in phase %s", currentPhase)
	}

	var performResult R

	targetPhase, err := s.moveService.WalkQ(ctx, querier, true)
	if err != nil {
		return fmt.Errorf("unable to walk to target phase: %w", err)
	}

	err = s.moveService.AdvanceQ(
		ctx,
		querier,
		targetPhase,
		performResult,
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	ctx.Log().Infow("phase advanced successfully", "from", currentPhase)

	return nil
}
