package smart

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// Strategy implements player.Strategy using valid move generation
// and personality-driven scoring.
type Strategy struct {
	graph       *mapgraph.Graph
	personality Personality

	// conqueredTurn tracks the last turn in which a conquest occurred,
	// keyed by GameState.ID. Enables card farming: after conquering once,
	// the bot can choose to stop attacking to preserve forces.
	conqueredTurn sync.Map // map[int64]int64
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
		return s.decideDeploy(snap, bv, userID)

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
	userID string,
) (*player.Action, error) {
	actions, err := GenerateDeploys(snap, bv)
	if err != nil {
		return nil, err
	}

	if len(actions) == 0 {
		return player.NewAdvanceAction(gamestate.Deploy), nil
	}

	// Parse deployable troops from phase state.
	var state gamestate.DeployPhaseState
	if err := json.Unmarshal(snap.GameState.Phase.State, &state); err != nil {
		return nil, fmt.Errorf("unmarshal deploy state: %w", err)
	}

	// Score border regions and pick highest.
	var best *player.Action

	bestScore := -1.0

	for _, a := range actions {
		score := scoreDeploy(
			bv.RegionMap[a.Deploy.RegionID],
			state.DeployableTroops,
			userID,
			bv,
			s.graph,
		)
		if score > bestScore {
			bestScore = score
			best = a
		}
	}

	if best == nil {
		return player.NewAdvanceAction(gamestate.Deploy), nil
	}

	return best, nil
}

func (s *Strategy) decideAttack(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
	userID string,
) (*player.Action, error) {
	actions := GenerateAttacks(snap, bv, userID, s.graph)
	if len(actions) == 0 {
		return player.NewAdvanceAction(gamestate.Attack), nil
	}

	// Filter by shouldAttack threshold, with stricter gate after securing card.
	maxAttacks := filterMaxTroopAttacks(actions)
	conquered := s.hasConqueredThisTurn(snap)

	var eligible []*player.Action

	for _, a := range maxAttacks {
		if conquered {
			if !s.shouldAttackAfterCard(a.Attack, bv) {
				continue
			}
		} else {
			if !shouldAttack(a.Attack.TroopsInSource, a.Attack.TroopsInTarget, s.personality) {
				continue
			}
		}

		eligible = append(eligible, a)
	}

	if len(eligible) == 0 {
		return player.NewAdvanceAction(gamestate.Attack), nil
	}

	// Pick highest-scored attack.
	var best *player.Action

	bestScore := -1.0

	for _, a := range eligible {
		score := scoreAttack(a.Attack, bv, s.graph, s.personality)
		if score > bestScore {
			bestScore = score
			best = a
		}
	}

	return best, nil
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
		observe.Warn(
			context.Background(),
			"no valid conquer moves generated, this should not happen",
		)

		return player.NewAdvanceAction(gamestate.Conquer), nil
	}

	// Deterministic aggression-proportional index: 0.0 -> first (min), 1.0 -> last (max).
	idx := int(s.personality.Aggression * float64(len(actions)-1))

	// Record conquest for card-farming tracking.
	s.conqueredTurn.Store(snap.GameState.ID, snap.GameState.Turn)

	return actions[idx], nil
}

func (s *Strategy) decideReinforce(
	snap gamestate.ViewSnapshot,
	bv *BoardView,
	userID string,
) (*player.Action, error) {
	actions := GenerateReinforces(snap, bv, userID, s.graph)
	if len(actions) == 0 {
		return player.NewAdvanceAction(gamestate.Reinforce), nil
	}

	// Phase 1: prefer interior→border reinforcements, pick randomly.
	borderReinforces := filterInteriorToBorderReinforces(actions, bv)
	if len(borderReinforces) > 0 {
		return pickRandom(borderReinforces), nil
	}

	return player.NewAdvanceAction(gamestate.Reinforce), nil
}

