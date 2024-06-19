package controller

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type MoveController interface {
	PerformDeployMove(
		ctx context.Context, gameID int64, userID string, deployMove request.DeployMove,
	) error
	PerformAttackMove(
		ctx context.Context, gameID int64, userID string, attackMove request.AttackMove,
	) error
}

type MoveControllerImpl struct {
	log                      *zap.SugaredLogger
	querier                  db.Querier
	attackService            attack.Service
	deployService            deploy.Service
	orchestrationService     orchestration.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

func NewMoveController(
	log *zap.SugaredLogger,
	querier db.Querier,
	attackService attack.Service,
	deployService deploy.Service,
	orchestrationService orchestration.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		log:                      log,
		querier:                  querier,
		attackService:            attackService,
		deployService:            deployService,
		orchestrationService:     orchestrationService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func performMove[T any](
	ctx context.Context,
	querier db.Querier,
	move move.Move[T],
	performer move.Performer[T],
	advancePhase func(context.Context, db.Querier, int64) error,
	publish func(context.Context, int64),
) error {
	_, err := querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(querier db.Querier) (interface{}, error) {
			// validate move
			if err := performer.PerformQ(ctx, querier, move); err != nil {
				return nil, fmt.Errorf("unable to perform move: %w", err)
			}

			err := advancePhase(ctx, querier, move.GameID)
			if err != nil {
				return nil, fmt.Errorf("unable to advancePhase game: %w", err)
			}

			return struct{}{}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	publish(ctx, move.GameID)

	return nil
}

func (c *MoveControllerImpl) publishMoveResult(ctx context.Context, gameID int64) {
	go c.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: gameID,
	})
	go c.playerStateChangedSignal.Emit(ctx, signals.PlayerStateChangedData{
		GameID: gameID,
	})
	go c.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		GameID: gameID,
	})
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx context.Context, gameID int64, userID string, deployMove request.DeployMove,
) error {
	err := performMove(ctx, c.querier, move.Move[deploy.MoveData]{
		UserID: userID,
		GameID: gameID,
		Payload: deploy.MoveData{
			RegionID:      deployMove.RegionID,
			CurrentTroops: deployMove.CurrentTroops,
			DesiredTroops: deployMove.DesiredTroops,
		},
	}, c.deployService, c.orchestrationService.OrchestrateQ, c.publishMoveResult)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx context.Context, gameID int64, userID string, attackMove request.AttackMove,
) error {
	err := performMove(ctx, c.querier, move.Move[attack.MoveData]{
		UserID: userID,
		GameID: gameID,
		Payload: attack.MoveData{
			SourceRegionID:  attackMove.SourceRegionID,
			TargetRegionID:  attackMove.TargetRegionID,
			TroopsInSource:  attackMove.TroopsInSource,
			TroopsInTarget:  attackMove.TroopsInTarget,
			AttackingTroops: attackMove.AttackingTroops,
		},
	}, c.attackService, c.orchestrationService.OrchestrateQ, c.publishMoveResult)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
