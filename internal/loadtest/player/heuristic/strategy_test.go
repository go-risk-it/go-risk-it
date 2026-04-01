package heuristic //nolint:testpackage // whitebox tests access unexported helpers

import (
	"slices"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// testGraph builds a small 4-region graph:
//
//	A -- B -- D
//	 \  /
//	  C
func testGraph() *mapgraph.Graph {
	return &mapgraph.Graph{
		Regions:    []string{"A", "B", "C", "D"},
		Continents: make(map[string]mapgraph.ContinentDTO),
		RegionTo:   make(map[string]string),
		Neighbours: map[string]map[string]struct{}{
			"A": {"B": {}, "C": {}},
			"B": {"A": {}, "C": {}, "D": {}},
			"C": {"A": {}, "B": {}},
			"D": {"B": {}},
		},
	}
}

func TestFindCardCombo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cards []gamestate.Card
		want  []int64 // nil means no combo; otherwise sorted IDs
	}{
		{
			name: "3 cavalry returns 3-of-a-kind",
			cards: []gamestate.Card{
				{ID: 1, Type: gamestate.Cavalry},
				{ID: 2, Type: gamestate.Cavalry},
				{ID: 3, Type: gamestate.Cavalry},
			},
			want: []int64{1, 2, 3},
		},
		{
			name: "one-of-each",
			cards: []gamestate.Card{
				{ID: 10, Type: gamestate.Cavalry},
				{ID: 20, Type: gamestate.Infantry},
				{ID: 30, Type: gamestate.Artillery},
			},
			want: []int64{10, 20, 30},
		},
		{
			name: "2 cavalry + jolly",
			cards: []gamestate.Card{
				{ID: 1, Type: gamestate.Cavalry},
				{ID: 2, Type: gamestate.Cavalry},
				{ID: 99, Type: gamestate.Jolly},
			},
			want: []int64{1, 2, 99},
		},
		{
			name: "only 2 cards returns nil",
			cards: []gamestate.Card{
				{ID: 1, Type: gamestate.Cavalry},
				{ID: 2, Type: gamestate.Infantry},
			},
			want: nil,
		},
		{
			name: "no valid combo: 2 cavalry + 1 infantry, no jolly",
			cards: []gamestate.Card{
				{ID: 1, Type: gamestate.Cavalry},
				{ID: 2, Type: gamestate.Cavalry},
				{ID: 3, Type: gamestate.Infantry},
			},
			want: nil,
		},
		{
			name: "4 of same type returns first 3",
			cards: []gamestate.Card{
				{ID: 1, Type: gamestate.Infantry},
				{ID: 2, Type: gamestate.Infantry},
				{ID: 3, Type: gamestate.Infantry},
				{ID: 4, Type: gamestate.Infantry},
			},
			want: []int64{1, 2, 3},
		},
		{
			name: "prefers 3-of-a-kind over one-of-each",
			cards: []gamestate.Card{
				{ID: 1, Type: gamestate.Cavalry},
				{ID: 2, Type: gamestate.Cavalry},
				{ID: 3, Type: gamestate.Cavalry},
				{ID: 4, Type: gamestate.Infantry},
				{ID: 5, Type: gamestate.Artillery},
			},
			want: []int64{1, 2, 3},
		},
		{
			name:  "empty cards returns nil",
			cards: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := player.FindCardCombo(tt.cards)

			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}

				return
			}

			if got == nil {
				t.Fatalf("expected %v, got nil", tt.want)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("expected len %d, got len %d: %v", len(tt.want), len(got), got)
			}

			// Sort both for stable comparison.
			sortedGot := make([]int64, len(got))
			copy(sortedGot, got)
			slices.Sort(sortedGot)

			sortedWant := make([]int64, len(tt.want))
			copy(sortedWant, tt.want)
			slices.Sort(sortedWant)

			for i := range sortedWant {
				if sortedGot[i] != sortedWant[i] {
					t.Fatalf("expected %v, got %v", sortedWant, sortedGot)
				}
			}
		})
	}
}

