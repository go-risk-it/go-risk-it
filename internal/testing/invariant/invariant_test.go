//go:build invariant

package invariant_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/testing/invariant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.Equal(t, sqlc.GamePhaseTypeDEPLOY, snap.Phase)
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

	assert.NotEmpty(t, result.Winner)
	t.Logf("game %d ended after %d moves, winner: %s",
		result.GameID, result.MoveCount, result.Winner)
}
