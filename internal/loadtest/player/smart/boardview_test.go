package smart //nolint:testpackage // whitebox tests for unexported functions

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
)

// testGraphWithContinents builds a small 7-region graph across 2 continents:
//
//	Continent "alpha" (4 regions): a1, a2, a3, a4
//	  a1 -- a2, a1 -- a3, a3 -- a4, a2 -- b1 (cross-continent)
//	Continent "beta" (3 regions): b1, b2, b3
//	  b1 -- b2, b2 -- b3, b3 -- a4 (cross-continent)
func testGraphWithContinents() *mapgraph.Graph {
	return &mapgraph.Graph{
		Regions: []string{"a1", "a2", "a3", "a4", "b1", "b2", "b3"},
		Continents: map[string]mapgraph.ContinentDTO{
			"alpha": {ID: "alpha", Name: "Alpha", BonusTroops: 5},
			"beta":  {ID: "beta", Name: "Beta", BonusTroops: 3},
		},
		RegionTo: map[string]string{
			"a1": "alpha",
			"a2": "alpha",
			"a3": "alpha",
			"a4": "alpha",
			"b1": "beta",
			"b2": "beta",
			"b3": "beta",
		},
		Neighbours: map[string]map[string]struct{}{
			"a1": {"a2": {}, "a3": {}},
			"a2": {"a1": {}, "b1": {}},
			"a3": {"a1": {}, "a4": {}},
			"a4": {"a3": {}, "b3": {}},
			"b1": {"a2": {}, "b2": {}},
			"b2": {"b1": {}, "b3": {}},
			"b3": {"b2": {}, "a4": {}},
		},
	}
}

func TestBestEnemyNeighborTroops(t *testing.T) {
	t.Parallel()

	graph := testGraphWithContinents()

	tests := []struct {
		name     string
		regionID string
		userID   string
		regions  []snapshot.RegionState
		want     int64
	}{
		{
			name:     "region with one enemy neighbor returns its troops",
			regionID: "a1",
			userID:   "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 5},
				{ID: "a2", OwnerID: "enemy", Troops: 8},
				{ID: "a3", OwnerID: "me", Troops: 3},
			},
			want: 8,
		},
		{
			name:     "region with multiple enemy neighbors returns max troops",
			regionID: "a2",
			userID:   "enemy",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 5},
				{ID: "a2", OwnerID: "enemy", Troops: 10},
				{ID: "b1", OwnerID: "me", Troops: 12},
			},
			// a2 neighbours: a1(me, 5), b1(me, 12) -> max is 12
			want: 12,
		},
		{
			name:     "interior region with all friendly neighbors returns 0",
			regionID: "a1",
			userID:   "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 5},
				{ID: "a2", OwnerID: "me", Troops: 3},
				{ID: "a3", OwnerID: "me", Troops: 4},
			},
			// a1 neighbours: a2(me), a3(me) -> all friendly
			want: 0,
		},
		{
			name:     "region not in graph returns 0",
			regionID: "z99",
			userID:   "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 5},
			},
			want: 0,
		},
		{
			name:     "neighbor not in region map is skipped",
			regionID: "a1",
			userID:   "me",
			regions: []snapshot.RegionState{
				// a1's neighbors are a2 and a3; only a2 present and friendly
				{ID: "a1", OwnerID: "me", Troops: 5},
				{ID: "a2", OwnerID: "me", Troops: 3},
				// a3 missing from region map entirely
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bv := &BoardView{
				RegionMap: make(map[string]*snapshot.RegionState),
			}
			for i := range tt.regions {
				r := &tt.regions[i]
				bv.RegionMap[r.ID] = r
			}

			got := bv.BestEnemyNeighborTroops(tt.regionID, tt.userID, graph)
			if got != tt.want {
				t.Fatalf(
					"BestEnemyNeighborTroops(%q, %q) = %d, want %d",
					tt.regionID,
					tt.userID,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestNewBoardView(t *testing.T) {
	t.Parallel()

	graph := testGraphWithContinents()

	t.Run("nil PlayerView returns empty view", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{PlayerView: nil}
		bv := NewBoardView(snap, "me", graph)

		if len(bv.RegionMap) != 0 {
			t.Fatalf("expected empty RegionMap, got %d entries", len(bv.RegionMap))
		}

		if len(bv.MyRegions) != 0 {
			t.Fatalf("expected empty MyRegions, got %d entries", len(bv.MyRegions))
		}
	})

	t.Run("classifies border and interior regions", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 5},
					{ID: "a2", OwnerID: "me", Troops: 3},
					{ID: "a3", OwnerID: "me", Troops: 4},
					{ID: "a4", OwnerID: "enemy", Troops: 2},
					{ID: "b1", OwnerID: "enemy", Troops: 6},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 3},
				},
			},
		}

		bv := NewBoardView(snap, "me", graph)

		if len(bv.MyRegions) != 3 {
			t.Fatalf("expected 3 MyRegions, got %d", len(bv.MyRegions))
		}

		if !bv.InteriorRegions["a1"] {
			t.Fatalf("expected a1 to be interior")
		}

		if !bv.BorderRegions["a2"] {
			t.Fatalf("expected a2 to be border")
		}

		if !bv.BorderRegions["a3"] {
			t.Fatalf("expected a3 to be border")
		}
	})

	t.Run("computes continent progress", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 5},
					{ID: "a2", OwnerID: "me", Troops: 3},
					{ID: "a3", OwnerID: "me", Troops: 4},
					{ID: "a4", OwnerID: "enemy", Troops: 2},
					{ID: "b1", OwnerID: "enemy", Troops: 6},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 3},
				},
			},
		}

		bv := NewBoardView(snap, "me", graph)

		alphaProgress := bv.ContinentProgress["alpha"]
		if alphaProgress != 0.75 {
			t.Fatalf("alpha progress = %v, want 0.75", alphaProgress)
		}

		betaProgress := bv.ContinentProgress["beta"]
		if betaProgress != 0.0 {
			t.Fatalf("beta progress = %v, want 0.0", betaProgress)
		}
	})

	t.Run("full continent ownership gives 1.0 progress", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 5},
					{ID: "a2", OwnerID: "me", Troops: 3},
					{ID: "a3", OwnerID: "me", Troops: 4},
					{ID: "a4", OwnerID: "me", Troops: 2},
					{ID: "b1", OwnerID: "enemy", Troops: 6},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 3},
				},
			},
		}

		bv := NewBoardView(snap, "me", graph)

		alphaProgress := bv.ContinentProgress["alpha"]
		if alphaProgress != 1.0 {
			t.Fatalf("alpha progress = %v, want 1.0", alphaProgress)
		}
	})
}

