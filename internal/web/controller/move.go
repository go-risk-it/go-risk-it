package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
)

type MoveController interface {
	PerformDeployMove(ctx ctx.GameContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx ctx.GameContext, attackMove request.AttackMove) error
	PerformConquerMove(ctx ctx.GameContext, conquerMove request.ConquerMove) error
	PerformReinforceMove(ctx ctx.GameContext, reinforceMove request.ReinforceMove) error
	PerformCardsMove(ctx ctx.GameContext, cardsMove request.CardsMove) error
}

type MoveControllerImpl struct {
	deployOrchestrator    orchestration.DeployOrchestrator
	attackOrchestrator    orchestration.AttackOrchestrator
	conquerOrchestrator   orchestration.ConquerOrchestrator
	reinforceOrchestrator orchestration.ReinforceOrchestrator
	cardsOrchestrator     orchestration.CardsOrchestrator
}

var _ MoveController = (*MoveControllerImpl)(nil)

func NewMoveController(
	deployOrchestrator orchestration.DeployOrchestrator,
	attackOrchestrator orchestration.AttackOrchestrator,
	conquerOrchestrator orchestration.ConquerOrchestrator,
	reinforceOrchestrator orchestration.ReinforceOrchestrator,
	cardsOrchestrator orchestration.CardsOrchestrator,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		deployOrchestrator:    deployOrchestrator,
		attackOrchestrator:    attackOrchestrator,
		conquerOrchestrator:   conquerOrchestrator,
		reinforceOrchestrator: reinforceOrchestrator,
		cardsOrchestrator:     cardsOrchestrator,
	}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx ctx.GameContext,
	deployMove request.DeployMove,
) error {
	move := deploy.Move{
		RegionID:      deployMove.RegionID,
		CurrentTroops: deployMove.CurrentTroops,
		DesiredTroops: deployMove.DesiredTroops,
	}

	err := c.deployOrchestrator.OrchestrateMove(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx ctx.GameContext,
	attackMove request.AttackMove,
) error {
	move := attack.Move{
		AttackingRegionID: attackMove.SourceRegionID,
		DefendingRegionID: attackMove.TargetRegionID,
		TroopsInSource:    attackMove.TroopsInSource,
		TroopsInTarget:    attackMove.TroopsInTarget,
		AttackingTroops:   attackMove.AttackingTroops,
	}

	err := c.attackOrchestrator.OrchestrateMove(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformConquerMove(
	ctx ctx.GameContext,
	conquerMove request.ConquerMove,
) error {
	move := conquer.Move{
		Troops: conquerMove.Troops,
	}

	err := c.conquerOrchestrator.OrchestrateMove(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform conquer move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformReinforceMove(
	ctx ctx.GameContext,
	reinforceMove request.ReinforceMove,
) error {
	move := reinforce.Move{
		SourceRegionID: reinforceMove.SourceRegionID,
		TargetRegionID: reinforceMove.TargetRegionID,
		TroopsInSource: reinforceMove.TroopsInSource,
		TroopsInTarget: reinforceMove.TroopsInTarget,
		MovingTroops:   reinforceMove.MovingTroops,
	}

	if err := c.reinforceOrchestrator.OrchestrateMove(ctx, move); err != nil {
		return fmt.Errorf("unable to perform reinforce move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformCardsMove(
	ctx ctx.GameContext,
	cardsMove request.CardsMove,
) error {
	combinations := make([]cards.CardCombination, len(cardsMove.Combinations))
	for i, combination := range cardsMove.Combinations {
		combinations[i] = cards.CardCombination{
			CardIDs: combination.CardIDs,
		}
	}

	move := cards.Move{
		Combinations: combinations,
	}

	if err := c.cardsOrchestrator.OrchestrateMove(ctx, move); err != nil {
		return fmt.Errorf("unable to perform cards move: %w", err)
	}

	return nil
}
