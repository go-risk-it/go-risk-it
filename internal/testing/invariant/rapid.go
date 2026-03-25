//go:build invariant

package invariant

import (
	"testing"

	"pgregory.net/rapid"
)

// RapidSimulationConfig holds parameters that are fixed per property test
// (not drawn by rapid). The seed is drawn inside RunGameProperty.
type RapidSimulationConfig struct {
	NumPlayers int
	MaxMoves   int
}

// RunGameProperty is meant to be called inside rapid.Check. It draws a
// random seed from rapid's generator for shrinking and reproducibility,
// builds a SimulationConfig, and delegates to RunGame.
//
// Because rapid.T does not satisfy testing.TB (sealed interface), the
// outer *testing.T must be passed separately for RunGame. Assertion
// failures from testify use t.Errorf which marks the *testing.T as
// failed; rapid.Check detects this after each iteration via tb.Failed().
// Hard failures (tb.Fatalf) cause runtime.Goexit and terminate the
// property iteration — rapid still reports the failure with the seed.
func RunGameProperty(
	t *testing.T,
	rt *rapid.T,
	harness *Harness,
	cfg RapidSimulationConfig,
) {
	t.Helper()

	seed := rapid.Uint64().Draw(rt, "seed")

	simCfg := SimulationConfig{
		NumPlayers: cfg.NumPlayers,
		MaxMoves:   cfg.MaxMoves,
		Seed:       seed,
	}

	RunGame(t, harness, simCfg)
}
