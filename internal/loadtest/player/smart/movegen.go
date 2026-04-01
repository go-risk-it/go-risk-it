package smart

import (
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// GenerateCardPlay returns a card play action if a valid combination exists,
// or an advance action otherwise. Cards are deterministic (no scoring needed).
func GenerateCardPlay(snap gamestate.ViewSnapshot) *player.Action {
	if snap.CardState == nil || len(snap.CardState.Cards) < 3 {
		return player.NewAdvanceAction(gamestate.Cards)
	}

	combo := player.FindCardCombo(snap.CardState.Cards)
	if combo == nil {
		return player.NewAdvanceAction(gamestate.Cards)
	}

	return &player.Action{
		Type: player.ActionPlayCards,
		Cards: &player.CardsAction{
			Combinations: [][]int64{combo},
		},
	}
}

// GenerateDeploys generates all valid deploy actions.
// Each action deploys all available troops to a single owned region.
func GenerateDeploys(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
) ([]*player.Action, error) {
	var state gamestate.DeployPhaseState

	if err := json.Unmarshal(snap.GameState.Phase.State, &state); err != nil {
		return nil, fmt.Errorf("unmarshal deploy state: %w", err)
	}

	if state.DeployableTroops == 0 {
		return nil, nil // caller should advance
	}

	var actions []*player.Action

	for _, r := range bv.MyRegions {
		actions = append(actions, &player.Action{
			Type: player.ActionDeploy,
			Deploy: &player.DeployAction{
				RegionID:      r.ID,
				CurrentTroops: r.Troops,
				DesiredTroops: r.Troops + state.DeployableTroops,
			},
		})
	}

	return actions, nil
}

// GenerateAttacks generates all valid attack actions.
// Each action attacks an enemy neighbour from an owned region with sufficient troops.
func GenerateAttacks(
	snap gamestate.ViewSnapshot, //nolint:unparam // interface conformance / future use
	bv *BoardView,
	userID string,
	graph *mapgraph.Graph,
) []*player.Action {
	var actions []*player.Action

	for _, src := range bv.MyRegions {
		if src.Troops < 2 { // need at least 2 to attack (keep 1 behind)
			continue
		}

		for _, neighbourID := range graph.NeighboursOf(src.ID) {
			tgt, ok := bv.RegionMap[neighbourID]
			if !ok || tgt.OwnerID == userID {
				continue
			}

			// Skip targets with 0 troops (server bug workaround).
			if tgt.Troops < 1 {
				continue
			}

			// Generate attacks with 1, 2, and 3 troops (all valid options).
			maxAttack := min(src.Troops-1, 3)
			for attackTroops := int64(1); attackTroops <= maxAttack; attackTroops++ {
				actions = append(actions, &player.Action{
					Type: player.ActionAttack,
					Attack: &player.AttackAction{
						SourceRegionID:  src.ID,
						TargetRegionID:  tgt.ID,
						TroopsInSource:  src.Troops,
						TroopsInTarget:  tgt.Troops,
						AttackingTroops: attackTroops,
					},
				})
			}
		}
	}

	return actions
}

// GenerateConquers generates all valid conquer actions.
// Reads CURRENT source troops from BoardView to avoid stale state bugs.
func GenerateConquers(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
) ([]*player.Action, error) {
	var state gamestate.ConquerPhaseState

	if err := json.Unmarshal(snap.GameState.Phase.State, &state); err != nil {
		return nil, fmt.Errorf("unmarshal conquer state: %w", err)
	}

	// Read current source troops from the board (critical: not cached/stale).
	srcRegion, ok := bv.RegionMap[state.AttackingRegionID]
	if !ok {
		return nil, fmt.Errorf("attacking region %s not found on board", state.AttackingRegionID)
	}

	minTroops := state.MinTroopsToMove
	maxTroops := srcRegion.Troops - 1 // must leave at least 1 in source

	if maxTroops < minTroops {
		// Only one valid option: move minimum.
		return []*player.Action{
			{
				Type:    player.ActionConquer,
				Conquer: &player.ConquerAction{Troops: minTroops},
			},
		}, nil
	}

	var actions []*player.Action

	for troops := minTroops; troops <= maxTroops; troops++ {
		actions = append(actions, &player.Action{
			Type:    player.ActionConquer,
			Conquer: &player.ConquerAction{Troops: troops},
		})
	}

	return actions, nil
}

// GenerateReinforces generates all valid reinforce actions.
// Source and target must both be owned and connected via owned regions.
func GenerateReinforces(
	snap gamestate.ViewSnapshot, //nolint:unparam // interface conformance / future use
	bv *BoardView,
	userID string,
	graph *mapgraph.Graph,
) []*player.Action {
	var actions []*player.Action

	for _, src := range bv.MyRegions {
		if src.Troops <= 1 {
			continue
		}

		for _, tgt := range bv.MyRegions {
			if src.ID == tgt.ID {
				continue
			}

			// Must be connected through owned territory.
			if !bv.ConnectedOwned(src.ID, tgt.ID, userID, graph) {
				continue
			}

			// Move all movable troops (src.Troops - 1).
			movable := src.Troops - 1
			actions = append(actions, &player.Action{
				Type: player.ActionReinforce,
				Reinforce: &player.ReinforceAction{
					SourceRegionID: src.ID,
					TargetRegionID: tgt.ID,
					TroopsInSource: src.Troops,
					TroopsInTarget: tgt.Troops,
					MovingTroops:   movable,
				},
			})
		}
	}

	return actions
}
