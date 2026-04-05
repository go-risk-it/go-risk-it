package attack

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// HasConquered returns true if the player has conquered any region.
// This is detected by checking that there is exactly one region
// (non owned by the player) that has 0 troops.
func (s *service) HasConquered(ctx ctx.GameContext, querier db.Querier) (bool, error) {
	regions, err := s.regionService.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	for _, region := range regions {
		if region.UserID != ctx.UserID() && region.Troops == conqueredRegionTroops {
			observe.SpanEvent(ctx, "player_has_conquered_a_region",
				attribute.String("region", region.ExternalReference),
			)

			return true, nil
		}
	}

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

	for _, region := range regions {
		if region.UserID == ctx.UserID() && region.Troops > minTroopsToLaunchAttack {
			return true, nil
		}
	}

	return false, nil
}
