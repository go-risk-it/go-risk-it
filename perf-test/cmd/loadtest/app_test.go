package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp_MinimalConfig(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.Server.MapFile = "../../map.json"
	cfg.Game.Strategy = "heuristic"

	app, err := NewApp(cfg)
	require.NoError(t, err)
	assert.NotNil(t, app.graph)
	assert.NotNil(t, app.strategy)
	assert.Nil(t, app.otel)
	assert.Nil(t, app.dbStats)

	app.Close()
}

func TestNewApp_InvalidMap(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.Server.MapFile = "nonexistent.json"
	cfg.Game.Strategy = "heuristic"

	_, err := NewApp(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load map")
}

func TestNewApp_AllStrategies(t *testing.T) {
	t.Parallel()

	strategies := []string{"heuristic", "beginner", "normal", "expert"}
	for _, s := range strategies {
		t.Run(s, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{}
			cfg.Server.MapFile = "../../map.json"
			cfg.Game.Strategy = s

			app, err := NewApp(cfg)
			require.NoError(t, err)
			assert.NotNil(t, app.strategy)

			app.Close()
		})
	}
}

func TestNewApp_ChaosConfig(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.Server.MapFile = "../../map.json"
	cfg.Game.Strategy = "heuristic"
	cfg.Chaos.DisconnectRate = 0.10
	cfg.Chaos.SlowMoveRate = 0.05
	cfg.Chaos.SlowMoveDelay = 1 * time.Second
	cfg.Chaos.ReconnectDelay = 2 * time.Second

	// Chaos config doesn't break App construction.
	// Actual chaos wrapping happens in per-mode methods where collectors exist.
	app, err := NewApp(cfg)
	require.NoError(t, err)
	assert.NotNil(t, app.strategy)

	app.Close()
}

func TestApp_Close_NilOptionals(t *testing.T) {
	t.Parallel()

	// App with nil otel and dbStats should close without panic.
	app := &App{}
	assert.NotPanics(t, func() {
		app.Close()
	})
}
