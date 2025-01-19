package attack

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
	performResult *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeCONQUER && targetPhase != sqlc.PhaseTypeREINFORCE {
		return fmt.Errorf("cannot advance attack phase to %s", targetPhase)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.PhaseTypeCONQUER {
		defendingPlayerRegions, err := querier.GetPlayerRegionsFromRegion(
			ctx,
			sqlc.GetPlayerRegionsFromRegionParams{
				GameID:            ctx.GameID(),
				ExternalReference: performResult.DefendingRegionID,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to get player regions: %w", err)
		}

		if defendingPlayerRegions.RegionCount == 1 {
			if err := s.handlePlayerEliminated(
				ctx,
				querier,
				defendingPlayerRegions.UserID); err != nil {
				return fmt.Errorf("unable to handle player eliminated: %w", err)
			}
		}

		return s.advanceToConquerPhase(ctx, querier, performResult, *phase)
	}

	return nil
}

func (s *ServiceImpl) handlePlayerEliminated(
	ctx ctx.GameContext,
	querier db.Querier,
	eliminatedUserID string,
) error {
	ctx.Log().Infow("defending player has been eliminated", "defender", eliminatedUserID)

	if err := s.cardService.TransferCardsOwnershipQ(
		ctx,
		querier,
		ctx.UserID(),
		eliminatedUserID,
	); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	if err := s.missionService.ReassignMissionsQ(
		ctx,
		querier,
		eliminatedUserID,
	); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	return nil
}

func (s *ServiceImpl) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	performResult *MoveResult,
	phase sqlc.Phase,
) error {
	if _, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             phase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   performResult.AttackingRegionID,
		ExternalReference_2: performResult.DefendingRegionID,
		MinimumTroops:       performResult.ConqueringTroops,
	}); err != nil {
		return fmt.Errorf("failed to create conquer phase: %w", err)
	}

	ctx.Log().Debugw("created conquer phase")

	return nil
}
