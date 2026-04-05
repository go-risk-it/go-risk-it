package cards

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// regionDivisor is the divisor applied to a player's region count to compute the base reward.
	regionDivisor = 3
	// minRegionReward is the minimum troop reward from region ownership.
	minRegionReward = 3
)

func (s *service) Advance(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult *MoveResult,
	advCtx moveservice.AdvanceContext,
) (moveservice.AdvanceEffect, error) {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeCARDS, targetPhase); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("invalid phase transition: %w", err)
	}

	dbPhase, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create phase: %w", err)
	}

	deployableTroops := computeDeployableTroops(ctx, advCtx, performResult)

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          dbPhase.ID,
		DeployableTroops: deployableTroops,
	}); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	return moveservice.AdvanceEffect{
		NewPhase:  snapshot.DeployPhaseState{DeployableTroops: deployableTroops},
		TurnEnded: true,
	}, nil
}

// computeDeployableTroops calculates the total deployable troops from pure
// in-memory data: region ownership from AdvanceContext.UpdatedRegions,
// continent bonuses from AdvanceContext.Continents, and card bonuses from
// the perform result. Zero DB reads.
func computeDeployableTroops(
	ctx ctx.GameContext,
	advCtx moveservice.AdvanceContext,
	performResult *MoveResult,
) int64 {
	// Count regions owned by the current player.
	playerRegions := make([]string, 0)
	for _, r := range advCtx.UpdatedRegions {
		if r.OwnerID == advCtx.CurrentUserID {
			playerRegions = append(playerRegions, r.ID)
		}
	}

	regionReward := max(int64(len(playerRegions)/regionDivisor), minRegionReward)

	// Compute continent bonus from fully-controlled continents.
	controlled := advCtx.Continents.GetContinentsControlledBy(playerRegions)

	continentReward := int64(0)
	for _, continent := range controlled {
		continentReward += int64(continent.BonusTroops)
	}

	// Card bonus from the perform result (extra troops from card combinations).
	cardReward := int64(0)
	if performResult != nil {
		cardReward = performResult.ExtraDeployableTroops
	}

	observe.SpanEvent(ctx, "awarding_deployable_troops",
		attribute.Int64("region_reward", regionReward),
		attribute.Int64("continent_reward", continentReward),
		attribute.Int64("card_reward", cardReward),
	)

	return regionReward + continentReward + cardReward
}
