package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

// BuildNewState produces a new CachedGameState by applying the move and advance
// effects to the previous state. It is a pure function: it never mutates prev,
// takes no context, touches no I/O, and panics only on invariant violations
// (nil inputs, unknown region IDs).
//
// The caller must supply the resolved targetPhase from the walker and winnerUserID
// from the mission check. When advEffect is nil, no advance occurred and only the
// move effect is applied.
func BuildNewState(
	prev *snapshot.CachedGameState,
	effect *service.MoveEffect,
	advEffect *service.AdvanceEffect,
	targetPhase sqlc.GamePhaseType,
	winnerUserID string,
) *snapshot.CachedGameState {
	if prev == nil {
		panic("BuildNewState: prev must not be nil")
	}

	if effect == nil {
		panic("BuildNewState: effect must not be nil")
	}

	regions := ApplyRegionUpdates(prev.PublicSnapshot.Regions, effect.RegionUpdates)
	privates := ApplyPrivateDeltas(prev.PrivateSnapshots, effect.CardDeltas, effect.Missions)

	if advEffect != nil {
		privates = applyCardDeltas(privates, advEffect.CardDeltas)
	}

	// Recompute players first — we need alive/dead status for turn skip logic.
	players := RecomputePlayers(prev.PublicSnapshot.Players, regions, privates)

	turn := prev.Turn
	if advEffect != nil && advEffect.TurnEnded {
		turn = nextAliveTurn(turn, players)
	}

	phase := resolvePhase(targetPhase, effect.UpdatedPhase, advEffect)
	conqueredInTurn := resolveConqueredInTurn(prev.ConqueredInTurn, targetPhase, advEffect)

	return &snapshot.CachedGameState{
		Turn:            turn,
		ConqueredInTurn: conqueredInTurn,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{
				ID:           prev.PublicSnapshot.Game.ID,
				Turn:         turn,
				WinnerUserID: winnerUserID,
			},
			Phase:   phase,
			Regions: regions,
			Players: players,
		},
		PrivateSnapshots: privates,
	}
}

// ApplyRegionUpdates returns a new slice of regions with the given updates
// applied. The input slice is never mutated. Panics if an update references
// a region ID that does not exist in the original slice.
func ApplyRegionUpdates(
	regions []snapshot.RegionState,
	updates []service.RegionUpdate,
) []snapshot.RegionState {
	cloned := make([]snapshot.RegionState, len(regions))
	copy(cloned, regions)

	if len(updates) == 0 {
		return cloned
	}

	index := make(map[string]int, len(cloned))
	for i, r := range cloned {
		index[r.ID] = i
	}

	for _, update := range updates {
		idx, ok := index[update.RegionID]
		if !ok {
			panic(fmt.Sprintf("ApplyRegionUpdates: unknown region ID %q", update.RegionID))
		}

		cloned[idx] = snapshot.RegionState{
			ID:      update.RegionID,
			OwnerID: update.NewOwner,
			Troops:  update.NewTroops,
		}
	}

	return cloned
}

// ApplyPrivateDeltas returns a new private snapshots map with card deltas and
// mission changes applied. The input map and its values are never mutated.
func ApplyPrivateDeltas(
	privates map[string]*snapshot.PlayerPrivate,
	cardDeltas []service.CardDelta,
	missions []service.MissionChange,
) map[string]*snapshot.PlayerPrivate {
	cloned := clonePrivates(privates)
	cloned = applyCardDeltas(cloned, cardDeltas)
	cloned = applyMissions(cloned, missions)

	return cloned
}

// RecomputePlayers derives a new player slice with CardCount and Status
// recomputed from the current regions and private snapshots. Name and Index
// are carried forward. The input slice is never mutated.
func RecomputePlayers(
	players []snapshot.PlayerState,
	regions []snapshot.RegionState,
	privates map[string]*snapshot.PlayerPrivate,
) []snapshot.PlayerState {
	ownershipCount := make(map[string]int, len(players))
	for _, r := range regions {
		ownershipCount[r.OwnerID]++
	}

	result := make([]snapshot.PlayerState, len(players))
	for index, player := range players {
		cardCount := int64(0)

		if priv, ok := privates[player.UserID]; ok {
			cardCount = int64(len(priv.Cards))
		}

		status := snapshot.PlayerAlive
		if ownershipCount[player.UserID] == 0 {
			status = snapshot.PlayerDead
		}

		result[index] = snapshot.PlayerState{
			UserID:    player.UserID,
			Name:      player.Name,
			Index:     player.Index,
			CardCount: cardCount,
			Status:    status,
		}
	}

	return result
}

