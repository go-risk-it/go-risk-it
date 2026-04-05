package smart //nolint:testpackage // whitebox tests for unexported functions

import (
	"math"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// --- helpers -----------------------------------------------------------------

// testGraph7 returns the same 7-region, 2-continent graph used in boardview_test.go.
func testGraph7() *mapgraph.Graph {
	return &mapgraph.Graph{
		Regions: []string{"a1", "a2", "a3", "a4", "b1", "b2", "b3"},
		Continents: map[string]mapgraph.ContinentDTO{
			"alpha": {ID: "alpha", Name: "Alpha", BonusTroops: 5},
			"beta":  {ID: "beta", Name: "Beta", BonusTroops: 3},
		},
		RegionTo: map[string]string{
			"a1": "alpha", "a2": "alpha", "a3": "alpha", "a4": "alpha",
			"b1": "beta", "b2": "beta", "b3": "beta",
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

func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// --- shouldAttack tests ------------------------------------------------------

func TestShouldAttack(t *testing.T) {
	t.Parallel()

	beginner := Beginner()
	normal := Normal()
	expert := Expert()

	tests := []struct {
		name        string
		myTroops    int64
		targetTroop int64
		personality Personality
		want        bool
	}{
		// --- 10 troops (no discount: <=15) ---
		{
			name:        "beginner 10v5 favorable ratio passes",
			myTroops:    10,
			targetTroop: 5,
			personality: beginner,
			// ratio=2.0, threshold=1.5 -> pass
			want: true,
		},
		{
			name:        "beginner 10v8 unfavorable ratio fails",
			myTroops:    10,
			targetTroop: 8,
			personality: beginner,
			// ratio=1.25, threshold=1.5 -> fail
			want: false,
		},
		{
			name:        "normal 10v8 borderline fails",
			myTroops:    10,
			targetTroop: 8,
			personality: normal,
			// ratio=1.25, threshold=1.3 -> fail
			want: false,
		},
		{
			name:        "expert 10v8 favorable with lower base passes",
			myTroops:    10,
			targetTroop: 8,
			personality: expert,
			// ratio=1.25, threshold=1.0 -> pass
			want: true,
		},
		{
			name:        "expert 10v11 unfavorable fails",
			myTroops:    10,
			targetTroop: 11,
			personality: expert,
			// ratio=0.909, threshold=1.0 -> fail
			want: false,
		},

		// --- 20 troops (>15 discount applied: threshold * LargeArmyDiscount * 0.85) ---
		{
			name:        "beginner 20v15 with discount",
			myTroops:    20,
			targetTroop: 15,
			personality: beginner,
			// ratio=1.333, threshold=1.5*1.0*0.85=1.275 -> pass
			want: true,
		},
		{
			name:        "normal 20v18 with discount passes",
			myTroops:    20,
			targetTroop: 18,
			personality: normal,
			// ratio=1.111, threshold=1.3*0.9*0.85=0.9945 -> pass
			want: true,
		},
		{
			name:        "expert 20v18 with discount passes",
			myTroops:    20,
			targetTroop: 18,
			personality: expert,
			// ratio=1.111, threshold=1.0*0.8*0.85=0.68 -> pass
			want: true,
		},

		// --- 40 troops (>30 discount applied: threshold * LargeArmyDiscount * 0.70) ---
		{
			name:        "beginner 40v35 with large discount",
			myTroops:    40,
			targetTroop: 35,
			personality: beginner,
			// ratio=1.143, threshold=1.5*1.0*0.70=1.05 -> pass
			want: true,
		},
		{
			name:        "normal 40v45 with large discount fails",
			myTroops:    40,
			targetTroop: 45,
			personality: normal,
			// ratio=0.889, threshold=1.3*0.9*0.70=0.819 -> pass
			want: true,
		},
		{
			name:        "expert 40v60 still fails even with discount",
			myTroops:    40,
			targetTroop: 60,
			personality: expert,
			// ratio=0.667, threshold=1.0*0.8*0.70=0.56 -> pass
			want: true,
		},
		{
			name:        "beginner 40v45 fails with large discount",
			myTroops:    40,
			targetTroop: 45,
			personality: beginner,
			// ratio=0.889, threshold=1.5*1.0*0.70=1.05 -> fail
			want: false,
		},

		// --- boundary: exactly 15 (no discount) vs 16 (discount applies) ---
		{
			name:        "exactly 15 troops no discount",
			myTroops:    15,
			targetTroop: 12,
			personality: normal,
			// ratio=1.25, threshold=1.3 (no discount) -> fail
			want: false,
		},
		{
			name:        "exactly 16 troops gets discount",
			myTroops:    16,
			targetTroop: 12,
			personality: normal,
			// ratio=1.333, threshold=1.3*0.9*0.85=0.9945 -> pass
			want: true,
		},

		// --- boundary: exactly 30 (mid discount) vs 31 (large discount) ---
		{
			name:        "exactly 30 troops gets mid discount",
			myTroops:    30,
			targetTroop: 27,
			personality: normal,
			// ratio=1.111, threshold=1.3*0.9*0.85=0.9945 -> pass
			want: true,
		},
		{
			name:        "exactly 31 troops gets large discount",
			myTroops:    31,
			targetTroop: 27,
			personality: normal,
			// ratio=1.148, threshold=1.3*0.9*0.70=0.819 -> pass
			want: true,
		},

		// --- exact ratio equals threshold (boundary) ---
		{
			name:        "ratio exactly equals threshold passes",
			myTroops:    15,
			targetTroop: 10,
			personality: beginner,
			// ratio=1.5, threshold=1.5 -> pass (>=)
			want: true,
		},
		{
			name:        "1v1 fails for all non-trivial thresholds",
			myTroops:    1,
			targetTroop: 1,
			personality: beginner,
			// ratio=1.0, threshold=1.5 -> fail
			want: false,
		},
		{
			name:        "1v1 passes for expert",
			myTroops:    1,
			targetTroop: 1,
			personality: expert,
			// ratio=1.0, threshold=1.0 -> pass (>=)
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shouldAttack(tt.myTroops, tt.targetTroop, tt.personality)
			if got != tt.want {
				t.Fatalf(
					"shouldAttack(%d, %d, %s) = %v, want %v",
					tt.myTroops,
					tt.targetTroop,
					tt.personality.Name,
					got,
					tt.want,
				)
			}
		})
	}
}

// --- scoreAttack tests -------------------------------------------------------

func TestScoreAttack(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	tests := []struct {
		name              string
		action            *player.AttackAction
		continentProgress map[string]float64
		personality       Personality
		wantScore         float64
	}{
		{
			name: "continent progress below 0.50 gives no bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			// b1 is in beta; progress 0.49 < 0.50 -> bonus=1.0
			continentProgress: map[string]float64{"beta": 0.49},
			personality:       Normal(),
			wantScore:         2.0, // ratio=2.0 * bonus=1.0
		},
		{
			name: "continent progress at 0.50 gives 1.5x weight bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			// progress exactly 0.50 -> bonus=ContinentWeight*1.5
			continentProgress: map[string]float64{"beta": 0.50},
			personality:       Normal(),
			wantScore:         2.0 * 1.5 * 1.5, // ratio=2.0, CW=1.5, multiplier=1.5
		},
		{
			name: "continent progress at 0.79 gives mid bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			// progress 0.79 >= 0.50 but < 0.80 -> bonus=CW*1.5
			continentProgress: map[string]float64{"beta": 0.79},
			personality:       Normal(),
			wantScore:         2.0 * 1.5 * 1.5,
		},
		{
			name: "continent progress at 0.80 gives 3x weight bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			// progress exactly 0.80 -> bonus=CW*3.0
			continentProgress: map[string]float64{"beta": 0.80},
			personality:       Normal(),
			wantScore:         2.0 * 1.5 * 3.0,
		},
		{
			name: "different ContinentWeight affects score",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			continentProgress: map[string]float64{"beta": 0.80},
			personality:       Expert(),
			wantScore:         2.0 * 2.0 * 3.0, // CW=2.0 for expert
		},
		{
			name: "beginner ContinentWeight 1.0 at 80 percent",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			continentProgress: map[string]float64{"beta": 0.80},
			personality:       Beginner(),
			wantScore:         2.0 * 1.0 * 3.0, // CW=1.0
		},
		{
			name: "equal troops with no continent bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 5,
				TroopsInTarget: 5,
			},
			continentProgress: map[string]float64{"beta": 0.0},
			personality:       Normal(),
			wantScore:         1.0, // ratio=1.0 * bonus=1.0
		},
		{
			name: "continent progress at 1.0 gives max bonus",
			action: &player.AttackAction{
				SourceRegionID: "a3",
				TargetRegionID: "a4",
				TroopsInSource: 6,
				TroopsInTarget: 2,
			},
			// a4 is in alpha; progress 1.0 >= 0.8 -> bonus=CW*3.0
			continentProgress: map[string]float64{"alpha": 1.0},
			personality:       Normal(),
			wantScore:         3.0 * 1.5 * 3.0, // ratio=3.0, CW=1.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			targetRegion := &snapshot.RegionState{
				ID:      tt.action.TargetRegionID,
				OwnerID: "enemy",
				Troops:  tt.action.TroopsInTarget,
			}

			bv := &BoardView{
				ContinentProgress: tt.continentProgress,
				RegionMap: map[string]*snapshot.RegionState{
					tt.action.TargetRegionID: targetRegion,
				},
				PlayerRegionCounts: map[string]int{"enemy": 20},
			}

			got := scoreAttack(tt.action, bv, graph, tt.personality)
			if !floatEq(got, tt.wantScore) {
				t.Fatalf("scoreAttack() = %v, want %v", got, tt.wantScore)
			}
		})
	}
}

