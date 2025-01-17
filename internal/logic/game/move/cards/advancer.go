package cards

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.PhaseType,
	moveResult *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeDEPLOY {
		return fmt.Errorf("cannot advance cards phase to %s", targetPhase)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	cardReward := int64(0)

	if moveResult != nil {
		ctx.Log().
			Infow("deployable troops from cards", "amount", moveResult.ExtraDeployableTroops)

		cardReward = moveResult.ExtraDeployableTroops
	}

	playerRegions, err := s.regionService.GetPlayerRegionsQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get regions: %w", err)
	}

	regionReward := int64(len(playerRegions) / 3)

	continents, err := s.boardService.GetContinentsControlledByPlayerQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get continents: %w", err)
	}

	continentReward := int64(0)
	for _, continent := range continents {
		continentReward += int64(continent.BonusTroops)
	}

	ctx.Log().Debugw("awarding deployable troops",
		"region",
		regionReward,
		"continent",
		continentReward,
		"card",
		cardReward)

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: regionReward + continentReward + cardReward,
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	ctx.Log().Infow("created deploy phase")

	return nil
}
