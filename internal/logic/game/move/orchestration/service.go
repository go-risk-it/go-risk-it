package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	OrchestrateMove(
		ctx ctx.MoveContext,
		phase sqlc.PhaseType,
		perform func(ctx.MoveContext, db.Querier) error,
		walk func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error),
		advance func(ctx.MoveContext, db.Querier, sqlc.PhaseType) error,
	) error
	OrchestrateMoveQ(
		ctx ctx.MoveContext,
		querier db.Querier,
		phase sqlc.PhaseType,
		perform func(ctx.MoveContext, db.Querier) error,
		walk func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error),
		advance func(ctx.MoveContext, db.Querier, sqlc.PhaseType) error,
	) error
}
type ServiceImpl struct {
	querier                  db.Querier
	gameService              state.Service
	validationService        validation.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	gameService state.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                  querier,
		gameService:              gameService,
		validationService:        validationService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func (s *ServiceImpl) OrchestrateMove(
	ctx ctx.MoveContext,
	phase sqlc.PhaseType,
	perform func(ctx.MoveContext, db.Querier) error,
	walk func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error),
	advance func(ctx.MoveContext, db.Querier, sqlc.PhaseType) error,
) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(q db.Querier) (interface{}, error) {
			if err := s.OrchestrateMoveQ(ctx, q, phase, perform, walk, advance); err != nil {
				return nil, fmt.Errorf("unable to perform move: %w", err)
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

func (s *ServiceImpl) OrchestrateMoveQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	phase sqlc.PhaseType,
	perform func(ctx.MoveContext, db.Querier) error,
	walk func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error),
	advance func(ctx.MoveContext, db.Querier, sqlc.PhaseType) error,
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

	if err := perform(ctx, querier); err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	targetPhase, err := walk(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == phase {
		ctx.Log().Infow("no need to advance")

		return nil
	}

	if err := advance(ctx, querier, targetPhase); err != nil {
		return fmt.Errorf("unable to advance move: %w", err)
	}

	return nil
}

func (s *ServiceImpl) publishMoveResult(ctx ctx.GameContext) {
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