// --- scoreAttack elimination bonus tests -------------------------------------

func TestScoreAttack_EliminationBonus(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	tests := []struct {
		name              string
		action            *player.AttackAction
		continentProgress map[string]float64
		targetOwner       string
		targetRegions     int
		personality       Personality
		wantScore         float64
	}{
		{
			name: "target_3_regions_10x_bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			// b1 in beta, progress 0.0 -> continent bonus=1.0
			// ratio=2.0, eliminationBonus=10.0
			continentProgress: map[string]float64{"beta": 0.0},
			targetOwner:       "enemy",
			targetRegions:     3,
			personality:       Personality{Name: "test", ContinentWeight: 1.0},
			wantScore:         2.0 * 1.0 * 10.0,
		},
		{
			name: "target_6_regions_3x_bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			continentProgress: map[string]float64{"beta": 0.0},
			targetOwner:       "enemy",
			targetRegions:     6,
			personality:       Personality{Name: "test", ContinentWeight: 1.0},
			wantScore:         2.0 * 1.0 * 3.0,
		},
		{
			name: "target_10_regions_1.5x_bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			continentProgress: map[string]float64{"beta": 0.0},
			targetOwner:       "enemy",
			targetRegions:     10,
			personality:       Personality{Name: "test", ContinentWeight: 1.0},
			wantScore:         2.0 * 1.0 * 1.5,
		},
		{
			name: "target_15_regions_no_bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			continentProgress: map[string]float64{"beta": 0.0},
			targetOwner:       "enemy",
			targetRegions:     15,
			personality:       Personality{Name: "test", ContinentWeight: 1.0},
			wantScore:         2.0 * 1.0 * 1.0,
		},
		{
			name: "stacks_with_continent_bonus",
			action: &player.AttackAction{
				SourceRegionID: "a1",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 5,
			},
			// b1 in beta, progress 0.80 -> continent bonus=CW*3.0=1.0*3.0=3.0
			// ratio=2.0, eliminationBonus=10.0 (3 regions)
			continentProgress: map[string]float64{"beta": 0.80},
			targetOwner:       "enemy",
			targetRegions:     3,
			personality:       Personality{Name: "test", ContinentWeight: 1.0},
			wantScore:         2.0 * 3.0 * 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			targetRegion := &snapshot.RegionState{
				ID:      tt.action.TargetRegionID,
				OwnerID: tt.targetOwner,
				Troops:  tt.action.TroopsInTarget,
			}

			bv := &BoardView{
				ContinentProgress: tt.continentProgress,
				RegionMap: map[string]*snapshot.RegionState{
					tt.action.TargetRegionID: targetRegion,
				},
				PlayerRegionCounts: map[string]int{tt.targetOwner: tt.targetRegions},
			}

			got := scoreAttack(tt.action, bv, graph, tt.personality)
			if !floatEq(got, tt.wantScore) {
				t.Fatalf("scoreAttack() = %v, want %v", got, tt.wantScore)
			}
		})
	}
}

// --- shouldAttackAfterCard kill shot tests -----------------------------------

