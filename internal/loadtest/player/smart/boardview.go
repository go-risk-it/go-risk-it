package smart

import (
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
)

// BoardView provides derived board analysis rebuilt from each snapshot.
type BoardView struct {
	RegionMap         map[string]*gamestate.Region
	MyRegions         []*gamestate.Region
	BorderRegions     map[string]bool     // my regions with at least one enemy neighbour
	InteriorRegions   map[string]bool     // my regions with no enemy neighbours
	ContinentProgress map[string]float64  // continent ID -> fraction owned (0.0-1.0)
	ContinentRegions  map[string][]string // continent ID -> all region IDs in it
}

// NewBoardView builds a BoardView from the current snapshot for a given player.
//
//nolint:cyclop // board analysis with region + continent classification
func NewBoardView(
	snap gamestate.ViewSnapshot,
	userID string,
	graph *mapgraph.Graph,
) *BoardView {
	bv := &BoardView{
		RegionMap:         make(map[string]*gamestate.Region),
		BorderRegions:     make(map[string]bool),
		InteriorRegions:   make(map[string]bool),
		ContinentProgress: make(map[string]float64),
		ContinentRegions:  make(map[string][]string),
	}

	if snap.BoardState == nil {
		return bv
	}

	// Build region map.
	for i := range snap.BoardState.Regions {
		r := &snap.BoardState.Regions[i]
		bv.RegionMap[r.ID] = r
	}

	// Identify my regions and classify border vs interior.
	for i := range snap.BoardState.Regions {
		r := &snap.BoardState.Regions[i]
		if r.OwnerID != userID {
			continue
		}

		bv.MyRegions = append(bv.MyRegions, r)

		if hasEnemyNeighbour(r.ID, userID, bv.RegionMap, graph) {
			bv.BorderRegions[r.ID] = true
		} else {
			bv.InteriorRegions[r.ID] = true
		}
	}

	// Build continent region map and compute ownership progress.
	continentTotal := make(map[string]int)
	continentOwned := make(map[string]int)

	for _, regionID := range graph.Regions {
		contID := graph.RegionTo[regionID]
		bv.ContinentRegions[contID] = append(bv.ContinentRegions[contID], regionID)
		continentTotal[contID]++

		if r, ok := bv.RegionMap[regionID]; ok && r.OwnerID == userID {
			continentOwned[contID]++
		}
	}

	for contID, total := range continentTotal {
		if total > 0 {
			bv.ContinentProgress[contID] = float64(continentOwned[contID]) / float64(total)
		}
	}

	return bv
}

// ConnectedOwned checks if two regions owned by the same player are connected
// through a path of owned regions. Used for reinforce validation.
func (bv *BoardView) ConnectedOwned(
	source, target, userID string,
	graph *mapgraph.Graph,
) bool {
	if source == target {
		return true
	}

	visited := make(map[string]bool)
	queue := []string{source}
	visited[source] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbour := range graph.NeighboursOf(current) {
			if visited[neighbour] {
				continue
			}

			r, ok := bv.RegionMap[neighbour]
			if !ok || r.OwnerID != userID {
				continue
			}

			if neighbour == target {
				return true
			}

			visited[neighbour] = true
			queue = append(queue, neighbour)
		}
	}

	return false
}

func hasEnemyNeighbour(
	regionID, userID string,
	regionMap map[string]*gamestate.Region,
	graph *mapgraph.Graph,
) bool {
	for _, neighbourID := range graph.NeighboursOf(regionID) {
		if r, ok := regionMap[neighbourID]; ok && r.OwnerID != userID {
			return true
		}
	}

	return false
}
