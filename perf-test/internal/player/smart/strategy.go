package smart

import (
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
)

// Strategy implements player.Strategy using valid move generation
// and random selection (Phase 1). Phase 2 adds scoring.
type Strategy struct {
	graph       *mapgraph.Graph
	personality Personality
}

// New creates a new smart strategy with the given personality.
func New(graph *mapgraph.Graph, personality Personality) *Strategy {
	return &Strategy{
		graph:       graph,
		personality: personality,
	}
}

func (s *Strategy) Name() string {
	return "smart-" + s.personality.Name
}

func (s *Strategy) DecideMove(
	snap gamestate.ViewSnapshot,
	userID string,
) (*player.Action, error) {
	phase := snap.CurrentPhase()
	bv := NewBoardView(snap, userID, s.graph)

	switch phase {
	case gamestate.Cards:
		return GenerateCardPlay(snap), nil

	case gamestate.Deploy:
		return s.decideDeploy(snap, bv)

	case gamestate.Attack:
		return s.decideAttack(snap, bv, userID)

	case gamestate.Conquer:
		return s.decideConquer(snap, bv)

	case gamestate.Reinforce:
		return s.decideReinforce(snap, bv, userID)

	default:
		return nil, fmt.Errorf("unknown phase: %s", phase)
	}
}

func (s *Strategy) decideDeploy(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
) (*player.Action, error) {
	actions, err := GenerateDeploys(snap, bv)
	if err != nil {
		return nil, err
	}

	if len(actions) == 0 {
		return advanceAction(gamestate.Deploy), nil
	}

	// Phase 1: prefer border regions, then pick randomly.
	borderActions := filterBorderDeploys(actions, bv)
	if len(borderActions) > 0 {
		return pickRandom(borderActions), nil
	}

	return pickRandom(actions), nil
}

func (s *Strategy) decideAttack(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
	userID string,
) (*player.Action, error) {
	actions := GenerateAttacks(snap, bv, userID, s.graph)
	if len(actions) == 0 {
		return advanceAction(gamestate.Attack), nil
	}

	// Phase 1: only consider max-troop attacks (3 when possible), pick randomly.
	maxAttacks := filterMaxTroopAttacks(actions)

	return pickRandom(maxAttacks), nil
}

func (s *Strategy) decideConquer(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
) (*player.Action, error) {
	actions, err := GenerateConquers(snap, bv)
	if err != nil {
		return nil, err
	}

	if len(actions) == 0 {
		log.Printf("[smart] WARNING: no valid conquer moves generated, this should not happen")

		return advanceAction(gamestate.Conquer), nil
	}

	// Always pick minTroopsToMove (first action). The board state may be stale
	// (gameState WS arrives before boardState WS), so max troops is unreliable.
	// minTroopsToMove comes from the gameState and is always valid server-side.
	return actions[0], nil
}

func (s *Strategy) decideReinforce(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
	userID string,
) (*player.Action, error) {
	actions := GenerateReinforces(snap, bv, userID, s.graph)
	if len(actions) == 0 {
		return advanceAction(gamestate.Reinforce), nil
	}

	// Phase 1: prefer interior→border reinforcements, pick randomly.
	borderReinforces := filterInteriorToBorderReinforces(actions, bv)
	if len(borderReinforces) > 0 {
		return pickRandom(borderReinforces), nil
	}

	return advanceAction(gamestate.Reinforce), nil
}

// filterBorderDeploys returns only deploy actions targeting border regions.
func filterBorderDeploys(actions []*player.Action, bv *BoardView) []*player.Action {
	var filtered []*player.Action

	for _, a := range actions {
		if bv.BorderRegions[a.Deploy.RegionID] {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

// filterMaxTroopAttacks keeps only attacks using the maximum troop count
// for each source→target pair.
func filterMaxTroopAttacks(actions []*player.Action) []*player.Action {
	type pairKey struct{ src, tgt string }

	best := make(map[pairKey]*player.Action)

	for _, a := range actions {
		key := pairKey{a.Attack.SourceRegionID, a.Attack.TargetRegionID}

		existing, ok := best[key]
		if !ok || a.Attack.AttackingTroops > existing.Attack.AttackingTroops {
			best[key] = a
		}
	}

	result := make([]*player.Action, 0, len(best))
	for _, a := range best {
		result = append(result, a)
	}

	return result
}

// filterInteriorToBorderReinforces returns reinforcements from interior
// regions to border regions.
func filterInteriorToBorderReinforces(
	actions []*player.Action,
	bv *BoardView,
) []*player.Action {
	var filtered []*player.Action

	for _, a := range actions {
		srcInterior := bv.InteriorRegions[a.Reinforce.SourceRegionID]
		tgtBorder := bv.BorderRegions[a.Reinforce.TargetRegionID]

		if srcInterior && tgtBorder {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

func pickRandom(actions []*player.Action) *player.Action {
	if len(actions) == 1 {
		return actions[0]
	}

	return actions[rand.IntN(len(actions))]
}
