package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/jackc/pgx/v5"
)

type Orchestator[T, R any] interface {
	OrchestrateMove(
		ctx ctx.MoveContext,
		phase sqlc.PhaseType,
		move T,
	) error
}

type OrchestatorImpl[T, R any] struct {
	querier                  db.Querier
	service                  service.Service[T, R]
	gameService              state.Service
	validationService        validation.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

func NewOrchestrator[T, R any](
	querier db.Querier,
	service service.Service[T, R],
	gameService state.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *OrchestatorImpl[T, R] {
	return &OrchestatorImpl[T, R]{
		querier:                  querier,
		service:                  service,
		gameService:              gameService,
		validationService:        validationService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func (s *OrchestatorImpl[T, R]) OrchestrateMove(
	ctx ctx.MoveContext,
	phase sqlc.PhaseType,
	move T,
) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(q db.Querier) (interface{}, error) {
			err := s.OrchestrateMoveQ(ctx, q, phase, move)
			if err != nil {
				return struct{}{}, err
			}

			return struct{}{}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.publishMoveResult(ctx)

	return nil
}

func (s *OrchestatorImpl[T, R]) OrchestrateMoveQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	phase sqlc.PhaseType,
	move T,
) error {
	ctx.Log().Infow("orchestrating move", "phase", phase)

	gameState, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to get game state: %w", err)
	}

	if gameState.Phase != phase {
		return fmt.Errorf("game is not in the correct phase to perform move")
	}

	if err := s.validationService.Validate(ctx, querier, gameState); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	performResult, err := s.service.PerformQ(ctx, querier, move)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	targetPhase, err := s.service.Walk(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == phase {
		ctx.Log().Infow("no need to advance")

		return nil
	}

	if err := s.service.AdvanceQ(ctx, querier, targetPhase, performResult); err != nil {
		return fmt.Errorf("unable to advance move: %w", err)
	}

	return nil
}

func (s *OrchestatorImpl[T, R]) publishMoveResult(ctx ctx.GameContext) {
	go s.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: ctx.GameID(),
	})
	go s.playerStateChangedSignal.Emit(ctx, signals.PlayerStateChangedData{
		GameID: ctx.GameID(),
	})
	go s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		GameID: ctx.GameID(),
	})
}
