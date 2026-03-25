package attack

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *service) Walk(
	ctx ctx.GameContext,
	querier db.Querier,
	voluntaryAdvancement bool,
) (sqlc.GamePhaseType, error) {
	hasConquered, err := s.HasConquered(ctx, querier)
	if err != nil {
		return sqlc.GamePhaseTypeATTACK, fmt.Errorf(
			"failed to check if attack has conquered: %w",
			err,
		)
	}

	if hasConquered {
		slog.InfoContext(ctx, "must advance phase to CONQUER")

		return sqlc.GamePhaseTypeCONQUER, nil
	}

	canContinueAttacking, err := s.CanContinueAttacking(ctx, querier)
	if err != nil {
		return sqlc.GamePhaseTypeATTACK, fmt.Errorf(
			"failed to check if attack can continue: %w",
			err,
		)
	}

	if voluntaryAdvancement || !canContinueAttacking {
		slog.InfoContext(ctx, "must advance phase to REINFORCE")

		return sqlc.GamePhaseTypeREINFORCE, nil
	}

	return sqlc.GamePhaseTypeATTACK, nil
}

// HasConquered returns true if the player has conquered any region.
// This is detected by checking that there is exactly one region
// (non owned by the player) that has 0 troops.
func (s *service) HasConquered(ctx ctx.GameContext, querier db.Querier) (bool, error) {
	regions, err := s.regionService.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	slog.InfoContext(ctx, "checking if player has conquered any region", "regions", len(regions))

	for _, region := range regions {
		if region.UserID != ctx.UserID() && region.Troops == 0 {
			slog.InfoContext(ctx, "player has conquered a region",
				"region", region.ExternalReference)

			return true, nil
		}
	}

	slog.InfoContext(ctx, "player has not conquered any region")

	return false, nil
}

// CanContinueAttacking returns true if the player does not have any attack move available.
// This is detected by checking that all of the regions owned by the player have exactly 1 troop.
func (s *service) CanContinueAttacking(
	ctx ctx.GameContext,
	querier db.Querier,
) (bool, error) {
	regions, err := s.regionService.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	slog.InfoContext(ctx, "checking if player can continue attacking", "regions", len(regions))

	for _, region := range regions {
		if region.UserID == ctx.UserID() && region.Troops > 1 {
			slog.InfoContext(ctx, "player can continue attacking")

			return true, nil
		}
	}

	slog.InfoContext(ctx, "player can not continue attacking")

	return false, nil
}
