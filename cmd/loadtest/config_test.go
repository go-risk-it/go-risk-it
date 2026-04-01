package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t)

	// Server defaults.
	assert.Equal(t, "http://localhost:8000", cfg.Server.URL)
	assert.Empty(t, cfg.Server.AnonKey)
	assert.Equal(t, "cmd/loadtest/map.json", cfg.Server.MapFile)
	assert.Equal(t, "ws://localhost:8000", cfg.Server.WSURL)

	// Game defaults.
	assert.Equal(t, 4, cfg.Game.NumPlayers)
	assert.Equal(t, 10*time.Minute, cfg.Game.GameTimeout)
	assert.Equal(t, time.Duration(0), cfg.Game.ThinkTime)
	assert.Equal(t, "heuristic", cfg.Game.Strategy)

	// Mode defaults.
	assert.Equal(t, "batch", cfg.Run.Mode)
	assert.Empty(t, cfg.Run.Preset)
	assert.Equal(t, 1, cfg.Run.Batch.NumGames)
	assert.Equal(t, time.Duration(0), cfg.Run.Batch.RampUp)

	// Output defaults.
	assert.Equal(t, "text", cfg.Report.Output.Format)
	assert.False(t, cfg.Report.Output.SaveBaseline)

	// Staircase defaults.
	assert.Equal(t, 60*time.Second, cfg.Run.Staircase.HoldDuration)
	assert.True(t, cfg.Run.Staircase.StopOnBreach)

	// Adaptive defaults.
	assert.Equal(t, 5, cfg.Run.Adaptive.Increase)
	assert.Equal(t, 20, cfg.Run.Adaptive.MaxSteps)
	assert.Equal(t, 500, cfg.Run.Adaptive.MaxGames)

	// Ramp defaults.
	assert.Equal(t, 10, cfg.Run.Ramp.Rate)
	assert.Equal(t, 100, cfg.Run.Ramp.MaxGames)
	assert.InDelta(t, 0.10, cfg.Run.Ramp.ErrorThreshold, 0.001)
	assert.InDelta(t, 0.0, cfg.Run.Ramp.Multiplier, 0.001)

	// Chaos defaults (all zero).
	assert.InDelta(t, 0.0, cfg.Chaos.DisconnectRate, 0.001)
	assert.InDelta(t, 0.0, cfg.Chaos.SlowMoveRate, 0.001)
	assert.InDelta(t, 0.0, cfg.Chaos.ErrorMoveRate, 0.001)
}

func TestConfig_ApplyPreset_Staircase(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t, "--preset", "staircase-light")
	require.NoError(t, cfg.ApplyPreset())

	assert.Equal(t, "staircase", cfg.Run.Mode)
	assert.Equal(t, []int{5, 10, 20, 40}, cfg.Run.staircaseCfg.Steps)
}

func TestConfig_ApplyPreset_Ramp(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t, "--preset", "ramp-slow")
	require.NoError(t, cfg.ApplyPreset())

	assert.Equal(t, "ramp", cfg.Run.Mode)
	assert.Equal(t, 5, cfg.Run.rampCfg.GamesPerMinute)
	assert.Equal(t, 100, cfg.Run.rampCfg.MaxGames)
}

func TestConfig_ApplyPreset_Batch(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t, "--preset", "smoke")
	require.NoError(t, cfg.ApplyPreset())

	assert.Equal(t, "batch", cfg.Run.Mode) // batch presets don't force mode
	assert.Equal(t, 1, cfg.Run.Batch.NumGames)
	assert.Equal(t, 3, cfg.Game.NumPlayers)
}

func TestConfig_ApplyPreset_ChaosOverride(t *testing.T) {
	t.Parallel()

	// CLI chaos flags should override preset chaos.
	cfg := parseTestFlags(t, "--preset", "chaos-light", "--chaos-disconnect", "0.50")
	require.NoError(t, cfg.ApplyPreset())

	// CLI value takes precedence — preset chaos is NOT applied.
	assert.InDelta(t, 0.50, cfg.Chaos.DisconnectRate, 0.001)
}

func TestConfig_ApplyPreset_Unknown(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t)
	cfg.Run.Preset = "nonexistent"

	err := cfg.ApplyPreset()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestConfig_Validate_MissingAnonKey(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t)
	cfg.Server.AnonKey = ""

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anon-key")
}

func TestConfig_Validate_BadStrategy(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t, "--anon-key", "test-key", "--strategy", "invalid")
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t, "--anon-key", "test-key")
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_StaircaseNeedsSteps(t *testing.T) {
	t.Parallel()

	cfg := parseTestFlags(t, "--anon-key", "test-key", "--mode", "staircase")
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "steps")
}

func TestConfig_WSURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url  string
		want string
	}{
		{"http://localhost:8000", "ws://localhost:8000"},
		{"https://example.com", "wss://example.com"},
		{"http://host:9000/path", "ws://host:9000/path"},
	}

	for _, tt := range tests {
		cfg := parseTestFlags(t, "--url", tt.url)
		assert.Equal(t, tt.want, cfg.Server.WSURL)
	}
}

func TestConfig_AnonKeyFromEnv(t *testing.T) {
	t.Setenv("ANON_KEY", "env-key-value")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := registerFlags(fs)
	require.NoError(t, fs.Parse(nil))
	cfg.resolve()

	assert.Equal(t, "env-key-value", cfg.Server.AnonKey)
}

// parseTestFlags creates a Config by parsing the given args against a fresh FlagSet.
// Note: clears ANON_KEY so tests get predictable defaults. Tests that need
// ANON_KEY set should use t.Setenv and call registerFlags/resolve directly.
func parseTestFlags(t *testing.T, args ...string) *Config {
	t.Helper()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := registerFlags(fs)
	require.NoError(t, fs.Parse(args))

	// Save and clear ANON_KEY to avoid env leaking into parallel tests.
	savedKey := os.Getenv("ANON_KEY")
	if savedKey != "" {
		os.Unsetenv("ANON_KEY")
		t.Setenv("ANON_KEY", "test-key-from-env")
	}

	cfg.resolve()

	return cfg
}
