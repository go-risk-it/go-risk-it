package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
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

	sourceRegion, targetRegion, err := s.loadRegionsFromCache(prev, conquerState)
	if err != nil {
		return struct{}{}, zero, err
	}

	sourceDBRegion := moveservice.ToDBRegion(sourceRegion)
	targetDBRegion := moveservice.ToDBRegion(targetRegion)

	if sourceDBRegion.Troops-move.Troops < minTroopsToRetain {
		return struct{}{}, zero, domainerrors.NewValidationError(
			"source region does not have enough troops",
		)
	}

	effect := s.buildMoveEffect(ctx, sourceDBRegion, targetDBRegion, move)

	isEliminated := s.isDefenderEliminatedFromCache(prev, targetDBRegion.UserID)
	if isEliminated {
		// Set EliminatedUserID so the persistence layer knows to trigger
		// card transfer and mission reassignment.
		effect.EliminatedUserID = targetDBRegion.UserID

		// Pre-compute card deltas and mission changes for the MoveEffect
		// before the DB mutations occur.
		eliminationEffect := s.buildEliminationEffect(prev, targetDBRegion.UserID, ctx.UserID())
		effect.CardDeltas = eliminationEffect.cardDeltas
		effect.Missions = eliminationEffect.missionChanges
	}

	return struct{}{}, effect, nil
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

// loadRegionsFromCache fetches both attacking and defending regions from the cached state.
func (s *service) loadRegionsFromCache(
	prev *snapshot.CachedGameState,
	state snapshot.ConquerPhaseState,
) (snapshot.RegionState, snapshot.RegionState, error) {
	sourceRegion, err := moveservice.FindRegion(
		prev.PublicSnapshot.Regions,
		state.AttackingRegionID,
	)
	if err != nil {
		return snapshot.RegionState{}, snapshot.RegionState{}, fmt.Errorf(
			"unable to get attacking region: %w", err,
		)
	}

	targetRegion, err := moveservice.FindRegion(
		prev.PublicSnapshot.Regions,
		state.DefendingRegionID,
	)
	if err != nil {
		return snapshot.RegionState{}, snapshot.RegionState{}, fmt.Errorf(
			"unable to get defending region: %w", err,
		)
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
	conquerorUserID string,
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
				{
					PlayerUserID: conquerorUserID,
					Gained:       eliminatedPrivate.Cards,
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

// isDefenderEliminatedFromCache checks if the defender is eliminated by counting
// their regions in the pre-move cached state. If they own exactly 1 region (the
// one being conquered), they have no regions left after the conquer.
func (s *service) isDefenderEliminatedFromCache(
	prev *snapshot.CachedGameState,
	eliminatedUserID string,
) bool {
	count := 0

	for _, r := range prev.PublicSnapshot.Regions {
		if r.OwnerID == eliminatedUserID {
			count++
		}
	}

	// The conquered region was their only one — they're eliminated.
	return count == 1
}
