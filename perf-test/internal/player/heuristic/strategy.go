package heuristic

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
)

// Strategy implements a simple rule-based AI.
type Strategy struct {
	graph *mapgraph.Graph
}

func New(graph *mapgraph.Graph) *Strategy {
	return &Strategy{graph: graph}
}

func (s *Strategy) Name() string {
	return "heuristic"
}

func (s *Strategy) DecideMove(snap gamestate.ViewSnapshot, userID string) (*player.Action, error) {
	phase := snap.CurrentPhase()

	switch phase {
	case gamestate.Cards:
		return s.decideCards(snap)
	case gamestate.Deploy:
		return s.decideDeploy(snap, userID, buildRegionMap(snap))
	case gamestate.Attack:
		return s.decideAttack(snap, userID, buildRegionMap(snap))
	case gamestate.Conquer:
		return s.decideConquer(snap)
	case gamestate.Reinforce:
		return s.decideReinforce(snap, userID, buildRegionMap(snap))
	default:
		return nil, fmt.Errorf("unknown phase: %s", phase)
	}
}

func (s *Strategy) decideCards(snap gamestate.ViewSnapshot) (*player.Action, error) {
	if snap.CardState == nil || len(snap.CardState.Cards) < 3 {
		return advanceAction(gamestate.Cards), nil
	}

	combo := findCardCombo(snap.CardState.Cards)
	if combo == nil {
		return advanceAction(gamestate.Cards), nil
	}

	return &player.Action{
		Type: player.ActionPlayCards,
		Cards: &player.CardsAction{
			Combinations: [][]int64{combo},
		},
	}, nil
}

func findCardCombo(cards []gamestate.Card) []int64 {
	byType := make(map[gamestate.CardType][]int64)
	var jollyIDs []int64

	for _, c := range cards {
		if c.Type == gamestate.Jolly {
			jollyIDs = append(jollyIDs, c.ID)
		} else {
			byType[c.Type] = append(byType[c.Type], c.ID)
		}
	}

	// Try 3-of-a-kind.
	for _, ids := range byType {
		if len(ids) >= 3 {
			return ids[:3]
		}
	}

	// Try one-of-each (cavalry + infantry + artillery).
	types := []gamestate.CardType{gamestate.Cavalry, gamestate.Infantry, gamestate.Artillery}
	var oneOfEach []int64

	for _, t := range types {
		if ids, ok := byType[t]; ok && len(ids) > 0 {
			oneOfEach = append(oneOfEach, ids[0])
		}
	}

	if len(oneOfEach) == 3 {
		return oneOfEach
	}

	// Try 2-of-a-kind + jolly (the only valid jolly combo).
	if len(jollyIDs) > 0 {
		for _, ids := range byType {
			if len(ids) >= 2 {
				return []int64{ids[0], ids[1], jollyIDs[0]}
			}
		}
	}

	return nil
}

func (s *Strategy) decideDeploy(
	snap gamestate.ViewSnapshot,
	userID string,
	regionMap map[string]*gamestate.Region,
) (*player.Action, error) {
	var state gamestate.DeployPhaseState

	if err := json.Unmarshal(snap.GameState.Phase.State, &state); err != nil {
		return nil, fmt.Errorf("unmarshal deploy state: %w", err)
	}

	if state.DeployableTroops == 0 {
		return advanceAction(gamestate.Deploy), nil
	}

	// Find the weakest border region (lowest troops adjacent to enemy).
	myRegions := snap.MyRegions(userID)
	var bestRegion *gamestate.Region
	bestScore := int64(math.MaxInt64)

	for i := range myRegions {
		r := &myRegions[i]
		if s.isBorderRegion(r.ID, userID, regionMap) && r.Troops < bestScore {
			bestScore = r.Troops
			bestRegion = r
		}
	}

	// Fallback: if no border region found, pick any region with fewest troops.
	if bestRegion == nil {
		for i := range myRegions {
			r := &myRegions[i]
			if r.Troops < bestScore {
				bestScore = r.Troops
				bestRegion = r
			}
		}
	}

	if bestRegion == nil {
		return advanceAction(gamestate.Deploy), nil
	}

	return &player.Action{
		Type: player.ActionDeploy,
		Deploy: &player.DeployAction{
			RegionID:      bestRegion.ID,
			CurrentTroops: bestRegion.Troops,
			DesiredTroops: bestRegion.Troops + state.DeployableTroops,
		},
	}, nil
}