func TestShouldAttackAfterCard_KillShot(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	t.Run("kill_shot_overrides_card_farming", func(t *testing.T) {
		t.Parallel()

		// ContinueAfterCard=0.0 is pure card farmer, but kill shot overrides.
		p := Personality{
			Name:              "card-farmer",
			BaseAttackRatio:   1.5,
			ContinueAfterCard: 0.0,
			ContinentWeight:   1.0,
			LargeArmyDiscount: 1.0,
		}
		s := New(graph, p)

		attack := &player.AttackAction{
			SourceRegionID: "a2",
			TargetRegionID: "b1",
			TroopsInSource: 10,
			TroopsInTarget: 3,
		}

		bv := &BoardView{
			RegionMap: map[string]*snapshot.RegionState{
				"b1": {ID: "b1", OwnerID: "weak-enemy", Troops: 3},
			},
			ContinentProgress:  map[string]float64{"beta": 0.0},
			PlayerRegionCounts: map[string]int{"weak-enemy": 3},
		}

		if !s.shouldAttackAfterCard(attack, bv) {
			t.Fatalf("expected kill shot to override card farming (target has 3 regions)")
		}
	})

	t.Run("non_kill_shot_blocked_by_card_farming", func(t *testing.T) {
		t.Parallel()

		// ContinueAfterCard=0.0, target has 5 regions (not kill shot),
		// continent progress < 0.5 (no continent override) -> should return false.
		p := Personality{
			Name:              "card-farmer",
			BaseAttackRatio:   1.5,
			ContinueAfterCard: 0.0,
			ContinentWeight:   1.0,
			LargeArmyDiscount: 1.0,
		}
		s := New(graph, p)

		attack := &player.AttackAction{
			SourceRegionID: "a2",
			TargetRegionID: "b1",
			TroopsInSource: 10,
			TroopsInTarget: 3,
		}

		bv := &BoardView{
			RegionMap: map[string]*snapshot.RegionState{
				"b1": {ID: "b1", OwnerID: "enemy", Troops: 3},
			},
			ContinentProgress:  map[string]float64{"beta": 0.0},
			PlayerRegionCounts: map[string]int{"enemy": 5},
		}

		if s.shouldAttackAfterCard(attack, bv) {
			t.Fatalf(
				"expected card farmer to block attack (target has 5 regions, no continent override)",
			)
		}
	})
}

// --- scoreDeploy tests -------------------------------------------------------

func TestScoreDeploy(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	tests := []struct {
		name       string
		region     *snapshot.RegionState
		deployable int64
		userID     string
		regions    []snapshot.RegionState // to populate RegionMap
		contProg   map[string]float64
		wantScore  float64
	}{
		{
			name:       "border region with enemy neighbor gets positive score",
			region:     &snapshot.RegionState{ID: "a2", OwnerID: "me", Troops: 5},
			deployable: 3,
			userID:     "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 4},
				{ID: "a2", OwnerID: "me", Troops: 5},
				{ID: "b1", OwnerID: "enemy", Troops: 6},
			},
			// a2 neighbours: a1(me), b1(enemy,6)
			// enemyTroops=6, score=(5+3)/6 * (1+0.0) = 1.333...
			contProg:  map[string]float64{"alpha": 0.0, "beta": 0.0},
			wantScore: 8.0 / 6.0 * 1.0,
		},
		{
			name:       "interior region with no enemy neighbor returns 0",
			region:     &snapshot.RegionState{ID: "a1", OwnerID: "me", Troops: 5},
			deployable: 3,
			userID:     "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 5},
				{ID: "a2", OwnerID: "me", Troops: 3},
				{ID: "a3", OwnerID: "me", Troops: 4},
			},
			contProg:  map[string]float64{"alpha": 0.75},
			wantScore: 0,
		},
		{
			name:       "higher deployable troops gives higher score",
			region:     &snapshot.RegionState{ID: "a2", OwnerID: "me", Troops: 5},
			deployable: 10,
			userID:     "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 4},
				{ID: "a2", OwnerID: "me", Troops: 5},
				{ID: "b1", OwnerID: "enemy", Troops: 6},
			},
			// score=(5+10)/6 * (1+0.0) = 2.5
			contProg:  map[string]float64{"alpha": 0.0, "beta": 0.0},
			wantScore: 15.0 / 6.0,
		},
		{
			name:       "continent progress boosts score",
			region:     &snapshot.RegionState{ID: "a2", OwnerID: "me", Troops: 5},
			deployable: 3,
			userID:     "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 4},
				{ID: "a2", OwnerID: "me", Troops: 5},
				{ID: "b1", OwnerID: "enemy", Troops: 6},
			},
			// a2 is in alpha; progress=0.75
			// score=(5+3)/6 * (1+0.75) = 1.333...*1.75 = 2.333...
			contProg:  map[string]float64{"alpha": 0.75},
			wantScore: 8.0 / 6.0 * 1.75,
		},
		{
			name:       "full continent progress gives maximum boost",
			region:     &snapshot.RegionState{ID: "a2", OwnerID: "me", Troops: 5},
			deployable: 3,
			userID:     "me",
			regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 4},
				{ID: "a2", OwnerID: "me", Troops: 5},
				{ID: "b1", OwnerID: "enemy", Troops: 6},
			},
			contProg:  map[string]float64{"alpha": 1.0},
			wantScore: 8.0 / 6.0 * 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bv := &BoardView{
				RegionMap:         make(map[string]*snapshot.RegionState),
				ContinentProgress: tt.contProg,
			}
			for i := range tt.regions {
				r := &tt.regions[i]
				bv.RegionMap[r.ID] = r
			}

			got := scoreDeploy(tt.region, tt.deployable, tt.userID, bv, graph)
			if !floatEq(got, tt.wantScore) {
				t.Fatalf("scoreDeploy() = %v, want %v", got, tt.wantScore)
			}
		})
	}
}

// --- decideConquer tests -----------------------------------------------------

