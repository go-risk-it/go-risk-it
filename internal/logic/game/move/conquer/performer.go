package conquer

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
)

const (
	// minTroopsToRetain is the minimum troops that must remain in the source
	// region after conquering.
	minTroopsToRetain = 1
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (struct{}, error) {
	slog.DebugContext(ctx, "performing conquer move", "move", move)

	phaseState, err := s.GetPhaseStateWithQuerier(ctx, querier)
	if err != nil {
		return struct{}{}, fmt.Errorf("unable to get phase state: %w", err)
	}

	if phaseState.MinimumTroops > move.Troops {
		return struct{}{}, domainerrors.NewValidationErrorf(
			"must move at least %d troops",
			phaseState.MinimumTroops,
		)
	}

	sourceRegion, err := s.regionService.GetRegion(ctx, querier, phaseState.SourceRegion)
	if err != nil {
		return struct{}{}, fmt.Errorf("unable to get attacking region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegion(ctx, querier, phaseState.TargetRegion)
	if err != nil {
		return struct{}{}, fmt.Errorf("unable to get defending region: %w", err)
	}

	if sourceRegion.Troops-move.Troops < minTroopsToRetain {
		return struct{}{}, domainerrors.NewValidationError(
			"source region does not have enough troops",
		)
	}

	defeatedPlayerID, err := s.updateRegionTroops(ctx, querier, move, sourceRegion, targetRegion)
	if err != nil {
		return struct{}{}, fmt.Errorf("failed to update region troops: %w", err)
	}

	isDefenderEliminated, err := s.isDefenderEliminated(ctx, querier, defeatedPlayerID)
	if err != nil {
		return struct{}{}, fmt.Errorf("failed to check if defender is eliminated: %w", err)
	}

	if isDefenderEliminated {
		if err := s.handlePlayerEliminated(
			ctx,
			querier,
			defeatedPlayerID,
		); err != nil {
			return struct{}{}, fmt.Errorf("unable to handle player eliminated: %w", err)
		}
	}

	slog.DebugContext(ctx, "conquer executed successfully")

	return struct{}{}, nil
}

func (s *service) updateRegionTroops(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
) (int64, error) {
	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		sourceRegion,
		-move.Troops,
	); err != nil {
		return 0, fmt.Errorf("failed to decrease troops in source region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		targetRegion,
		move.Troops,
	); err != nil {
		return 0, fmt.Errorf("failed to increase troops in target region: %w", err)
	}

	slog.DebugContext(ctx, "troops updated successfully")

	defeatedPlayerID, err := s.regionService.UpdateRegionOwner(
		ctx,
		querier,
		targetRegion,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update region owner: %w", err)
	}

	return defeatedPlayerID, nil
}

func (s *service) isDefenderEliminated(
	ctx ctx.GameContext,
	querier db.Querier,
	defeatedPlayerID int64,
) (bool, error) {
	defeatedPlayerRegions, err := s.regionService.GetRegionsControlledByPlayer(
		ctx,
		querier,
		defeatedPlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get regions controlled by player: %w", err)
	}

	return len(defeatedPlayerRegions) == 0, nil
}

func (s *service) handlePlayerEliminated(
	ctx ctx.GameContext,
	querier db.Querier,
	eliminatedPlayerID int64,
) error {
	slog.InfoContext(ctx, "defending player has been eliminated", "defender", eliminatedPlayerID)

	if err := s.cardService.TransferCardsOwnership(ctx, querier, eliminatedPlayerID); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	if err := s.missionService.ReassignMissions(
		ctx,
		querier,
		eliminatedPlayerID,
	); err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	return nil
}
