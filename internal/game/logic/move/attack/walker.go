package attack

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
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
		observe.SpanEvent(ctx, "must_advance_phase_to_conquer")

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
		observe.SpanEvent(ctx, "must_advance_phase_to_reinforce")

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