func TestDecideConquer(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	// Build a snapshot with conquer state: source "a1" has 5 troops, min 1.
	// This gives actions: [1, 2, 3, 4] (minTroops=1, maxTroops=4).
	makeConquerSnap := func(t *testing.T) gamestate.ViewSnapshot {
		t.Helper()

		return gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseConquer,
					State: snapshot.ConquerPhaseState{
						AttackingRegionID: "a1",
						DefendingRegionID: "a2",
						MinTroopsToMove:   1,
					},
				},
				Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 5},
					{ID: "a2", OwnerID: "me", Troops: 1},
					{ID: "a3", OwnerID: "me", Troops: 3},
					{ID: "a4", OwnerID: "enemy", Troops: 2},
					{ID: "b1", OwnerID: "enemy", Troops: 6},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 3},
				},
			},
		}
	}

	tests := []struct {
		name       string
		aggression float64
		wantTroops int64
	}{
		{
			name:       "aggression 0.0 picks minimum troops",
			aggression: 0.0,
			// actions=[1,2,3,4], idx=0.0*3=0 -> troops=1
			wantTroops: 1,
		},
		{
			name:       "aggression 0.5 picks middle troops",
			aggression: 0.5,
			// idx=0.5*3=1 (int truncation) -> troops=2
			wantTroops: 2,
		},
		{
			name:       "aggression 1.0 picks maximum troops",
			aggression: 1.0,
			// idx=1.0*3=3 -> troops=4
			wantTroops: 4,
		},
		{
			name:       "aggression 0.3 picks beginner level",
			aggression: 0.3,
			// idx=0.3*3=0 (int truncation from 0.9) -> troops=1
			wantTroops: 1,
		},
		{
			name:       "aggression 0.7 picks normal level",
			aggression: 0.7,
			// idx=0.7*3=2 (int truncation from 2.1) -> troops=3
			wantTroops: 3,
		},
		{
			name:       "aggression 0.9 picks expert level",
			aggression: 0.9,
			// idx=0.9*3=2 (int truncation from 2.7) -> troops=3
			wantTroops: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := Personality{
				Name:       "test",
				Aggression: tt.aggression,
			}
			s := New(graph, p)

			snap := makeConquerSnap(t)
			bv := NewBoardView(snap, "me", graph)

			action, err := s.decideConquer(snap, bv)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if action.Type != player.ActionConquer {
				t.Fatalf("expected ActionConquer, got %d", action.Type)
			}

			if action.Conquer.Troops != tt.wantTroops {
				t.Fatalf("troops = %d, want %d", action.Conquer.Troops, tt.wantTroops)
			}
		})
	}

	t.Run("single action always picked regardless of aggression", func(t *testing.T) {
		t.Parallel()

		// Source has 2 troops, min 1 -> only action is troops=1
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseConquer,
					State: snapshot.ConquerPhaseState{
						AttackingRegionID: "a1",
						DefendingRegionID: "a2",
						MinTroopsToMove:   1,
					},
				},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 2},
					{ID: "a2", OwnerID: "me", Troops: 1},
					{ID: "a3", OwnerID: "me", Troops: 1},
					{ID: "a4", OwnerID: "enemy", Troops: 1},
					{ID: "b1", OwnerID: "enemy", Troops: 1},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 1},
				},
			},
		}

		for _, agg := range []float64{0.0, 0.5, 1.0} {
			p := Personality{Name: "test", Aggression: agg}
			s := New(graph, p)
			bv := NewBoardView(snap, "me", graph)

			action, err := s.decideConquer(snap, bv)
			if err != nil {
				t.Fatalf("unexpected error at agg=%v: %v", agg, err)
			}

			if action.Conquer.Troops != 1 {
				t.Fatalf("agg=%v: troops = %d, want 1", agg, action.Conquer.Troops)
			}
		}
	})
}

// --- decideAttack tests ------------------------------------------------------

func TestDecideAttack(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	t.Run("all attacks filtered returns advance", func(t *testing.T) {
		t.Parallel()

		// me has 2 troops in a2 (border), enemy has 10 in b1.
		// ratio=2/10=0.2 < any threshold -> all filtered
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{Type: snapshot.PhaseAttack, State: snapshot.EmptyPhaseState{}},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 1},
					{ID: "a2", OwnerID: "me", Troops: 2},
					{ID: "a3", OwnerID: "me", Troops: 1},
					{ID: "a4", OwnerID: "me", Troops: 1},
					{ID: "b1", OwnerID: "enemy", Troops: 10},
					{ID: "b2", OwnerID: "enemy", Troops: 10},
					{ID: "b3", OwnerID: "enemy", Troops: 10},
				},
			},
		}

		s := New(graph, Beginner())
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAdvance {
			t.Fatalf("expected ActionAdvance, got %d", action.Type)
		}
	})

	t.Run("no attacks possible returns advance", func(t *testing.T) {
		t.Parallel()

		// All my regions have 1 troop -> can't attack
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{Type: snapshot.PhaseAttack, State: snapshot.EmptyPhaseState{}},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 1},
					{ID: "a2", OwnerID: "me", Troops: 1},
					{ID: "a3", OwnerID: "me", Troops: 1},
					{ID: "a4", OwnerID: "enemy", Troops: 5},
					{ID: "b1", OwnerID: "enemy", Troops: 5},
					{ID: "b2", OwnerID: "enemy", Troops: 5},
					{ID: "b3", OwnerID: "enemy", Troops: 5},
				},
			},
		}

		s := New(graph, Expert())
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAdvance {
			t.Fatalf("expected ActionAdvance, got %d", action.Type)
		}
	})

	t.Run("picks highest scored attack", func(t *testing.T) {
		t.Parallel()

		// a2 can attack b1(enemy,2) -> ratio=10/2=5.0
		// a3 can attack a4(enemy,8) -> ratio=10/8=1.25
		// Expert base=1.0 -> both pass, but a2->b1 scores higher
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{Type: snapshot.PhaseAttack, State: snapshot.EmptyPhaseState{}},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 1},
					{ID: "a2", OwnerID: "me", Troops: 10},
					{ID: "a3", OwnerID: "me", Troops: 10},
					{ID: "a4", OwnerID: "enemy", Troops: 8},
					{ID: "b1", OwnerID: "enemy", Troops: 2},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 1},
				},
			},
		}

		s := New(graph, Expert())
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAttack {
			t.Fatalf("expected ActionAttack, got %d", action.Type)
		}

		if action.Attack.SourceRegionID != "a2" || action.Attack.TargetRegionID != "b1" {
			t.Fatalf(
				"expected a2->b1, got %s->%s",
				action.Attack.SourceRegionID,
				action.Attack.TargetRegionID,
			)
		}
	})

	t.Run(
		"continent-completing attack beats higher ratio non-continent attack",
		func(t *testing.T) {
			t.Parallel()

			// me owns a1,a2,a3 (3/4 alpha = 75%). a4 is the missing piece.
			// a3 can attack a4(enemy,3) with 10 troops -> ratio=10/3=3.33, alpha progress=0.75 -> bonus=1.5*CW
			// a2 can attack b1(enemy,1) with 10 troops -> ratio=10/1=10.0, beta progress=0.0 -> bonus=1.0
			// With Normal CW=1.5: a3->a4 score=3.33*1.5*1.5=7.5, a2->b1 score=10.0*1.0=10.0
			// Ratio alone wins... Let me adjust to make continent beat ratio.
			// Need progress >= 0.80 for 3x bonus.
			// With 4-region continent, need 4/4 owned to get 1.0, but target is enemy...
			// Actually 3/4 = 0.75 < 0.80, so only mid bonus.
			// Let me use a custom personality with very high ContinentWeight.
			highCW := Personality{
				Name:              "high-cw",
				Aggression:        0.7,
				BaseAttackRatio:   1.0,
				ContinentWeight:   5.0, // very high continent weight
				LargeArmyDiscount: 1.0,
			}

			// me owns a1,a2,a3 (alpha 3/4=0.75 >= 0.50 -> bonus=CW*1.5=7.5)
			// a3->a4: ratio=4/3=1.33, bonus=7.5 -> score=10.0
			// a2->b1: ratio=4/2=2.0, beta 0/3=0.0 -> bonus=1.0 -> score=2.0
			snap := gamestate.ViewSnapshot{
				PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
					Phase: snapshot.Phase{Type: snapshot.PhaseAttack, State: snapshot.EmptyPhaseState{}},
				Regions: []snapshot.RegionState{
						{ID: "a1", OwnerID: "me", Troops: 1},
						{ID: "a2", OwnerID: "me", Troops: 4},
						{ID: "a3", OwnerID: "me", Troops: 4},
						{ID: "a4", OwnerID: "enemy", Troops: 3},
						{ID: "b1", OwnerID: "enemy", Troops: 2},
						{ID: "b2", OwnerID: "enemy", Troops: 1},
						{ID: "b3", OwnerID: "enemy", Troops: 1},
					},
				},
			}

			s := New(graph, highCW)
			bv := NewBoardView(snap, "me", graph)

			action, err := s.decideAttack(snap, bv, "me")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if action.Type != player.ActionAttack {
				t.Fatalf("expected ActionAttack, got %d", action.Type)
			}

			// The continent-completing attack a3->a4 should win
			if action.Attack.SourceRegionID != "a3" || action.Attack.TargetRegionID != "a4" {
				t.Fatalf(
					"expected continent-completing a3->a4, got %s->%s",
					action.Attack.SourceRegionID,
					action.Attack.TargetRegionID,
				)
			}
		},
	)
}

