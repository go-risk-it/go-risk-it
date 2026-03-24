//go:build invariant

package invariant

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/stretchr/testify/assert"
)

// Invariant is a named check that verifies a game property holds.
type Invariant struct {
	Name  string
	Check func(tb testing.TB, snap *GameSnapshot, prev *GameSnapshot)
}

// AllInvariants is the registry of all invariant checkers.
var AllInvariants = []Invariant{
	{"EveryRegionHasMinOneTroop", checkMinTroops},
	{"EveryRegionHasExactlyOneOwner", checkRegionOwnership},
	{"RegionCountEquals42", checkRegionCount},
	{"PhaseIsValid", checkPhaseValid},
	{"TurnNeverDecreases", checkTurnMonotonic},
	{"TroopDeltaMatchesPhase", checkTroopDelta},
	{
		"AllRegionsAccountedForInPlayerCounts",
		checkPlayerRegionConsistency,
	},
	{"EliminatedPlayersOwnNoRegions", checkEliminatedPlayers},
}

// CheckAll runs every registered invariant against the snapshot.
func CheckAll(
	tb testing.TB,
	snap *GameSnapshot,
	prev *GameSnapshot,
) {
	tb.Helper()

	for _, inv := range AllInvariants {
		inv.Check(tb, snap, prev)
	}
}

func checkMinTroops(
	tb testing.TB,
	snap *GameSnapshot,
	_ *GameSnapshot,
) {
	tb.Helper()

	// During CONQUER phase, the just-conquered region has
	// 0 troops until the conquer move fills it.
	if snap.Phase == sqlc.GamePhaseTypeCONQUER {
		zeroCount := 0
		for _, region := range snap.Regions {
			if region.Troops == 0 {
				zeroCount++
			}
		}

		assert.LessOrEqualf(tb, zeroCount, 1,
			"CONQUER: expected at most 1 region with "+
				"0 troops, got %d", zeroCount)

		return
	}

	for _, region := range snap.Regions {
		assert.GreaterOrEqualf(tb, region.Troops, int64(1),
			"region %s has %d troops (min 1)",
			region.ExternalReference, region.Troops)
	}
}

func checkRegionOwnership(
	tb testing.TB,
	snap *GameSnapshot,
	_ *GameSnapshot,
) {
	tb.Helper()

	for _, region := range snap.Regions {
		assert.NotEmptyf(tb, region.UserID,
			"region %s has no owner",
			region.ExternalReference)
	}
}

func checkRegionCount(
	tb testing.TB,
	snap *GameSnapshot,
	_ *GameSnapshot,
) {
	tb.Helper()

	assert.Lenf(tb, snap.Regions, 42,
		"expected 42 regions, got %d", len(snap.Regions))
}

func checkPhaseValid(
	tb testing.TB,
	snap *GameSnapshot,
	_ *GameSnapshot,
) {
	tb.Helper()

	validPhases := map[sqlc.GamePhaseType]bool{
		sqlc.GamePhaseTypeCARDS:     true,
		sqlc.GamePhaseTypeDEPLOY:    true,
		sqlc.GamePhaseTypeATTACK:    true,
		sqlc.GamePhaseTypeCONQUER:   true,
		sqlc.GamePhaseTypeREINFORCE: true,
	}
	assert.Truef(tb, validPhases[snap.Phase],
		"invalid phase: %s", snap.Phase)
}

func checkTurnMonotonic(
	tb testing.TB,
	snap *GameSnapshot,
	prev *GameSnapshot,
) {
	tb.Helper()

	if prev == nil {
		return
	}

	assert.GreaterOrEqualf(tb, snap.Turn, prev.Turn,
		"turn decreased from %d to %d", prev.Turn, snap.Turn)
}

func checkTroopDelta(
	tb testing.TB,
	snap *GameSnapshot,
	prev *GameSnapshot,
) {
	tb.Helper()

	if prev == nil || prev.IsGameOver() || snap.IsGameOver() {
		return
	}

	delta := snap.TotalTroops() - prev.TotalTroops()

	switch prev.Phase {
	case sqlc.GamePhaseTypeDEPLOY:
		assert.GreaterOrEqualf(tb, delta, int64(0),
			"troops decreased after DEPLOY: delta=%d", delta)

	case sqlc.GamePhaseTypeATTACK:
		assert.LessOrEqualf(tb, delta, int64(0),
			"troops increased after ATTACK: delta=%d", delta)

	case sqlc.GamePhaseTypeCONQUER:
		assert.Equalf(tb, int64(0), delta,
			"troops changed after CONQUER: delta=%d", delta)

	case sqlc.GamePhaseTypeREINFORCE:
		assert.Equalf(tb, int64(0), delta,
			"troops changed after REINFORCE: delta=%d", delta)

	case sqlc.GamePhaseTypeCARDS:
		assert.GreaterOrEqualf(tb, delta, int64(0),
			"troops decreased after CARDS: delta=%d", delta)
	}
}

func checkPlayerRegionConsistency(
	tb testing.TB,
	snap *GameSnapshot,
	_ *GameSnapshot,
) {
	tb.Helper()

	regionCounts := make(map[string]int64)
	for _, region := range snap.Regions {
		regionCounts[region.UserID]++
	}

	for _, player := range snap.Players {
		actual := regionCounts[player.UserID]
		assert.Equalf(tb, player.RegionCount, actual,
			"player %s: RegionCount=%d but owns %d regions",
			player.UserID, player.RegionCount, actual)
	}
}

func checkEliminatedPlayers(
	tb testing.TB,
	snap *GameSnapshot,
	_ *GameSnapshot,
) {
	tb.Helper()

	for _, player := range snap.Players {
		if player.RegionCount == 0 {
			owned := snap.RegionsOwnedBy(player.UserID)
			assert.Emptyf(tb, owned,
				"eliminated player %s still owns %d regions",
				player.UserID, len(owned))
		}
	}
}
