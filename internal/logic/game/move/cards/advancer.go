package cards

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	moveResult *MoveResult,
) error {
	if targetPhase != sqlc.GamePhaseTypeDEPLOY {
		return fmt.Errorf("cannot advance cards phase to %s", targetPhase)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	deployableTroops, err := s.getDeployableTroops(ctx, querier, moveResult)
	if err != nil {
		return fmt.Errorf("failed to get deployable troops: %w", err)
	}

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: deployableTroops,
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	ctx.Log().Infow("created deploy phase")

	return nil
}

func (s *ServiceImpl) getDeployableTroops(
	ctx ctx.GameContext,
	querier db.Querier,
	moveResult *MoveResult,
) (int64, error) {
	currentPlayer, err := s.playerService.GetCurrentPlayerQ(ctx, querier)
	if err != nil {
		return -1, fmt.Errorf("failed to get player: %w", err)
	}

	cardReward := int64(0)
	if moveResult != nil {
		cardReward = moveResult.ExtraDeployableTroops
	}

	playerRegions, err := s.regionService.GetRegionsControlledByPlayerQ(
		ctx,
		querier,
		currentPlayer.ID,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to get regions: %w", err)
	}

	regionReward := int64(len(playerRegions) / 3)

	continentReward, err := s.getContinentReward(ctx, querier, currentPlayer)
	if err != nil {
		return -1, fmt.Errorf("failed to get continent reward: %w", err)
	}

	ctx.Log().Debugw("awarding deployable troops",
		"region",
		regionReward,
		"continent",
		continentReward,
		"card",
		cardReward)

	return regionReward + continentReward + cardReward, nil
}

func (s *ServiceImpl) getContinentReward(
	ctx ctx.GameContext,
	querier db.Querier,
	currentPlayer sqlc.GamePlayer,
) (int64, error) {
	continents, err := s.boardService.GetContinentsControlledByPlayerQ(
		ctx,
		querier,
		currentPlayer.ID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get continents: %w", err)
	}

	continentReward := int64(0)
	for _, continent := range continents {
		continentReward += int64(continent.BonusTroops)
	}

	return continentReward, nil
}