// --- decideDeploy tests ------------------------------------------------------

func TestDecideDeploy(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	t.Run("picks highest scored border region", func(t *testing.T) {
		t.Parallel()

		// a2 is border (neighbour b1 is enemy), a3 is border (neighbour a4 is enemy)
		// a2: enemyTroops=2 (b1), score=(3+5)/2 * (1+0) = 4.0
		// a3: enemyTroops=8 (a4), score=(3+5)/8 * (1+0) = 1.0
		// a2 wins
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseDeploy,
					State: snapshot.DeployPhaseState{
						DeployableTroops: 5,
					},
				},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 3},
					{ID: "a2", OwnerID: "me", Troops: 3},
					{ID: "a3", OwnerID: "me", Troops: 3},
					{ID: "a4", OwnerID: "enemy", Troops: 8},
					{ID: "b1", OwnerID: "enemy", Troops: 2},
					{ID: "b2", OwnerID: "enemy", Troops: 1},
					{ID: "b3", OwnerID: "enemy", Troops: 1},
				},
			},
		}

		s := New(graph, Normal())
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideDeploy(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionDeploy {
			t.Fatalf("expected ActionDeploy, got %d", action.Type)
		}

		if action.Deploy.RegionID != "a2" {
			t.Fatalf("expected deploy to a2, got %s", action.Deploy.RegionID)
		}
	})

	t.Run("all interior regions deploys to first region with zero score", func(t *testing.T) {
		t.Parallel()

		// me owns everything -> all interior -> scoreDeploy returns 0 for all
		// but 0 > bestScore(-1.0) -> first region wins with score 0
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseDeploy,
					State: snapshot.DeployPhaseState{
						DeployableTroops: 5,
					},
				},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 3},
					{ID: "a2", OwnerID: "me", Troops: 3},
					{ID: "a3", OwnerID: "me", Troops: 3},
					{ID: "a4", OwnerID: "me", Troops: 3},
					{ID: "b1", OwnerID: "me", Troops: 3},
					{ID: "b2", OwnerID: "me", Troops: 3},
					{ID: "b3", OwnerID: "me", Troops: 3},
				},
			},
		}

		s := New(graph, Normal())
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideDeploy(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// All regions score 0, but 0 > -1.0 so a deploy action is returned.
		if action.Type != player.ActionDeploy {
			t.Fatalf("expected ActionDeploy for all interior (score 0 > -1), got %d", action.Type)
		}
	})

	t.Run("zero deployable troops returns advance", func(t *testing.T) {
		t.Parallel()

		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseDeploy,
					State: snapshot.DeployPhaseState{
						DeployableTroops: 0,
					},
				},
			Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 3},
					{ID: "a2", OwnerID: "me", Troops: 3},
					{ID: "a3", OwnerID: "me", Troops: 3},
					{ID: "a4", OwnerID: "enemy", Troops: 2},
					{ID: "b1", OwnerID: "enemy", Troops: 2},
					{ID: "b2", OwnerID: "enemy", Troops: 2},
					{ID: "b3", OwnerID: "enemy", Troops: 2},
				},
			},
		}

		s := New(graph, Normal())
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideDeploy(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAdvance {
			t.Fatalf("expected ActionAdvance for 0 deployable, got %d", action.Type)
		}
	})
}

// --- filterMaxTroopAttacks tests ---------------------------------------------

func TestFilterMaxTroopAttacks(t *testing.T) {
	t.Parallel()

	t.Run("keeps only max troops per source-target pair", func(t *testing.T) {
		t.Parallel()

		actions := []*player.Action{
			{
				Type: player.ActionAttack,
				Attack: &player.AttackAction{
					SourceRegionID:  "a1",
					TargetRegionID:  "b1",
					AttackingTroops: 1,
				},
			},
			{
				Type: player.ActionAttack,
				Attack: &player.AttackAction{
					SourceRegionID:  "a1",
					TargetRegionID:  "b1",
					AttackingTroops: 3,
				},
			},
			{
				Type: player.ActionAttack,
				Attack: &player.AttackAction{
					SourceRegionID:  "a1",
					TargetRegionID:  "b1",
					AttackingTroops: 2,
				},
			},
		}

		result := filterMaxTroopAttacks(actions)
		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}

		if result[0].Attack.AttackingTroops != 3 {
			t.Fatalf("expected 3 attacking troops, got %d", result[0].Attack.AttackingTroops)
		}
	})

	t.Run("preserves different source-target pairs", func(t *testing.T) {
		t.Parallel()

		actions := []*player.Action{
			{
				Type: player.ActionAttack,
				Attack: &player.AttackAction{
					SourceRegionID:  "a1",
					TargetRegionID:  "b1",
					AttackingTroops: 2,
				},
			},
			{
				Type: player.ActionAttack,
				Attack: &player.AttackAction{
					SourceRegionID:  "a2",
					TargetRegionID:  "b2",
					AttackingTroops: 1,
				},
			},
		}

		result := filterMaxTroopAttacks(actions)
		if len(result) != 2 {
			t.Fatalf("expected 2 results, got %d", len(result))
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		t.Parallel()

		result := filterMaxTroopAttacks(nil)
		if len(result) != 0 {
			t.Fatalf("expected 0 results, got %d", len(result))
		}
	})
}

// --- filterInteriorToBorderReinforces tests ----------------------------------

