package attack

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

const (
	// conqueredRegionTroops is the troop count that indicates a region has been conquered.
	conqueredRegionTroops = 0
	// minTroopsToLaunchAttack is the minimum troops a region must exceed to launch an attack.
	// A region must retain at least one troop, so it needs more than this value.
	minTroopsToLaunchAttack = 1
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
		slog.DebugContext(ctx, "must advance phase to CONQUER")

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
		slog.DebugContext(ctx, "must advance phase to REINFORCE")

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

	slog.DebugContext(ctx, "checking if player has conquered any region", "regions", len(regions))

	for _, region := range regions {
		if region.UserID != ctx.UserID() && region.Troops == conqueredRegionTroops {
			slog.DebugContext(ctx, "player has conquered a region",
				"region", region.ExternalReference)

			return true, nil
		}
	}

	slog.DebugContext(ctx, "player has not conquered any region")

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

	slog.DebugContext(ctx, "checking if player can continue attacking", "regions", len(regions))

	for _, region := range regions {
		if region.UserID == ctx.UserID() && region.Troops > minTroopsToLaunchAttack {
			slog.DebugContext(ctx, "player can continue attacking")

			return true, nil
		}
	}

	slog.DebugContext(ctx, "player can not continue attacking")

	return false, nil
}
