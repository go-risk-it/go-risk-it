package cards

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
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
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeCARDS, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	phase, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	deployableTroops, err := s.getDeployableTroops(ctx, querier, performResult)
	if err != nil {
		return fmt.Errorf("failed to get deployable troops: %w", err)
	}

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: deployableTroops,
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	slog.InfoContext(ctx, "created deploy phase")

	return nil
}

func (s *service) getDeployableTroops(
	ctx ctx.GameContext,
	querier db.Querier,
	performResult *MoveResult,
) (int64, error) {
	currentPlayer, err := s.playerService.GetCurrentPlayer(ctx, querier)
	if err != nil {
		return -1, fmt.Errorf("failed to get player: %w", err)
	}

	cardReward := int64(0)
	if performResult != nil {
		cardReward = performResult.ExtraDeployableTroops
	}

	playerRegions, err := s.regionService.GetRegionsControlledByPlayer(
		ctx,
		querier,
		currentPlayer.ID,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to get regions: %w", err)
	}

	regionReward := max(int64(len(playerRegions)/regionDivisor), minRegionReward)

	continentReward, err := s.getContinentReward(ctx, querier, currentPlayer)
	if err != nil {
		return -1, fmt.Errorf("failed to get continent reward: %w", err)
	}

	slog.DebugContext(ctx, "awarding deployable troops",
		"region",
		regionReward,
		"continent",
		continentReward,
		"card",
		cardReward)

	return regionReward + continentReward + cardReward, nil
}

func (s *service) getContinentReward(
	ctx ctx.GameContext,
	querier db.Querier,
	currentPlayer sqlc.GamePlayer,
) (int64, error) {
	continents, err := s.boardService.GetContinentsControlledByPlayer(
		ctx,
		querier,
		currentPlayer.ID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get continents: %w", err)
	}

	continentReward := int64(0)
	for _, continent := range continents {
		continentReward += int64(continent.BonusTroops)
	}

	return continentReward, nil
}