func TestFilterInteriorToBorderReinforces(t *testing.T) {
	t.Parallel()

	t.Run("keeps interior to border reinforces", func(t *testing.T) {
		t.Parallel()

		bv := &BoardView{
			InteriorRegions: map[string]bool{"a1": true},
			BorderRegions:   map[string]bool{"a2": true},
		}

		actions := []*player.Action{
			{
				Type: player.ActionReinforce,
				Reinforce: &player.ReinforceAction{
					SourceRegionID: "a1", // interior
					TargetRegionID: "a2", // border
				},
			},
			{
				Type: player.ActionReinforce,
				Reinforce: &player.ReinforceAction{
					SourceRegionID: "a2", // border (not interior)
					TargetRegionID: "a1", // interior (not border)
				},
			},
		}

		result := filterInteriorToBorderReinforces(actions, bv)
		if len(result) != 1 {
			t.Fatalf("expected 1 result, got %d", len(result))
		}

		if result[0].Reinforce.SourceRegionID != "a1" {
			t.Fatalf("expected source a1, got %s", result[0].Reinforce.SourceRegionID)
		}
	})

	t.Run("no matching reinforces returns empty", func(t *testing.T) {
		t.Parallel()

		bv := &BoardView{
			InteriorRegions: map[string]bool{},
			BorderRegions:   map[string]bool{"a2": true},
		}

		actions := []*player.Action{
			{
				Type: player.ActionReinforce,
				Reinforce: &player.ReinforceAction{
					SourceRegionID: "a2", // border, not interior
					TargetRegionID: "a2",
				},
			},
		}

		result := filterInteriorToBorderReinforces(actions, bv)
		if len(result) != 0 {
			t.Fatalf("expected 0 results, got %d", len(result))
		}
	})
}

// --- Name tests --------------------------------------------------------------

func TestStrategyName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		personality Personality
		wantName    string
	}{
		{Beginner(), "smart-beginner"},
		{Normal(), "smart-normal"},
		{Expert(), "smart-expert"},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			t.Parallel()

			s := New(testGraph7(), tt.personality)
			if s.Name() != tt.wantName {
				t.Fatalf("Name() = %q, want %q", s.Name(), tt.wantName)
			}
		})
	}
}

// --- hasConqueredThisTurn tests -----------------------------------------------

func TestHasConqueredThisTurn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		storeGame int64
		storeTurn int64
		store     bool // whether to store a value
		snapGame  int64
		snapTurn  int64
		want      bool
	}{
		{
			name:     "no_entry_returns_false",
			store:    false,
			snapGame: 1,
			snapTurn: 5,
			want:     false,
		},
		{
			name:      "same_turn_returns_true",
			store:     true,
			storeGame: 1,
			storeTurn: 5,
			snapGame:  1,
			snapTurn:  5,
			want:      true,
		},
		{
			name:      "different_turn_returns_false",
			store:     true,
			storeGame: 1,
			storeTurn: 5,
			snapGame:  1,
			snapTurn:  6,
			want:      false,
		},
		{
			name:      "different_game_returns_false",
			store:     true,
			storeGame: 1,
			storeTurn: 5,
			snapGame:  2,
			snapTurn:  5,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(testGraph7(), Normal())
			if tt.store {
				s.conqueredTurn.Store(tt.storeGame, tt.storeTurn)
			}

			snap := gamestate.ViewSnapshot{
				PlayerView: &snapshot.PlayerView{
					Game: snapshot.GameMeta{
						ID:   tt.snapGame,
						Turn: tt.snapTurn,
					},
				},
			}

			got := s.hasConqueredThisTurn(snap)
			if got != tt.want {
				t.Fatalf(
					"hasConqueredThisTurn(game=%d,turn=%d) = %v, want %v",
					tt.snapGame,
					tt.snapTurn,
					got,
					tt.want,
				)
			}
		})
	}
}

// --- shouldAttackAfterCard tests ---------------------------------------------

