package orchestration

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/validation"
	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Service interface {
	PerformMove(
		ctx context.Context,
		gameID int64,
		userID string,
		phase sqlc.Phase,
		perform func(ctx context.Context, querier db.Querier, game *sqlc.Game) error,
	) error
	PerformMoveQ(
		ctx context.Context,
		querier db.Querier,
		gameID int64,
		phase sqlc.Phase,
		userID string,
		perform func(ctx context.Context, querier db.Querier, game *sqlc.Game) error,
	) (interface{}, error)
}
type ServiceImpl struct {
	log                      *zap.SugaredLogger
	querier                  db.Querier
	gameService              game.Service
	phaseService             phase.Service
	validationService        validation.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

func NewService(
	log *zap.SugaredLogger,
	querier db.Querier,
	phaseWalkingService phase.Service,
	gameService game.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		log:                      log,
		querier:                  querier,
		phaseService:             phaseWalkingService,
		gameService:              gameService,
		validationService:        validationService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func (s *ServiceImpl) PerformMove(
	ctx context.Context,
	gameID int64,
	userID string,
	phase sqlc.Phase,
	perform func(ctx context.Context, querier db.Querier, game *sqlc.Game) error,
) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(querier db.Querier) (interface{}, error) {
			if err := s.PerformMoveQ(ctx, querier, gameID, phase, userID, perform); err != nil {
				return nil, fmt.Errorf("unable to perform move: %w", err)
			}

			return struct{}{}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.publishMoveResult(ctx, gameID)

	return nil
}

func (s *ServiceImpl) PerformMoveQ(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	phase sqlc.Phase,
	userID string,
	perform func(ctx context.Context, querier db.Querier, game *sqlc.Game) error,
) error {
	gameState, err := s.gameService.GetGameStateQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("unable to get game state: %w", err)
	}

	if gameState.Phase != phase {
		return fmt.Errorf("game is not in the correct phase to perform move")
	}

	if err := s.validationService.Validate(
		ctx,
		querier,
		gameState,
		userID); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	if err := perform(ctx, querier, gameState); err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	if err := s.phaseService.AdvanceQ(
		ctx,
		querier,
		gameID); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	return nil
}

func (s *ServiceImpl) publishMoveResult(ctx context.Context, gameID int64) {
	go s.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: gameID,
	})
	go s.playerStateChangedSignal.Emit(ctx, signals.PlayerStateChangedData{
		GameID: gameID,
	})
	go s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		GameID: gameID,
	})
}
