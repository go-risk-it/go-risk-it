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

	// InsertPhase advances the turn in the DB when transitioning from REINFORCE.
	// After this call, GetCurrentPlayer returns the next player.
	dbPhase, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create phase: %w", err)
	}

	// Resolve the deploy-for player from the DB. After InsertPhase, the turn
	// has been advanced, so GetCurrentPlayer returns the correct next player.
	currentPlayer, err := s.playerService.GetCurrentPlayer(ctx, querier)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to get current player: %w", err)
	}

	deployableTroops := computeDeployableTroops(ctx, advCtx, currentPlayer.UserID, performResult)

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          dbPhase.ID,
		DeployableTroops: deployableTroops,
	}); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	return moveservice.AdvanceEffect{
		NewPhase: snapshot.DeployPhaseState{DeployableTroops: deployableTroops},
		// TurnEnded is false here: when cards.Advance is called directly
		// (CARDS→DEPLOY), the turn was already incremented at the REINFORCE→CARDS
		// boundary. When called via reinforce.Advance delegation (REINFORCE→DEPLOY),
		// the reinforce advancer overrides TurnEnded to true in its own return.
		TurnEnded: false,
	}, nil
}

// computeDeployableTroops calculates the total deployable troops from pure
// in-memory data: region ownership from AdvanceContext.UpdatedRegions,
// continent bonuses from AdvanceContext.Continents, and card bonuses from
// the perform result.
func computeDeployableTroops(
	ctx ctx.GameContext,
	advCtx moveservice.AdvanceContext,
	deployForUserID string,
	performResult *MoveResult,
) int64 {
	// Count regions owned by the player who will deploy.
	playerRegions := make([]string, 0)
	for _, r := range advCtx.UpdatedRegions {
		if r.OwnerID == deployForUserID {
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