func (s *Strategy) decideAttack(
	snap gamestate.ViewSnapshot,
	userID string,
	regionMap map[string]*gamestate.Region,
) (*player.Action, error) {
	myRegions := snap.MyRegions(userID)

	var bestSource, bestTarget *gamestate.Region
	bestRatio := 0.0

	for i := range myRegions {
		src := &myRegions[i]
		if src.Troops < 4 { // Need at least 4 to attack with 3
			continue
		}

		for _, neighbourID := range s.graph.NeighboursOf(src.ID) {
			tgt, ok := regionMap[neighbourID]
			if !ok || tgt.OwnerID == userID {
				continue
			}

			ratio := float64(src.Troops) / float64(tgt.Troops)
			if ratio >= 2.0 && ratio > bestRatio {
				bestRatio = ratio
				bestSource = src
				bestTarget = tgt
			}
		}
	}

	if bestSource == nil {
		return advanceAction(gamestate.Attack), nil
	}

	attackingTroops := min(bestSource.Troops-1, 3)

	return &player.Action{
		Type: player.ActionAttack,
		Attack: &player.AttackAction{
			SourceRegionID:  bestSource.ID,
			TargetRegionID:  bestTarget.ID,
			TroopsInSource:  bestSource.Troops,
			TroopsInTarget:  bestTarget.Troops,
			AttackingTroops: attackingTroops,
		},
	}, nil
}

func (s *Strategy) decideConquer(snap gamestate.ViewSnapshot) (*player.Action, error) {
	var state gamestate.ConquerPhaseState
	if err := json.Unmarshal(snap.GameState.Phase.State, &state); err != nil {
		return nil, fmt.Errorf("unmarshal conquer state: %w", err)
	}

	return &player.Action{
		Type: player.ActionConquer,
		Conquer: &player.ConquerAction{
			Troops: state.MinTroopsToMove,
		},
	}, nil
}

func (s *Strategy) decideReinforce(
	snap gamestate.ViewSnapshot,
	userID string,
	regionMap map[string]*gamestate.Region,
) (*player.Action, error) {
	myRegions := snap.MyRegions(userID)

	// Find the most interior region (fewest enemy neighbours) with the most troops.
	var bestSource *gamestate.Region
	bestInterior := -1
	bestTroops := int64(0)

	for i := range myRegions {
		r := &myRegions[i]
		if r.Troops <= 1 {
			continue
		}

		enemyNeighbours := s.countEnemyNeighbours(r.ID, userID, regionMap)
		interior := len(s.graph.NeighboursOf(r.ID)) - enemyNeighbours

		if interior > bestInterior || (interior == bestInterior && r.Troops > bestTroops) {
			bestInterior = interior
			bestTroops = r.Troops
			bestSource = r
		}
	}

	if bestSource == nil || !s.isBorderRegion(bestSource.ID, userID, regionMap) {
		// Find a border target to reinforce.
		if bestSource != nil {
			for _, neighbourID := range s.graph.NeighboursOf(bestSource.ID) {
				for j := range myRegions {
					if myRegions[j].ID == neighbourID &&
						s.isBorderRegion(neighbourID, userID, regionMap) {
						tgt := &myRegions[j]
						movable := bestSource.Troops - 1

						if movable > 0 {
							return &player.Action{
								Type: player.ActionReinforce,
								Reinforce: &player.ReinforceAction{
									SourceRegionID: bestSource.ID,
									TargetRegionID: tgt.ID,
									TroopsInSource: bestSource.Troops,
									TroopsInTarget: tgt.Troops,
									MovingTroops:   movable,
								},
							}, nil
						}
					}
				}
			}
		}
	}

	return advanceAction(gamestate.Reinforce), nil
}

func (s *Strategy) isBorderRegion(
	regionID, userID string,
	regionMap map[string]*gamestate.Region,
) bool {
	for _, neighbourID := range s.graph.NeighboursOf(regionID) {
		if r, ok := regionMap[neighbourID]; ok && r.OwnerID != userID {
			return true
		}
	}

	return false
}

func (s *Strategy) countEnemyNeighbours(
	regionID, userID string,
	regionMap map[string]*gamestate.Region,
) int {
	count := 0

	for _, neighbourID := range s.graph.NeighboursOf(regionID) {
		if r, ok := regionMap[neighbourID]; ok && r.OwnerID != userID {
			count++
		}
	}

	return count
}

func buildRegionMap(snap gamestate.ViewSnapshot) map[string]*gamestate.Region {
	m := make(map[string]*gamestate.Region)
	if snap.BoardState == nil {
		return m
	}

	for i := range snap.BoardState.Regions {
		r := &snap.BoardState.Regions[i]
		m[r.ID] = r
	}

	return m
}

func advanceAction(phase gamestate.PhaseType) *player.Action {
	return &player.Action{
		Type: player.ActionAdvance,
		Advance: &player.AdvanceAction{
			CurrentPhase: string(phase),
		},
	}
}
