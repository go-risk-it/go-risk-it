package controller

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration"
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
	log                  *zap.SugaredLogger
	attackService        attack.Service
	deployService        deploy.Service
	gameService          game.Service
	orchestrationService orchestration.Service
}

func NewMoveController(
	log *zap.SugaredLogger,
	attackService attack.Service,
	deployService deploy.Service,
	gameService game.Service,
	orchestrationService orchestration.Service,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		log:                  log,
		attackService:        attackService,
		deployService:        deployService,
		gameService:          gameService,
		orchestrationService: orchestrationService,
	}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx context.Context, gameID int64, userID string, deployMove request.DeployMove,
) error {
	err := c.orchestrationService.PerformMove(
		ctx,
		gameID,
		userID,
		c.deployService.ValidatePhase,
		func(ctx context.Context, querier db.Querier, game *sqlc.Game) error {
			err := c.deployService.PerformQ(ctx, querier, game, move.Move[deploy.MoveData]{
				UserID: userID,
				GameID: gameID,
				Payload: deploy.MoveData{
					RegionID:      deployMove.RegionID,
					CurrentTroops: deployMove.CurrentTroops,
					DesiredTroops: deployMove.DesiredTroops,
				},
			})
			if err != nil {
				return fmt.Errorf("unable to perform deploy move: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx context.Context, gameID int64, userID string, attackMove request.AttackMove,
) error {
	err := c.orchestrationService.PerformMove(
		ctx,
		gameID,
		userID,
		c.attackService.ValidatePhase,
		func(ctx context.Context, querier db.Querier, game *sqlc.Game) error {
			err := c.attackService.PerformQ(ctx, querier, game, move.Move[attack.MoveData]{
				UserID: userID,
				GameID: gameID,
				Payload: attack.MoveData{
					SourceRegionID:  attackMove.SourceRegionID,
					TargetRegionID:  attackMove.TargetRegionID,
					TroopsInSource:  attackMove.TroopsInSource,
					TroopsInTarget:  attackMove.TroopsInTarget,
					AttackingTroops: attackMove.AttackingTroops,
				},
			})
			if err != nil {
				return fmt.Errorf("unable to perform attack move: %w", err)
			}

			return nil
		})
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
