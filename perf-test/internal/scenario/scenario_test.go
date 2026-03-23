package scenario_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_StaircasePreset(t *testing.T) {
	t.Parallel()

	s, err := scenario.Get("staircase")
	require.NoError(t, err)
	require.NotNil(t, s.StaircaseConfig)
	assert.Nil(t, s.RampConfig)
	assert.Equal(t, []int{5, 10, 20, 40, 60, 80}, s.StaircaseConfig.Steps)
	assert.Equal(t, 4, s.StaircaseConfig.NumPlayers)
	assert.True(t, s.StaircaseConfig.StopOnBreach)
}

func TestGet_StaircaseLightPreset(t *testing.T) {
	t.Parallel()

	s, err := scenario.Get("staircase-light")
	require.NoError(t, err)
	require.NotNil(t, s.StaircaseConfig)
	assert.Equal(t, []int{2, 5, 10, 20}, s.StaircaseConfig.Steps)
}

func TestGet_StaircaseHeavyPreset(t *testing.T) {
	t.Parallel()

	s, err := scenario.Get("staircase-heavy")
	require.NoError(t, err)
	require.NotNil(t, s.StaircaseConfig)
	assert.Equal(t, []int{10, 20, 40, 60, 80, 100, 120}, s.StaircaseConfig.Steps)
}

func TestGet_ExistingPresetsUnchanged(t *testing.T) {
	t.Parallel()

	s, err := scenario.Get("smoke")
	require.NoError(t, err)
	assert.Equal(t, 1, s.Config.NumGames)
	assert.Nil(t, s.StaircaseConfig)

	s, err = scenario.Get("light")
	require.NoError(t, err)
	assert.Equal(t, 10, s.Config.NumGames)
}

func TestList_IncludesStaircasePresets(t *testing.T) {
	t.Parallel()

	names := scenario.List()
	assert.Contains(t, names, "staircase")
	assert.Contains(t, names, "staircase-light")
	assert.Contains(t, names, "staircase-heavy")
}
