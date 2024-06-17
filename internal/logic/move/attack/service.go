package attack

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type MoveData struct {
	SourceRegionID  string
	TargetRegionID  string
	TroopsInSource  int64
	TroopsInTarget  int64
	AttackingTroops int64
}

type Service interface {
	Perform(
		ctx context.Context,
		move move.Move[MoveData],
	) error
	PerformAttackMoveQ(
		ctx context.Context,
		querier db.Querier,
		move move.Move[MoveData],
	) error
}

type ServiceImpl struct {
	log                     *zap.SugaredLogger
	querier                 db.Querier
	gameService             game.Service
	playerService           game.Service
	regionService           region.Service
	validationService       validation.Service
	boardStateChangedSignal signals.BoardStateChangedSignal
	gameStateChangedSignal  signals.GameStateChangedSignal
}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService game.Service,
	playerService game.Service,
	regionService region.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                 que,
		log:                     log,
		gameService:             gameService,
		playerService:           playerService,
		regionService:           regionService,
		validationService:       validationService,
		boardStateChangedSignal: boardStateChangedSignal,
		gameStateChangedSignal:  gameStateChangedSignal,
	}
}

func (s *ServiceImpl) Perform(
	ctx context.Context,
	move move.Move[MoveData],
) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(qtx db.Querier) (interface{}, error) {
			return nil, s.PerformAttackMoveQ(
				ctx,
				qtx,
				move,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to perform deploy move: %w", err)
	}

	go s.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: move.GameID,
	})
	go s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		GameID: move.GameID,
	})

	return nil
}

func (s *ServiceImpl) PerformAttackMoveQ(
	ctx context.Context,
	querier db.Querier,
	move move.Move[MoveData],
) error {
	s.log.Infow(
		"performing attack move",
		"move", move,
	)

	return nil
}
