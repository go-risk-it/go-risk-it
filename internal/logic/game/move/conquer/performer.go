package conquer

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *ServiceImpl) PerformQ(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	ctx.Log().Infow("performing conquer move", "move", move)

	phaseState, err := s.GetPhaseStateQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("unable to get phase state: %w", err)
	}

	if phaseState.MinimumTroops > move.Troops {
		return nil, fmt.Errorf("must move at least %d troops", phaseState.MinimumTroops)
	}

	sourceRegion, err := s.regionService.GetRegionQ(ctx, querier, phaseState.SourceRegion)
	if err != nil {
		return nil, fmt.Errorf("unable to get attacking region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegionQ(ctx, querier, phaseState.TargetRegion)
	if err != nil {
		return nil, fmt.Errorf("unable to get defending region: %w", err)
	}

	if sourceRegion.Troops-move.Troops < 1 {
		return nil, errors.New("source region does not have enough troops")
	}

	defeatedPlayerID, err := s.updateRegionTroops(ctx, querier, move, sourceRegion, targetRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to update region troops: %w", err)
	}

	isDefenderEliminated, err := s.isDefenderEliminated(ctx, querier, defeatedPlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if defender is eliminated: %w", err)
	}

	if isDefenderEliminated {
		if err := s.handlePlayerEliminated(
			ctx,
			querier,
			defeatedPlayerID,
		); err != nil {
			return nil, fmt.Errorf("unable to handle player eliminated: %w", err)
		}
	}

	ctx.Log().Infow("conquer executed successfully")

	return &MoveResult{}, nil
}

func (s *ServiceImpl) updateRegionTroops(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
) (int64, error) {
	if err := s.regionService.UpdateTroopsInRegionQ(
		ctx,
		querier,
		sourceRegion,
		-move.Troops,
	); err != nil {
		return 0, fmt.Errorf("failed to decrease troops in source region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegionQ(
		ctx,
		querier,
		targetRegion,
		move.Troops,
	); err != nil {
		return 0, fmt.Errorf("failed to increase troops in target region: %w", err)
	}

	ctx.Log().Infow("troops updated successfully")

	defeatedPlayerID, err := s.regionService.UpdateRegionOwnerQ(
		ctx,
		querier,
		targetRegion,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update region owner: %w", err)
	}

	return defeatedPlayerID, nil
}

func (s *ServiceImpl) isDefenderEliminated(
	ctx ctx.GameContext,
	querier db.Querier,
	defeatedPlayerID int64,
) (bool, error) {
	defeatedPlayerRegions, err := s.regionService.GetRegionsControlledByPlayerQ(
		ctx,
		querier,
		defeatedPlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get regions controlled by player: %w", err)
	}

	return len(defeatedPlayerRegions) == 0, nil
}

func (s *ServiceImpl) handlePlayerEliminated(
	ctx ctx.GameContext,
	querier db.Querier,
	eliminatedPlayerID int64,
) error {
	ctx.Log().Infow("defending player has been eliminated", "defender", eliminatedPlayerID)

	if err := s.cardService.TransferCardsOwnershipQ(ctx, querier, eliminatedPlayerID); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	if err := s.missionService.ReassignMissionsQ(
		ctx,
		querier,
		eliminatedPlayerID,
	); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	return nil
}
