package heuristic

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
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
	case snapshot.PhaseCards:
		return s.decideCards(snap)
	case snapshot.PhaseDeploy:
		return s.decideDeploy(snap, userID, buildRegionMap(snap))
	case snapshot.PhaseAttack:
		return s.decideAttack(snap, userID, buildRegionMap(snap))
	case snapshot.PhaseConquer:
		return s.decideConquer(snap)
	case snapshot.PhaseReinforce:
		return s.decideReinforce(snap, userID, buildRegionMap(snap))
	default:
		return nil, fmt.Errorf("unknown phase: %s", phase)
	}
}

func (s *Strategy) decideCards(snap gamestate.ViewSnapshot) (*player.Action, error) {
	cards := snap.Cards()
	if len(cards) < 3 {
		return player.NewAdvanceAction(snapshot.PhaseCards), nil
	}

	combo := player.FindCardCombo(cards)
	if combo == nil {
		return player.NewAdvanceAction(snapshot.PhaseCards), nil
	}

	return &player.Action{
		Type: player.ActionPlayCards,
		Cards: &player.CardsAction{
			Combinations: [][]int64{combo},
		},
	}, nil
}

func (s *Strategy) decideDeploy(
	snap gamestate.ViewSnapshot,
	userID string,
	regionMap map[string]*snapshot.RegionState,
) (*player.Action, error) {
	if snap.PlayerView == nil {
		return player.NewAdvanceAction(snapshot.PhaseDeploy), nil
	}

	state, ok := snap.PlayerView.Phase.State.(snapshot.DeployPhaseState)
	if !ok {
		return nil, fmt.Errorf("expected DeployPhaseState, got %T", snap.PlayerView.Phase.State)
	}

	if state.DeployableTroops == 0 {
		return player.NewAdvanceAction(snapshot.PhaseDeploy), nil
	}

	bestRegion := s.selectDeployTarget(userID, snap.MyRegions(userID), regionMap)
	if bestRegion == nil {
		return player.NewAdvanceAction(snapshot.PhaseDeploy), nil
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

// selectDeployTarget finds the best region to deploy troops to.
// Prefers weakest border regions (adjacent to enemies), falls back to any weak region.
func (s *Strategy) selectDeployTarget(
	userID string,
	myRegions []snapshot.RegionState,
	regionMap map[string]*snapshot.RegionState,
) *snapshot.RegionState {
	// Find the weakest border region (lowest troops adjacent to enemy).
	var bestRegion *snapshot.RegionState
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

	return bestRegion
}

func (s *Strategy) decideAttack(
	snap gamestate.ViewSnapshot,
	userID string,
	regionMap map[string]*snapshot.RegionState,
) (*player.Action, error) {
	myRegions := snap.MyRegions(userID)

	var bestSource, bestTarget *snapshot.RegionState
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
		return player.NewAdvanceAction(snapshot.PhaseAttack), nil
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
	if snap.PlayerView == nil {
		return nil, errors.New("nil PlayerView in conquer phase")
	}

	state, ok := snap.PlayerView.Phase.State.(snapshot.ConquerPhaseState)
	if !ok {
		return nil, fmt.Errorf("expected ConquerPhaseState, got %T", snap.PlayerView.Phase.State)
	}

	return &player.Action{
		Type: player.ActionConquer,
		Conquer: &player.ConquerAction{
			Troops: state.MinTroopsToMove,
		},
	}, nil
}

//nolint:cyclop // reinforcement decision tree with inherent branching
func (s *Strategy) decideReinforce(
	snap gamestate.ViewSnapshot,
	userID string,
	regionMap map[string]*snapshot.RegionState,
) (*player.Action, error) {
	myRegions := snap.MyRegions(userID)

	// Find the most interior region (fewest enemy neighbours) with the most troops.
	var bestSource *snapshot.RegionState
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

	//nolint:nestif // complex decision tree
	//nolint:nestif // reinforcement decision tree has inherent branching
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

	return player.NewAdvanceAction(snapshot.PhaseReinforce), nil
}

func (s *Strategy) isBorderRegion(
	regionID, userID string,
	regionMap map[string]*snapshot.RegionState,
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
	regionMap map[string]*snapshot.RegionState,
) int {
	count := 0

	for _, neighbourID := range s.graph.NeighboursOf(regionID) {
		if r, ok := regionMap[neighbourID]; ok && r.OwnerID != userID {
			count++
		}
	}

	return count
}

func buildRegionMap(snap gamestate.ViewSnapshot) map[string]*snapshot.RegionState {
	m := make(map[string]*snapshot.RegionState)
	if snap.PlayerView == nil {
		return m
	}

	for i := range snap.PlayerView.Regions {
		r := &snap.PlayerView.Regions[i]
		m[r.ID] = r
	}

	return m
}