func TestBuildRegionMap(t *testing.T) {
	t.Parallel()

	t.Run("normal snapshot with regions", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{
			BoardState: &gamestate.BoardState{
				Regions: []gamestate.Region{
					{ID: "r1", OwnerID: "u1", Troops: 5},
					{ID: "r2", OwnerID: "u2", Troops: 3},
				},
			},
		}

		m := buildRegionMap(snap)

		if len(m) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(m))
		}

		if m["r1"].OwnerID != "u1" {
			t.Fatalf("expected r1 owner u1, got %s", m["r1"].OwnerID)
		}

		if m["r2"].Troops != 3 {
			t.Fatalf("expected r2 troops 3, got %d", m["r2"].Troops)
		}
	})

	t.Run("nil BoardState returns empty map", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{BoardState: nil}
		m := buildRegionMap(snap)

		if len(m) != 0 {
			t.Fatalf("expected empty map, got %d entries", len(m))
		}
	})
}

func TestIsBorderRegion(t *testing.T) {
	t.Parallel()

	g := testGraph()
	s := New(g)

	regionMap := map[string]*gamestate.Region{
		"A": {ID: "A", OwnerID: "me", Troops: 5},
		"B": {ID: "B", OwnerID: "enemy", Troops: 3},
		"C": {ID: "C", OwnerID: "me", Troops: 4},
		"D": {ID: "D", OwnerID: "enemy", Troops: 2},
	}

	tests := []struct {
		name     string
		regionID string
		userID   string
		want     bool
	}{
		{
			name:     "region with enemy neighbour is border",
			regionID: "A",
			userID:   "me",
			want:     true, // A neighbours B (enemy)
		},
		{
			name:     "region with only friendly neighbours is not border",
			regionID: "D",
			userID:   "enemy",
			want:     false, // D neighbours B (also enemy)
		},
		{
			name:     "region not in graph returns false",
			regionID: "Z",
			userID:   "me",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := s.isBorderRegion(tt.regionID, tt.userID, regionMap)
			if got != tt.want {
				t.Fatalf(
					"isBorderRegion(%q, %q) = %v, want %v",
					tt.regionID,
					tt.userID,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestCountEnemyNeighbours(t *testing.T) {
	t.Parallel()

	g := testGraph()
	s := New(g)

	regionMap := map[string]*gamestate.Region{
		"A": {ID: "A", OwnerID: "me", Troops: 5},
		"B": {ID: "B", OwnerID: "enemy", Troops: 3},
		"C": {ID: "C", OwnerID: "me", Troops: 4},
		"D": {ID: "D", OwnerID: "enemy", Troops: 2},
	}

	tests := []struct {
		name     string
		regionID string
		userID   string
		want     int
	}{
		{
			name:     "B has 2 enemy neighbours for user me",
			regionID: "B",
			userID:   "enemy",
			want:     2, // B neighbours A(me), C(me), D(enemy) -> A,C are enemies of "enemy"
		},
		{
			name:     "D has 0 enemy neighbours for user enemy",
			regionID: "D",
			userID:   "enemy",
			want:     0, // D neighbours B(enemy) -> 0 enemies
		},
		{
			name:     "A has 1 enemy neighbour for user me",
			regionID: "A",
			userID:   "me",
			want:     1, // A neighbours B(enemy), C(me) -> B is enemy
		},
		{
			name:     "region not in graph returns 0",
			regionID: "Z",
			userID:   "me",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := s.countEnemyNeighbours(tt.regionID, tt.userID, regionMap)
			if got != tt.want {
				t.Fatalf(
					"countEnemyNeighbours(%q, %q) = %d, want %d",
					tt.regionID,
					tt.userID,
					got,
					tt.want,
				)
			}
		})
	}
}
