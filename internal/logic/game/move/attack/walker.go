package attack

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) Walk(
	ctx ctx.GameContext,
	querier db.Querier,
	voluntaryAdvancement bool,
) (sqlc.PhaseType, error) {
	hasConquered, err := s.HasConqueredQ(ctx, querier)
	if err != nil {
		return sqlc.PhaseTypeATTACK, fmt.Errorf("failed to check if attack has conquered: %w", err)
	}

	if hasConquered {
		ctx.Log().Infow("must advance phase to CONQUER")

		return sqlc.PhaseTypeCONQUER, nil
	}

	canContinueAttacking, err := s.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return sqlc.PhaseTypeATTACK, fmt.Errorf("failed to check if attack can continue: %w", err)
	}

	if voluntaryAdvancement || !canContinueAttacking {
		ctx.Log().Infow("must advance phase to REINFORCE")

		return sqlc.PhaseTypeREINFORCE, nil
	}

	return sqlc.PhaseTypeATTACK, nil
}

// HasConqueredQ returns true if the player has conquered any region.
// This is detected by checking that there is exactly one region
// (non owned by the player) that has 0 troops.
func (s *ServiceImpl) HasConqueredQ(ctx ctx.GameContext, querier db.Querier) (bool, error) {
	regions, err := s.regionService.GetRegionsQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	ctx.Log().Infow("checking if player has conquered any region", "regions", len(regions))

	for _, region := range regions {
		if region.UserID != ctx.UserID() && region.Troops == 0 {
			ctx.Log().Infow("player has conquered a region", "region", region.ExternalReference)

			return true, nil
		}
	}

	ctx.Log().Infow("player has not conquered any region")

	return false, nil
}

// CanContinueAttackingQ returns true if the player does not have any attack move available.
// This is detected by checking that all of the regions owned by the player have exactly 1 troop.
func (s *ServiceImpl) CanContinueAttackingQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (bool, error) {
	regions, err := s.regionService.GetRegionsQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	ctx.Log().Infow("checking if player can continue attacking", "regions", len(regions))

	for _, region := range regions {
		if region.UserID == ctx.UserID() && region.Troops > 1 {
			ctx.Log().Infow("player can continue attacking")

			return true, nil
		}
	}

	ctx.Log().Infow("player can not continue attacking")

	return false, nil
}
