//go:build invariant

package invariant_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"

	"github.com/go-risk-it/go-risk-it/internal/testing/invariant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

var harness *invariant.Harness

func TestMain(m *testing.M) {
	// Change to repo root so map.json and migration paths resolve.
	_, callerPath, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(
		filepath.Dir(callerPath), "..", "..", "..",
	)

	if err := os.Chdir(repoRoot); err != nil {
		panic("failed to chdir to repo root: " + err.Error())
	}

	harness = invariant.NewHarness()

	code := m.Run()

	harness.Close()
	os.Exit(code)
}

func TestCanCreateGame(t *testing.T) {
	t.Parallel()

	handle := harness.CreateGame(t, 3)
	snap := invariant.TakeSnapshot(t, harness, handle.GameID)

	assert.Equal(t, gameapi.GamePhaseTypeDEPLOY, snap.Phase)
	assert.Len(t, snap.Regions, 42)
	assert.Len(t, snap.Players, 3)

	for _, region := range snap.Regions {
		require.GreaterOrEqual(t, region.Troops, int64(1),
			"region %s has %d troops",
			region.ExternalReference, region.Troops)
	}

	invariant.CheckAll(t, snap, nil)
}

func TestGameToCompletion(t *testing.T) {
	t.Parallel()

	result := invariant.RunGame(t, harness, invariant.SimulationConfig{
		NumPlayers: 3,
		MaxMoves:   5000,
		Seed:       42,
	})

	require.True(t, result.Completed,
		"game did not complete within %d moves", 5000)
	assert.NotEmpty(t, result.Winner)
	t.Logf("game %d ended after %d moves, winner: %s",
		result.GameID, result.MoveCount, result.Winner)
}

// TestPropertyInvariantsHold runs random games via rapid's property-based
// testing framework. Each iteration draws a unique seed, plays moves up to
// MaxMoves, and checks all invariants after every move. Games that don't
// complete within MaxMoves are acceptable — the invariants were still
// verified on every move.
//
// Trial count is controlled by -rapid.checks flag (default 100).
// CI sets this to INVARIANT_GAME_COUNT (default 200).
// Extended fuzzing: -rapid.checks=10000.
func TestPropertyInvariantsHold(t *testing.T) {
	cfg := invariant.RapidSimulationConfig{
		NumPlayers: 3,
		MaxMoves:   200,
	}

	rapid.Check(t, func(rt *rapid.T) {
		invariant.RunGameProperty(t, rt, harness, cfg)
	})
}
