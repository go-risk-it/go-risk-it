package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
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
	prev *snapshot.CachedGameState,
) (struct{}, moveservice.MoveEffect, error) {
	var zero moveservice.MoveEffect

	observe.SpanEvent(ctx, "performing_conquer_move",
		attribute.Int64("troops", move.Troops),
	)

	conquerState, err := s.extractConquerState(prev)
	if err != nil {
		return struct{}{}, zero, err
	}

	if conquerState.MinTroopsToMove > move.Troops {
		return struct{}{}, zero, domainerrors.NewValidationErrorf(
			"must move at least %d troops",
			conquerState.MinTroopsToMove,
		)
	}

	sourceRegion, targetRegion, err := s.loadRegions(ctx, querier, conquerState)
	if err != nil {
		return struct{}{}, zero, err
	}

	if sourceRegion.Troops-move.Troops < minTroopsToRetain {
		return struct{}{}, zero, domainerrors.NewValidationError(
			"source region does not have enough troops",
		)
	}

	eliminatedUserID := targetRegion.UserID

	defeatedPlayerID, err := s.updateRegionTroops(ctx, querier, move, sourceRegion, targetRegion)
	if err != nil {
		return struct{}{}, zero, fmt.Errorf("failed to update region troops: %w", err)
	}

	effect := s.buildMoveEffect(ctx, sourceRegion, targetRegion, move)

	return s.applyEliminationIfNeeded(
		ctx,
		querier,
		prev,
		defeatedPlayerID,
		eliminatedUserID,
		effect,
	)
}

// extractConquerState validates that phase state is ConquerPhaseState and returns it.
func (s *service) extractConquerState(
	prev *snapshot.CachedGameState,
) (snapshot.ConquerPhaseState, error) {
	conquerState, ok := prev.PublicSnapshot.Phase.State.(snapshot.ConquerPhaseState)
	if !ok {
		return snapshot.ConquerPhaseState{}, fmt.Errorf(
			"expected ConquerPhaseState, got %T",
			prev.PublicSnapshot.Phase.State,
		)
	}

	return conquerState, nil
}

// loadRegions fetches both attacking and defending regions.
func (s *service) loadRegions(
	ctx ctx.GameContext,
	querier db.Querier,
	state snapshot.ConquerPhaseState,
) (*sqlc.GetRegionsByGameRow, *sqlc.GetRegionsByGameRow, error) {
	sourceRegion, err := s.regionService.GetRegion(ctx, querier, state.AttackingRegionID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get attacking region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegion(ctx, querier, state.DefendingRegionID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get defending region: %w", err)
	}

	return sourceRegion, targetRegion, nil
}

// buildMoveEffect constructs the MoveEffect for region updates.
func (s *service) buildMoveEffect(
	ctx ctx.GameContext,
	sourceRegion, targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) moveservice.MoveEffect {
	return moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{
				RegionID:  sourceRegion.ExternalReference,
				NewOwner:  sourceRegion.UserID,
				NewTroops: sourceRegion.Troops - move.Troops,
			},
			{
				RegionID:  targetRegion.ExternalReference,
				NewOwner:  ctx.UserID(),
				NewTroops: targetRegion.Troops + move.Troops,
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}
}

// applyEliminationIfNeeded checks if defender was eliminated and applies elimination logic.
func (s *service) applyEliminationIfNeeded(
	ctx ctx.GameContext,
	querier db.Querier,
	prev *snapshot.CachedGameState,
	defeatedPlayerID int64,
	eliminatedUserID string,
	effect moveservice.MoveEffect,
) (struct{}, moveservice.MoveEffect, error) {
	isDefenderEliminated, err := s.isDefenderEliminated(ctx, querier, defeatedPlayerID)
	if err != nil {
		return struct{}{}, moveservice.MoveEffect{}, fmt.Errorf(
			"failed to check if defender is eliminated: %w",
			err,
		)
	}

	if isDefenderEliminated {
		// Pre-read the eliminated player's cards for the MoveEffect before the
		// DB transfer mutates ownership.
		eliminationEffect := s.buildEliminationEffect(prev, eliminatedUserID)

		if err := s.handlePlayerEliminated(ctx, querier, defeatedPlayerID); err != nil {
			return struct{}{}, moveservice.MoveEffect{}, fmt.Errorf(
				"unable to handle player eliminated: %w",
				err,
			)
		}

		effect.CardDeltas = eliminationEffect.cardDeltas
		effect.Missions = eliminationEffect.missionChanges
	}

	return struct{}{}, effect, nil
}

// eliminationEffect holds the pre-computed deltas for player elimination.
type eliminationEffect struct {
	cardDeltas     []moveservice.CardDelta
	missionChanges []moveservice.MissionChange
}

// buildEliminationEffect reads the cached state to compute card transfers and
// mission reassignments that will result from eliminating a player. This must
// be called BEFORE the DB mutations so the cached state is still accurate.
func (s *service) buildEliminationEffect(
	prev *snapshot.CachedGameState,
	eliminatedUserID string,
) eliminationEffect {
	var result eliminationEffect

	// Card transfer: eliminated player's cards move to the conquering player.
	if eliminatedPrivate, ok := prev.PrivateSnapshots[eliminatedUserID]; ok {
		if len(eliminatedPrivate.Cards) > 0 {
			lostCardIDs := make([]int64, 0, len(eliminatedPrivate.Cards))

			for _, c := range eliminatedPrivate.Cards {
				lostCardIDs = append(lostCardIDs, c.ID)
			}

			result.cardDeltas = []moveservice.CardDelta{
				{
					PlayerUserID: eliminatedUserID,
					Lost:         lostCardIDs,
				},
			}
		}
	}

	// Mission reassignment: any player whose mission was to eliminate the
	// eliminated player gets reassigned to TwentyFourTerritories.
	for playerUserID, private := range prev.PrivateSnapshots {
		if playerUserID == eliminatedUserID {
			continue
		}

		eliminateMission, ok := private.Mission.Detail.(snapshot.EliminatePlayerMission)
		if !ok {
			continue
		}

		if eliminateMission.TargetUserID != eliminatedUserID {
			continue
		}

		result.missionChanges = append(result.missionChanges, moveservice.MissionChange{
			PlayerUserID: playerUserID,
			NewMission: snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		})
	}

	return result
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
