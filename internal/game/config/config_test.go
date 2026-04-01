package config_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/config"
	kernelconfig "github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/knadh/koanf/parsers/yaml"
	env "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameConfig_ComponentTestValues(t *testing.T) {
	t.Parallel()

	raw := []byte(`
game:
  dice:
    roll_strategy: attacker_always_wins
  regionassignment:
    assignment_strategy: sequential
  history:
    size: 999
`)

	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider(raw), yaml.Parser()))

	result, err := config.NewGameConfig(k)
	require.NoError(t, err)

	assert.Equal(t, "attacker_always_wins", result.DiceConfig.RollStrategy)
	assert.Equal(t, "sequential", result.RegionassignmentConfig.AssignmentStrategy)
	assert.Equal(t, int64(999), result.HistoryConfig.Size)
}

func TestGameConfig_ProdValues(t *testing.T) {
	t.Parallel()

	raw := []byte(`
game:
  dice:
    roll_strategy: random
  regionassignment:
    assignment_strategy: random
  history:
    size: 50
`)

	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider(raw), yaml.Parser()))

	result, err := config.NewGameConfig(k)
	require.NoError(t, err)

	assert.Equal(t, "random", result.DiceConfig.RollStrategy)
	assert.Equal(t, "random", result.RegionassignmentConfig.AssignmentStrategy)
	assert.Equal(t, int64(50), result.HistoryConfig.Size)
}

func TestGameConfig_MissingGameSection(t *testing.T) {
	t.Parallel()

	raw := []byte(`
database:
  host: localhost
`)

	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider(raw), yaml.Parser()))

	result, err := config.NewGameConfig(k)
	require.NoError(t, err)

	// Zero values when game section is absent
	assert.Empty(t, result.DiceConfig.RollStrategy)
	assert.Empty(t, result.RegionassignmentConfig.AssignmentStrategy)
	assert.Equal(t, int64(0), result.HistoryConfig.Size)
}

func TestGameConfig_EnvCannotOverrideMultiWordKeys(t *testing.T) {
	t.Parallel()

	raw := []byte(`
game:
  dice:
    roll_strategy: yaml-strategy
  history:
    size: 42
`)

	koanfManager := koanf.New(".")
	require.NoError(t, koanfManager.Load(rawbytes.Provider(raw), yaml.Parser()))
	require.NoError(t, koanfManager.Load(env.Provider(".", env.Opt{
		TransformFunc: kernelconfig.TransformKey,
		EnvironFunc: func() []string {
			// GAME_DICE_ROLL_STRATEGY maps to "game.dice.roll.strategy" (extra dot)
			// which does NOT match the YAML key "game.dice.roll_strategy".
			// This documents the intentional design: game config values come from
			// YAML files, not env vars.
			return []string{
				"GAME_DICE_ROLL_STRATEGY=env-strategy",
				// Single-word keys like "game.history.size" CAN be overridden.
				"GAME_HISTORY_SIZE=99",
			}
		},
	}), nil))

	result, err := config.NewGameConfig(koanfManager)
	require.NoError(t, err)

	assert.Equal(t, "yaml-strategy", result.DiceConfig.RollStrategy,
		"multi-word key should keep YAML value — env var lands on wrong koanf path")
	assert.Equal(t, int64(99), result.HistoryConfig.Size,
		"single-word key should be overridden by env var")
}
