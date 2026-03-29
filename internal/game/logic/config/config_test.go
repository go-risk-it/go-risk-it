package config_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/logic/config"
	"github.com/knadh/koanf/parsers/yaml"
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
