package attack

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

const (
	// conqueredRegionTroops is the troop count that indicates a region has been conquered.
	conqueredRegionTroops = 0
	// minTroopsToLaunchAttack is the minimum troops a region must exceed to launch an attack.
	// A region must retain at least one troop, so it needs more than this value.
	minTroopsToLaunchAttack = 1
)

func (s *service) Walk(wctx moveservice.WalkContext) (sqlc.GamePhaseType, error) {
	regions := MergeRegions(wctx.PrevSnapshot.Regions, wctx.Effect.RegionUpdates)

	if hasConquered(regions, wctx.CurrentUserID) {
		return sqlc.GamePhaseTypeCONQUER, nil
	}

	if wctx.Voluntary || !CanContinueAttacking(regions, wctx.CurrentUserID) {
		return sqlc.GamePhaseTypeREINFORCE, nil
	}

	return sqlc.GamePhaseTypeATTACK, nil
}

// MergeRegions applies region updates to a base snapshot, returning the
// effective region state after a move. Updates are matched by RegionID;
// unmatched base regions pass through unchanged.
func MergeRegions(
	base []snapshot.RegionState,
	updates []moveservice.RegionUpdate,
) []snapshot.RegionState {
	updateMap := make(map[string]moveservice.RegionUpdate, len(updates))
	for _, u := range updates {
		updateMap[u.RegionID] = u
	}

	merged := make([]snapshot.RegionState, 0, len(base))

	for _, region := range base {
		if update, ok := updateMap[region.ID]; ok {
			merged = append(merged, snapshot.RegionState{
				InternalID: region.InternalID,
				ID:         region.ID,
				OwnerID:    update.NewOwner,
				Troops:     update.NewTroops,
			})
		} else {
			merged = append(merged, region)
		}
	}

	return merged
}

// hasConquered returns true if any region not owned by the player has zero
// troops — indicating a successful conquest that must be followed up.
func hasConquered(regions []snapshot.RegionState, currentUserID string) bool {
	for _, r := range regions {
		if r.OwnerID != currentUserID && r.Troops == conqueredRegionTroops {
			return true
		}
	}

	return false
}

// CanContinueAttacking returns true if the player owns at least one region
// with more than 1 troop, meaning they can still launch attacks.
func CanContinueAttacking(regions []snapshot.RegionState, currentUserID string) bool {
	for _, r := range regions {
		if r.OwnerID == currentUserID && r.Troops > minTroopsToLaunchAttack {
			return true
		}
	}

	return false
}