// nextAliveTurn computes the next turn number, skipping dead players.
// This mirrors the DB-side logic in phase.Service.getNextTurn to keep
// the cached state consistent with the database turn sequence.
func nextAliveTurn(currentTurn int64, players []snapshot.PlayerState) int64 {
	count := int64(len(players))
	if count == 0 {
		return currentTurn + 1
	}

	// Build status-by-index for O(1) lookup.
	statusByIndex := make(map[int64]snapshot.PlayerStatus, len(players))
	for _, p := range players {
		statusByIndex[p.Index] = p.Status
	}

	turn := currentTurn + 1

	for range count {
		if statusByIndex[turn%count] == snapshot.PlayerAlive {
			return turn
		}

		turn++
	}

	// All players dead — shouldn't happen in practice.
	return currentTurn + 1
}

// resolveConqueredInTurn determines the ConqueredInTurn flag for the new state:
//   - Set true when entering the CONQUER phase (a conquest just happened).
//   - Reset to false when the turn ends (via AdvanceEffect.TurnEnded).
//   - Carry forward the previous value otherwise.
func resolveConqueredInTurn(
	prev bool,
	targetPhase sqlc.GamePhaseType,
	advEffect *service.AdvanceEffect,
) bool {
	if advEffect != nil && advEffect.TurnEnded {
		return false
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		return true
	}

	return prev
}

// resolvePhase picks the correct phase state and maps it to a snapshot.Phase.
// When an advance occurred, the advance effect's NewPhase takes precedence.
// Otherwise, the move effect's UpdatedPhase is used.
func resolvePhase(
	targetPhase sqlc.GamePhaseType,
	movePhaseState snapshot.PhaseState,
	advEffect *service.AdvanceEffect,
) snapshot.Phase {
	phaseType := sqlcPhaseToSnapshotPhase(targetPhase)

	var state snapshot.PhaseState
	if advEffect != nil {
		state = advEffect.NewPhase
	} else {
		state = movePhaseState
	}

	return snapshot.Phase{
		Type:  phaseType,
		State: state,
	}
}

// sqlcPhaseToSnapshotPhase converts a sqlc.GamePhaseType to a snapshot.PhaseType.
func sqlcPhaseToSnapshotPhase(phase sqlc.GamePhaseType) snapshot.PhaseType {
	switch phase {
	case sqlc.GamePhaseTypeCARDS:
		return snapshot.PhaseCards
	case sqlc.GamePhaseTypeDEPLOY:
		return snapshot.PhaseDeploy
	case sqlc.GamePhaseTypeATTACK:
		return snapshot.PhaseAttack
	case sqlc.GamePhaseTypeCONQUER:
		return snapshot.PhaseConquer
	case sqlc.GamePhaseTypeREINFORCE:
		return snapshot.PhaseReinforce
	default:
		panic(fmt.Sprintf("sqlcPhaseToSnapshotPhase: unknown phase %q", phase))
	}
}

// clonePrivates returns a deep copy of the private snapshots map. Each
// PlayerPrivate is cloned with its own Cards slice and Mission value.
func clonePrivates(
	privates map[string]*snapshot.PlayerPrivate,
) map[string]*snapshot.PlayerPrivate {
	cloned := make(map[string]*snapshot.PlayerPrivate, len(privates))
	for userID, priv := range privates {
		cards := make([]snapshot.CardState, len(priv.Cards))
		copy(cards, priv.Cards)

		cloned[userID] = &snapshot.PlayerPrivate{
			Cards:   cards,
			Mission: priv.Mission,
		}
	}

	return cloned
}

// applyCardDeltas applies card gained/lost deltas to the private snapshots.
// The input map must already be a clone (this function mutates the cloned map
// entries in place).
func applyCardDeltas(
	privates map[string]*snapshot.PlayerPrivate,
	deltas []service.CardDelta,
) map[string]*snapshot.PlayerPrivate {
	for _, delta := range deltas {
		priv, ok := privates[delta.PlayerUserID]
		if !ok {
			continue
		}

		// Remove lost cards.
		if len(delta.Lost) > 0 {
			lostSet := make(map[int64]struct{}, len(delta.Lost))
			for _, id := range delta.Lost {
				lostSet[id] = struct{}{}
			}

			kept := make([]snapshot.CardState, 0, len(priv.Cards))
			for _, c := range priv.Cards {
				if _, removed := lostSet[c.ID]; !removed {
					kept = append(kept, c)
				}
			}

			priv.Cards = kept
		}

		// Add gained cards.
		priv.Cards = append(priv.Cards, delta.Gained...)
	}

	return privates
}

// applyMissions applies mission reassignments to the private snapshots.
// The input map must already be a clone.
func applyMissions(
	privates map[string]*snapshot.PlayerPrivate,
	missions []service.MissionChange,
) map[string]*snapshot.PlayerPrivate {
	for _, missionChange := range missions {
		priv, ok := privates[missionChange.PlayerUserID]
		if !ok {
			continue
		}

		priv.Mission = missionChange.NewMission
	}

	return privates
}