// shouldAttack returns true if the troop ratio meets the personality-adjusted threshold.
func shouldAttack(myTroops, targetTroops int64, p Personality) bool {
	ratio := float64(myTroops) / float64(targetTroops)

	threshold := p.BaseAttackRatio
	if myTroops > 30 {
		threshold *= p.LargeArmyDiscount * 0.70
	} else if myTroops > 15 {
		threshold *= p.LargeArmyDiscount * 0.85
	}

	return ratio >= threshold
}

// hasConqueredThisTurn reports whether the bot has already conquered
// a region during the current turn for the given game.
func (s *Strategy) hasConqueredThisTurn(snap gamestate.ViewSnapshot) bool {
	val, ok := s.conqueredTurn.Load(snap.GameState.ID)
	if !ok {
		return false
	}

	turn, ok := val.(int64)
	if !ok {
		return false
	}

	return turn == snap.GameState.Turn
}

// shouldAttackAfterCard applies a stricter threshold after securing the
// turn's card. Continent-completing attacks always pass. For other attacks,
// the threshold is raised inversely proportional to ContinueAfterCard.
func (s *Strategy) shouldAttackAfterCard(
	attack *player.AttackAction,
	bv *BoardView,
) bool {
	// Always allow continent-progressing attacks (>= 50% owned).
	contID := s.graph.RegionTo[attack.TargetRegionID]
	if bv.ContinentProgress[contID] >= 0.5 {
		return true
	}

	// Always allow kill shots on near-eliminated players.
	targetOwner := bv.RegionMap[attack.TargetRegionID].OwnerID
	if bv.PlayerRegionCounts[targetOwner] <= 3 {
		return true
	}

	// Pure card farming: stop after securing the card.
	if s.personality.ContinueAfterCard == 0.0 {
		return false
	}

	// Raised threshold: BaseAttackRatio / ContinueAfterCard.
	ratio := float64(attack.TroopsInSource) / float64(attack.TroopsInTarget)

	threshold := s.personality.BaseAttackRatio / s.personality.ContinueAfterCard
	if attack.TroopsInSource > 30 {
		threshold *= s.personality.LargeArmyDiscount * 0.70
	} else if attack.TroopsInSource > 15 {
		threshold *= s.personality.LargeArmyDiscount * 0.85
	}

	return ratio >= threshold
}

// scoreAttack scores an attack action based on troop ratio and continent progress bonus.
func scoreAttack(
	action *player.AttackAction,
	bv *BoardView,
	graph *mapgraph.Graph,
	p Personality,
) float64 {
	ratio := float64(action.TroopsInSource) / float64(action.TroopsInTarget)

	contID := graph.RegionTo[action.TargetRegionID]
	progress := bv.ContinentProgress[contID]

	bonus := 1.0
	if progress >= 0.8 {
		bonus = p.ContinentWeight * 3.0
	} else if progress >= 0.5 {
		bonus = p.ContinentWeight * 1.5
	}

	// Focus fire: prioritize attacking weak players to force eliminations.
	targetOwner := bv.RegionMap[action.TargetRegionID].OwnerID
	targetRegions := bv.PlayerRegionCounts[targetOwner]

	eliminationBonus := 1.0

	switch {
	case targetRegions <= 3:
		eliminationBonus = 10.0
	case targetRegions <= 6:
		eliminationBonus = 3.0
	case targetRegions <= 10:
		eliminationBonus = 1.5
	}

	return ratio * bonus * eliminationBonus
}

// scoreDeploy scores a deploy target based on troop advantage over strongest enemy neighbor
// and continent ownership progress.
func scoreDeploy(
	region *gamestate.Region,
	deployable int64,
	userID string,
	bv *BoardView,
	graph *mapgraph.Graph,
) float64 {
	enemyTroops := bv.BestEnemyNeighborTroops(region.ID, userID, graph)
	if enemyTroops == 0 {
		return 0 // interior region, skip
	}

	contID := graph.RegionTo[region.ID]
	progress := bv.ContinentProgress[contID]

	return (float64(region.Troops) + float64(deployable)) / float64(enemyTroops) * (1 + progress)
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
