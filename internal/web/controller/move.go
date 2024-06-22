package controller

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/validation"
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
	gameService              game.Service
	orchestrationService     orchestration.Service
	validationService        validation.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

func NewMoveController(
	log *zap.SugaredLogger,
	querier db.Querier,
	attackService attack.Service,
	deployService deploy.Service,
	gameService game.Service,
	orchestrationService orchestration.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		log:                      log,
		querier:                  querier,
		attackService:            attackService,
		deployService:            deployService,
		gameService:              gameService,
		validationService:        validationService,
		orchestrationService:     orchestrationService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func performMove[T any](
	controller *MoveControllerImpl,
	ctx context.Context,
	querier db.Querier,
	move move.Move[T],
	performer move.Performer[T],
) error {
	_, err := querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(querier db.Querier) (interface{}, error) {
			gameState, err := controller.gameService.GetGameStateQ(ctx, querier, move.GameID)
			if err != nil {
				return nil, fmt.Errorf("unable to get game state: %w", err)
			}

			if !performer.ValidatePhase(gameState) {
				return nil, fmt.Errorf("game is not in the correct phase to perform move")
			}

			err = controller.validationService.Validate(ctx, querier, gameState, move.UserID)
			if err != nil {
				return nil, fmt.Errorf("invalid move: %w", err)
			}

			// validate move
			if err := performer.PerformQ(ctx, querier, move, gameState); err != nil {
				return nil, fmt.Errorf("unable to perform move: %w", err)
			}

			err = controller.orchestrationService.AdvancePhaseQ(ctx, querier, move.GameID)
			if err != nil {
				return nil, fmt.Errorf("unable to advancePhase game: %w", err)
			}

			return struct{}{}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	controller.publishMoveResult(ctx, move.GameID)

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
	err := performMove(c, ctx, c.querier, move.Move[deploy.MoveData]{
		UserID: userID,
		GameID: gameID,
		Payload: deploy.MoveData{
			RegionID:      deployMove.RegionID,
			CurrentTroops: deployMove.CurrentTroops,
			DesiredTroops: deployMove.DesiredTroops,
		},
	}, c.deployService)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx context.Context, gameID int64, userID string, attackMove request.AttackMove,
) error {
	err := performMove(c, ctx, c.querier, move.Move[attack.MoveData]{
		UserID: userID,
		GameID: gameID,
		Payload: attack.MoveData{
			SourceRegionID:  attackMove.SourceRegionID,
			TargetRegionID:  attackMove.TargetRegionID,
			TroopsInSource:  attackMove.TroopsInSource,
			TroopsInTarget:  attackMove.TroopsInTarget,
			AttackingTroops: attackMove.AttackingTroops,
		},
	}, c.attackService)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