func TestNewBoardView_PlayerRegionCounts(t *testing.T) {
	t.Parallel()

	graph := testGraphWithContinents()

	tests := []struct {
		name       string
		snap       gamestate.ViewSnapshot
		wantCounts map[string]int
	}{
		{
			name: "counts_regions_per_player",
			snap: gamestate.ViewSnapshot{
				PlayerView: &snapshot.PlayerView{
					Regions: []snapshot.RegionState{
						{ID: "a1", OwnerID: "alice", Troops: 1},
						{ID: "a2", OwnerID: "alice", Troops: 1},
						{ID: "a3", OwnerID: "alice", Troops: 1},
						{ID: "a4", OwnerID: "alice", Troops: 1},
						{ID: "b1", OwnerID: "alice", Troops: 1},
						{ID: "b2", OwnerID: "bob", Troops: 1},
						{ID: "b3", OwnerID: "bob", Troops: 1},
					},
				},
			},
			wantCounts: map[string]int{"alice": 5, "bob": 2},
		},
		{
			name: "empty_owner_excluded",
			snap: gamestate.ViewSnapshot{
				PlayerView: &snapshot.PlayerView{
					Regions: []snapshot.RegionState{
						{ID: "a1", OwnerID: "alice", Troops: 1},
						{ID: "a2", OwnerID: "", Troops: 1},
						{ID: "a3", OwnerID: "bob", Troops: 1},
						{ID: "a4", OwnerID: "", Troops: 1},
						{ID: "b1", OwnerID: "bob", Troops: 1},
						{ID: "b2", OwnerID: "bob", Troops: 1},
						{ID: "b3", OwnerID: "", Troops: 1},
					},
				},
			},
			wantCounts: map[string]int{"alice": 1, "bob": 3},
		},
		{
			name: "nil_PlayerView",
			snap: gamestate.ViewSnapshot{
				PlayerView: nil,
			},
			wantCounts: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bv := NewBoardView(tt.snap, "alice", graph)

			if len(bv.PlayerRegionCounts) != len(tt.wantCounts) {
				t.Fatalf(
					"PlayerRegionCounts has %d entries, want %d",
					len(bv.PlayerRegionCounts),
					len(tt.wantCounts),
				)
			}

			for player, wantCount := range tt.wantCounts {
				gotCount := bv.PlayerRegionCounts[player]
				if gotCount != wantCount {
					t.Fatalf(
						"PlayerRegionCounts[%q] = %d, want %d",
						player,
						gotCount,
						wantCount,
					)
				}
			}
		})
	}
}

func TestConnectedOwned(t *testing.T) {
	t.Parallel()

	graph := testGraphWithContinents()

	t.Run("same region is connected", func(t *testing.T) {
		t.Parallel()

		bv := &BoardView{RegionMap: make(map[string]*snapshot.RegionState)}

		if !bv.ConnectedOwned("a1", "a1", "me", graph) {
			t.Fatal("expected same region to be connected")
		}
	})

	t.Run("adjacent owned regions are connected", func(t *testing.T) {
		t.Parallel()

		bv := &BoardView{
			RegionMap: map[string]*snapshot.RegionState{
				"a1": {ID: "a1", OwnerID: "me", Troops: 5},
				"a2": {ID: "a2", OwnerID: "me", Troops: 3},
			},
		}

		if !bv.ConnectedOwned("a1", "a2", "me", graph) {
			t.Fatal("expected a1-a2 to be connected")
		}
	})

	t.Run("path through enemy territory is not connected", func(t *testing.T) {
		t.Parallel()

		// a1 -- a3 -- a4 but a3 is enemy
		bv := &BoardView{
			RegionMap: map[string]*snapshot.RegionState{
				"a1": {ID: "a1", OwnerID: "me", Troops: 5},
				"a3": {ID: "a3", OwnerID: "enemy", Troops: 4},
				"a4": {ID: "a4", OwnerID: "me", Troops: 2},
			},
		}

		if bv.ConnectedOwned("a1", "a4", "me", graph) {
			t.Fatal("expected a1-a4 to not be connected through enemy a3")
		}
	})
}