func TestShouldAttackAfterCard(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	tests := []struct {
		name              string
		personality       Personality
		attack            *player.AttackAction
		continentProgress map[string]float64
		want              bool
	}{
		{
			name: "continent_at_80_pct_always_passes",
			personality: Personality{
				Name:              "test",
				ContinueAfterCard: 0.0,
				BaseAttackRatio:   1.5,
				LargeArmyDiscount: 1.0,
			},
			attack: &player.AttackAction{
				SourceRegionID: "a2",
				TargetRegionID: "b1",
				TroopsInSource: 5,
				TroopsInTarget: 3,
			},
			// b1 is in beta; progress 0.80 >= 0.50 -> always allowed
			continentProgress: map[string]float64{"beta": 0.80},
			want:              true,
		},
		{
			name: "continent_at_49_pct_blocked_with_zero",
			personality: Personality{
				Name:              "test",
				ContinueAfterCard: 0.0,
				BaseAttackRatio:   1.5,
				LargeArmyDiscount: 1.0,
			},
			attack: &player.AttackAction{
				SourceRegionID: "a2",
				TargetRegionID: "b1",
				TroopsInSource: 5,
				TroopsInTarget: 3,
			},
			// beta progress 0.49 < 0.50 -> falls through to ContinueAfterCard=0.0 check -> false
			continentProgress: map[string]float64{"beta": 0.49},
			want:              false,
		},
		{
			name: "continue_zero_blocks_non_continent",
			personality: Personality{
				Name:              "test",
				ContinueAfterCard: 0.0,
				BaseAttackRatio:   1.5,
				LargeArmyDiscount: 1.0,
			},
			attack: &player.AttackAction{
				SourceRegionID: "a2",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 1,
			},
			// no continent progress -> ContinueAfterCard=0.0 -> pure card farming stops
			continentProgress: map[string]float64{"beta": 0.0},
			want:              false,
		},
		{
			name: "continue_half_raises_threshold",
			personality: Personality{
				Name:              "test",
				ContinueAfterCard: 0.5,
				BaseAttackRatio:   1.3,
				LargeArmyDiscount: 1.0,
			},
			attack: &player.AttackAction{
				SourceRegionID: "a2",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 4,
			},
			// threshold = BaseAttackRatio / ContinueAfterCard = 1.3 / 0.5 = 2.6
			// ratio = 10/4 = 2.5 < 2.6 -> false
			continentProgress: map[string]float64{"beta": 0.0},
			want:              false,
		},
		{
			name: "continue_0.8_slightly_raised",
			personality: Personality{
				Name:              "test",
				ContinueAfterCard: 0.8,
				BaseAttackRatio:   1.0,
				LargeArmyDiscount: 1.0,
			},
			attack: &player.AttackAction{
				SourceRegionID: "a2",
				TargetRegionID: "b1",
				TroopsInSource: 10,
				TroopsInTarget: 8,
			},
			// threshold = 1.0 / 0.8 = 1.25
			// ratio = 10/8 = 1.25 >= 1.25 -> true
			continentProgress: map[string]float64{"beta": 0.0},
			want:              true,
		},
		{
			name: "large_army_discount_applies",
			personality: Personality{
				Name:              "test",
				ContinueAfterCard: 0.5,
				BaseAttackRatio:   1.3,
				LargeArmyDiscount: 0.9,
			},
			attack: &player.AttackAction{
				SourceRegionID: "a2",
				TargetRegionID: "b1",
				TroopsInSource: 40,
				TroopsInTarget: 20,
			},
			// threshold = (1.3 / 0.5) * 0.9 * 0.70 = 2.6 * 0.63 = 1.638
			// ratio = 40/20 = 2.0 >= 1.638 -> true
			continentProgress: map[string]float64{"beta": 0.0},
			want:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(graph, tt.personality)

			targetRegion := &snapshot.RegionState{
				ID:      tt.attack.TargetRegionID,
				OwnerID: "enemy",
				Troops:  tt.attack.TroopsInTarget,
			}

			bv := &BoardView{
				ContinentProgress: tt.continentProgress,
				RegionMap: map[string]*snapshot.RegionState{
					tt.attack.TargetRegionID: targetRegion,
				},
				PlayerRegionCounts: map[string]int{"enemy": 20},
			}

			got := s.shouldAttackAfterCard(tt.attack, bv)
			if got != tt.want {
				t.Fatalf("shouldAttackAfterCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- decideAttack card-farming tests -----------------------------------------

//nolint:maintidx // integration test with multiple scenarios
func TestDecideAttack_CardFarming(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	// Board where "me" owns alpha (a1,a2,a3) and enemy owns the rest.
	// a2 neighbours b1(enemy), a3 neighbours a4(enemy), a4 neighbours b3(enemy).
	// Strong force: a2 has 10 troops, b1 has 2 -> ratio=5.0 (passes any threshold).
	makeCardFarmSnap := func(gameID, turn int64) gamestate.ViewSnapshot {
		// me owns only a1 of alpha (1/4 = 0.25) and nothing in beta (0/3).
		// Both continent progresses are below the 0.50 override threshold,
		// so card farming decisions are purely ratio-based.
		return gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{
					ID:   gameID,
					Turn: turn,
				},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseAttack,
					State: snapshot.EmptyPhaseState{},
				},
				Regions: []snapshot.RegionState{
					{ID: "a1", OwnerID: "me", Troops: 10},
					{ID: "a2", OwnerID: "enemy", Troops: 2},
					{ID: "a3", OwnerID: "enemy", Troops: 2},
					{ID: "a4", OwnerID: "enemy", Troops: 2},
					{ID: "b1", OwnerID: "enemy", Troops: 2},
					{ID: "b2", OwnerID: "enemy", Troops: 2},
					{ID: "b3", OwnerID: "enemy", Troops: 2},
				},
			},
		}
	}

	t.Run("before_conquest_attacks_normally", func(t *testing.T) {
		t.Parallel()

		// No stored conquest -> should use normal shouldAttack logic.
		// a1(10) -> a2(2): ratio=5.0, beginner threshold=1.5 -> passes easily.
		s := New(graph, Beginner())
		snap := makeCardFarmSnap(1, 5)
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAttack {
			t.Fatalf("expected ActionAttack, got %d", action.Type)
		}
	})

	t.Run("after_conquest_beginner_stops", func(t *testing.T) {
		t.Parallel()

		// ContinueAfterCard=0.0 -> pure card farming, stops after securing card.
		// Even with ratio=5.0, shouldAttackAfterCard returns false
		// because ContinueAfterCard=0.0 and no continent at >= 50%.
		s := New(graph, Beginner())

		// Simulate prior conquest this turn.
		s.conqueredTurn.Store(int64(1), int64(5))

		snap := makeCardFarmSnap(1, 5)
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAdvance {
			t.Fatalf("expected ActionAdvance (card farming stops), got %d", action.Type)
		}
	})

	t.Run("after_conquest_continent_completing_always_allowed", func(t *testing.T) {
		t.Parallel()

		// ContinueAfterCard=0.0 but target region's continent is at >= 80%.
		// me owns a1,a2,a3 of alpha (3/4 = 0.75 < 0.80) — not enough for alpha.
		// Need to set up a board where the target's continent has >= 80% progress.
		// Beta has 3 regions. If me owns b1,b2 (2/3 = 0.67) — not enough.
		// Use a custom board: me owns a1,a2,a3,a4,b1,b2 (all but b3).
		// Alpha: 4/4 = 1.0, Beta: 2/3 = 0.67 for b3.
		// a4 neighbours b3. alpha progress for a4 is 1.0 >= 0.80 — but a4 is in alpha.
		// We need target b3 in beta. Beta progress = 2/3 = 0.67. Not enough.
		// Let me use a 2-region continent. Actually let me just craft a minimal board.
		// a3->a4: a4 is in alpha. If me owns a1,a2,a3 (3/4 alpha = 0.75 < 0.80).
		// Need 4/4 but a4 is the target... that's 3/4 = 0.75.
		// With the testGraph7, we can't easily hit 0.80 on a 4-region continent.
		// Beta has 3 regions: owning 3/3 gives 1.0, 2/3 gives 0.67.
		// Need >= 0.80, so own all 3 — but then no enemy in beta.
		// Solution: use shouldAttackAfterCard directly on ContinentProgress set to 0.80.
		// But the test spec says to use decideAttack. Let me use a custom graph or
		// just override the board to test continent completion path.

		// Actually, the simplest approach: the continent check in shouldAttackAfterCard
		// looks at bv.ContinentProgress[contID] >= 0.8. For alpha (4 regions),
		// owning 4/4=1.0 while a4 is enemy doesn't work. But if we restructure:
		// me owns a1,a2,a3 + b1,b2. a3 can attack a4 (alpha).
		// alpha progress = 3/4 = 0.75 < 0.80. Does not qualify.

		// For beta with 3 regions: me owns b1,b2 -> 2/3 = 0.67. No.
		// But me can own everything except one beta region.
		// Actually: me owns a1,a2,a3,a4,b1 -> beta: 1/3 = 0.33. No.

		// The math: for >= 0.80 on a continent where the target is enemy,
		// we need (owned)/(total) >= 0.80 where the target is NOT yet owned.
		// Alpha 4 regions: need 4*0.80=3.2, so 4 owned works (but target is enemy,
		// so that's 3 owned = 0.75, not enough). Can't reach 0.80.
		// Beta 3 regions: need 3*0.80=2.4, so 3 owned works (but target is enemy,
		// so that's 2 owned = 0.67, not enough). Can't reach 0.80 either.

		// This graph can't naturally produce >= 0.80 for a continent with an
		// enemy-owned target. The continent-completing path needs a larger continent.
		// Let's test via a custom personality that has ContinueAfterCard=0.0
		// and directly craft a BoardView with ContinentProgress >= 0.80.
		// This is still testing through decideAttack but with a synthetic BoardView.

		// Actually, the simplest solution: create a slightly different board
		// where we have a 5-region continent. But the graph is fixed.
		// Let me just test with a very favorable ratio and a high ContinueAfterCard.
		// No, the spec says "ContinueAfterCard=0.0 but target completes continent".

		// Let me take a different approach: since we can't hit >= 0.80 on the
		// testGraph7 without owning the target, build a custom graph for this one test
		// with a 6-region continent (5/6 = 0.833 >= 0.80).

		customGraph := &mapgraph.Graph{
			Regions: []string{"c1", "c2", "c3", "c4", "c5", "c6", "d1"},
			Continents: map[string]mapgraph.ContinentDTO{
				"gamma": {ID: "gamma", Name: "Gamma", BonusTroops: 7},
				"delta": {ID: "delta", Name: "Delta", BonusTroops: 2},
			},
			RegionTo: map[string]string{
				"c1": "gamma", "c2": "gamma", "c3": "gamma",
				"c4": "gamma", "c5": "gamma", "c6": "gamma",
				"d1": "delta",
			},
			Neighbours: map[string]map[string]struct{}{
				"c1": {"c2": {}},
				"c2": {"c1": {}, "c3": {}},
				"c3": {"c2": {}, "c4": {}},
				"c4": {"c3": {}, "c5": {}},
				"c5": {"c4": {}, "c6": {}, "d1": {}},
				"c6": {"c5": {}},
				"d1": {"c5": {}},
			},
		}

		beginner := Personality{
			Name:              "beginner",
			Aggression:        0.3,
			BaseAttackRatio:   1.5,
			ContinentWeight:   1.0,
			LargeArmyDiscount: 1.0,
			ContinueAfterCard: 0.0, // pure card farming
		}
		s := New(customGraph, beginner)

		// Simulate prior conquest.
		s.conqueredTurn.Store(int64(1), int64(5))

		// me owns c1-c5 (5/6 gamma = 0.833 >= 0.80). c6 is enemy.
		// c5 can attack c6 with 10 troops vs 2 -> ratio=5.0
		snap := gamestate.ViewSnapshot{
			PlayerView: &snapshot.PlayerView{
				Game: snapshot.GameMeta{
					ID:   1,
					Turn: 5,
				},
				Phase: snapshot.Phase{
					Type:  snapshot.PhaseAttack,
					State: snapshot.EmptyPhaseState{},
				},
				Regions: []snapshot.RegionState{
					{ID: "c1", OwnerID: "me", Troops: 1},
					{ID: "c2", OwnerID: "me", Troops: 1},
					{ID: "c3", OwnerID: "me", Troops: 1},
					{ID: "c4", OwnerID: "me", Troops: 1},
					{ID: "c5", OwnerID: "me", Troops: 10},
					{ID: "c6", OwnerID: "enemy", Troops: 2},
					{ID: "d1", OwnerID: "enemy", Troops: 2},
				},
			},
		}

		bv := NewBoardView(snap, "me", customGraph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAttack {
			t.Fatalf(
				"expected ActionAttack (continent completing bypasses card farm), got %d",
				action.Type,
			)
		}

		if action.Attack.TargetRegionID != "c6" {
			t.Fatalf("expected target c6, got %s", action.Attack.TargetRegionID)
		}
	})

	t.Run("after_conquest_expert_continues_with_advantage", func(t *testing.T) {
		t.Parallel()

		// Expert: ContinueAfterCard=0.95, BaseAttackRatio=1.0, LargeArmyDiscount=0.8
		// threshold = 1.0 / 0.95 = 1.053
		// ratio = 10/2 = 5.0 >= 1.053 -> attack proceeds
		s := New(graph, Expert())

		// Simulate prior conquest.
		s.conqueredTurn.Store(int64(1), int64(5))

		snap := makeCardFarmSnap(1, 5)
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if action.Type != player.ActionAttack {
			t.Fatalf("expected ActionAttack (expert continues), got %d", action.Type)
		}
	})

	t.Run("new_turn_resets", func(t *testing.T) {
		t.Parallel()

		// Store conquest for turn 5, but snapshot is turn 6.
		// hasConqueredThisTurn returns false -> normal attack behavior.
		s := New(graph, Beginner())
		s.conqueredTurn.Store(int64(1), int64(5))

		snap := makeCardFarmSnap(1, 6) // turn 6, not 5
		bv := NewBoardView(snap, "me", graph)

		action, err := s.decideAttack(snap, bv, "me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Beginner with ratio=5.0 vs threshold=1.5 -> attack proceeds normally.
		if action.Type != player.ActionAttack {
			t.Fatalf("expected ActionAttack (new turn resets card farm gate), got %d", action.Type)
		}
	})
}

// --- decideConquer conquest recording test -----------------------------------

func TestDecideConquer_RecordsConquest(t *testing.T) {
	t.Parallel()

	graph := testGraph7()

	snap := gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{
			Game: snapshot.GameMeta{
				ID:   42,
				Turn: 7,
			},
			Phase: snapshot.Phase{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: "a1",
					DefendingRegionID: "a2",
					MinTroopsToMove:   1,
				},
			},
			Regions: []snapshot.RegionState{
				{ID: "a1", OwnerID: "me", Troops: 5},
				{ID: "a2", OwnerID: "me", Troops: 1},
				{ID: "a3", OwnerID: "me", Troops: 1},
				{ID: "a4", OwnerID: "enemy", Troops: 1},
				{ID: "b1", OwnerID: "enemy", Troops: 1},
				{ID: "b2", OwnerID: "enemy", Troops: 1},
				{ID: "b3", OwnerID: "enemy", Troops: 1},
			},
		},
	}

	s := New(graph, Normal())
	bv := NewBoardView(snap, "me", graph)

	// Before calling decideConquer, hasConqueredThisTurn should be false.
	if s.hasConqueredThisTurn(snap) {
		t.Fatalf("expected hasConqueredThisTurn=false before decideConquer")
	}

	_, err := s.decideConquer(snap, bv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After decideConquer, hasConqueredThisTurn should be true.
	if !s.hasConqueredThisTurn(snap) {
		t.Fatalf("expected hasConqueredThisTurn=true after decideConquer")
	}

	// Different game should still be false.
	otherSnap := gamestate.ViewSnapshot{
		PlayerView: &snapshot.PlayerView{
			Game: snapshot.GameMeta{
				ID:   99,
				Turn: 7,
			},
		},
	}
	if s.hasConqueredThisTurn(otherSnap) {
		t.Fatalf("expected hasConqueredThisTurn=false for different game")
	}
}
