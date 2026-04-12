package orchestration

import (
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

// buildPersistenceEffect converts the pure-computation results (MoveEffect,
// AdvanceEffect, orchestration context) into a PersistenceEffect ready for
// database writes. This is a pure function: no IO, no context beyond the
// provided data, no errors (panics on invariant violations).
//
// When moveEffect is nil (voluntary advance), only PhaseTransition is populated.
// When advEffect is nil, no phase transition occurred.
//
// The function maps external IDs (region/user strings) to PersistenceEffect
// types which also use external IDs. The Persist() implementation resolves
// these to internal DB IDs.
func buildPersistenceEffect(
	moveCtx MoveContext,
	moveEffect *moveservice.MoveEffect,
	advEffect *moveservice.AdvanceEffect,
	prevState *snapshot.CachedGameState,
	gameOver bool,
) *PersistenceEffect {
	if prevState == nil {
		panic("buildPersistenceEffect: prevState must not be nil")
	}

	result := &PersistenceEffect{}

	// 1. MoveLog — only when moveEffect exists (not voluntary advance)
	if moveEffect != nil {
		result.MoveLog = buildMoveLogEntry(moveCtx)
		result.MoveExecution = buildMoveExecution(moveEffect, prevState)
	}

	// 2. Elimination — when conquer performer sets EliminatedUserID
	if moveEffect != nil && moveEffect.EliminatedUserID != "" {
		result.Elimination = &EliminationEffect{
			EliminatedUserID: moveEffect.EliminatedUserID,
			ConquerorUserID:  moveCtx.userID,
		}
	}

	// 3. CardDraw — when a card was drawn (reinforce with conquest, or advancer grants)
	if moveEffect != nil && len(moveEffect.CardDeltas) > 0 {
		result.CardDraw = extractCardDraw(moveEffect.CardDeltas, moveCtx.userID)
	}

	// 4. PhaseTransition — when advEffect triggers a phase change
	if advEffect != nil && phaseTypeChanged(prevState, advEffect.NewPhase) {
		result.PhaseTransition = buildPhaseTransition(
			advEffect,
			prevState,
		)
	}

	// 5. GameConclusion — when gameOver=true
	if gameOver {
		result.GameConclusion = &GameConclusion{
			WinnerUserID: moveCtx.userID,
		}
	}

	return result
}

// buildMoveLogEntry constructs the MoveLogEntry from MoveContext and MoveEffect.
func buildMoveLogEntry(
	moveCtx MoveContext,
) *MoveLogEntry {
	// Marshal moveData and result as JSON. Panic on error (invariant violation).
	moveData, err := json.Marshal(moveCtx.moveData)
	if err != nil {
		panic(fmt.Sprintf("buildMoveLogEntry: failed to marshal moveData: %v", err))
	}

	result, err := json.Marshal(moveCtx.result)
	if err != nil {
		panic(fmt.Sprintf("buildMoveLogEntry: failed to marshal result: %v", err))
	}

	return &MoveLogEntry{
		GameID:    moveCtx.gameID,
		UserID:    moveCtx.userID,
		PhaseType: sqlcPhaseToString(moveCtx.phaseType),
		MoveData:  moveData,
		Result:    result,
	}
}

// buildMoveExecution constructs the MoveExecution from MoveEffect and prevState.
// Splits the work into 5 groups: RegionTroopUpdates, OwnershipChanges,
// DeployableDelta, CardUnlinks, RegionBonuses.
func buildMoveExecution(
	moveEffect *moveservice.MoveEffect,
	prevState *snapshot.CachedGameState,
) *MoveExecution {
	exec := &MoveExecution{}

	// Build region index for fast lookup
	regionIndex := buildRegionIndex(prevState.PublicSnapshot.Regions)

	// Process RegionUpdates into troop deltas and ownership changes
	for _, update := range moveEffect.RegionUpdates {
		prev, ok := regionIndex[update.RegionID]
		if !ok {
			panic(fmt.Sprintf("buildMoveExecution: unknown region %q", update.RegionID))
		}

		// Troop delta
		if update.NewTroops != prev.Troops {
			exec.RegionTroopUpdates = append(exec.RegionTroopUpdates, RegionTroopUpdate{
				RegionID: prev.InternalID,
				Delta:    update.NewTroops - prev.Troops,
			})
		}

		// Ownership change (conquest)
		if update.NewOwner != prev.OwnerID {
			exec.OwnershipChanges = append(exec.OwnershipChanges, OwnershipChange{
				RegionID:       prev.InternalID,
				NewOwnerUserID: update.NewOwner,
				// OldOwnerInternalID is resolved by Persist() via GetPlayerByUserId
				OldOwnerInternalID: 0,
			})
		}
	}

	// DeployableDelta — from deploy move only
	if moveEffect.DeployableDelta != 0 {
		exec.DeployableDelta = &DeployableDelta{
			PlayerID: 0, // resolved by Persist()
			Delta:    moveEffect.DeployableDelta,
		}
	}

	// CardUnlinks — from cards move only
	if len(moveEffect.CardDeltas) > 0 {
		for _, delta := range moveEffect.CardDeltas {
			if len(delta.Lost) > 0 {
				exec.CardUnlinks = append(exec.CardUnlinks, delta.Lost...)
			}
		}
	}

	return exec
}

// extractCardDraw finds the first card draw in CardDeltas (if any) and returns
// it as a CardDraw. Returns nil if no card was gained.
func extractCardDraw(
	deltas []moveservice.CardDelta,
	currentUserID string,
) *CardDraw {
	for _, delta := range deltas {
		if len(delta.Gained) > 0 && delta.PlayerUserID == currentUserID {
			// Take the first gained card
			card := delta.Gained[0]

			return &CardDraw{
				CardID: card.ID,
				UserID: delta.PlayerUserID,
			}
		}
	}

	return nil
}

// phaseTypeChanged returns true if the new phase state represents a different
// phase type than the current phase.
func phaseTypeChanged(
	prevState *snapshot.CachedGameState,
	newPhase snapshot.PhaseState,
) bool {
	currentType := getPhaseStateType(prevState.PublicSnapshot.Phase.State)
	newType := getPhaseStateType(newPhase)

	return currentType != newType
}

// getPhaseStateType returns the PhaseType discriminator for a PhaseState.
func getPhaseStateType(state snapshot.PhaseState) snapshot.PhaseType {
	switch state.(type) {
	case snapshot.EmptyPhaseState:
		return "" // EmptyPhaseState has no type
	case snapshot.DeployPhaseState:
		return snapshot.PhaseDeploy
	case snapshot.ConquerPhaseState:
		return snapshot.PhaseConquer
	default:
		return ""
	}
}

// buildPhaseTransition constructs PhaseTransition from AdvanceEffect and prevState.
func buildPhaseTransition(
	advEffect *moveservice.AdvanceEffect,
	prevState *snapshot.CachedGameState,
) *PhaseTransition {
	phaseType := getPhaseStateType(advEffect.NewPhase)

	// Compute turn — increment if TurnEnded
	turn := prevState.Turn
	if advEffect.TurnEnded {
		turn++
	}

	// Build player refs from cached state
	players := make([]PlayerRef, 0, len(prevState.PublicSnapshot.Players))
	for _, p := range prevState.PublicSnapshot.Players {
		players = append(players, PlayerRef{
			UserID:     p.UserID,
			InternalID: 0, // resolved by Persist()
		})
	}

	transition := &PhaseTransition{
		Turn:      turn,
		PhaseType: snapshotPhaseToString(phaseType),
		Players:   players,
	}

	// Populate ConquerData or DeployData if applicable
	if conquerState, ok := advEffect.NewPhase.(snapshot.ConquerPhaseState); ok {
		transition.ConquerData = &ConquerData{
			SourceRegionName: conquerState.AttackingRegionID,
			TargetRegionName: conquerState.DefendingRegionID,
			MinTroops:        conquerState.MinTroopsToMove,
		}
	}

	if deployState, ok := advEffect.NewPhase.(snapshot.DeployPhaseState); ok {
		transition.DeployData = &DeployData{
			DeployableTroops: deployState.DeployableTroops,
		}
	}

	return transition
}

// buildRegionIndex creates a map from region external ID to RegionState for
// fast lookup.
func buildRegionIndex(
	regions []snapshot.RegionState,
) map[string]snapshot.RegionState {
	index := make(map[string]snapshot.RegionState, len(regions))
	for _, r := range regions {
		index[r.ID] = r
	}

	return index
}

// sqlcPhaseToString converts a sqlc.GamePhaseType to its string representation.
func sqlcPhaseToString(phase sqlc.GamePhaseType) string {
	switch phase {
	case sqlc.GamePhaseTypeCARDS:
		return "CARDS"
	case sqlc.GamePhaseTypeDEPLOY:
		return "DEPLOY"
	case sqlc.GamePhaseTypeATTACK:
		return "ATTACK"
	case sqlc.GamePhaseTypeCONQUER:
		return "CONQUER"
	case sqlc.GamePhaseTypeREINFORCE:
		return "REINFORCE"
	default:
		panic(fmt.Sprintf("sqlcPhaseToString: unknown phase %q", phase))
	}
}

// snapshotPhaseToString converts a snapshot.PhaseType to its string representation.
func snapshotPhaseToString(phase snapshot.PhaseType) string {
	switch phase {
	case snapshot.PhaseCards:
		return "CARDS"
	case snapshot.PhaseDeploy:
		return "DEPLOY"
	case snapshot.PhaseAttack:
		return "ATTACK"
	case snapshot.PhaseConquer:
		return "CONQUER"
	case snapshot.PhaseReinforce:
		return "REINFORCE"
	default:
		panic(fmt.Sprintf("snapshotPhaseToString: unknown phase %q", phase))
	}
}
