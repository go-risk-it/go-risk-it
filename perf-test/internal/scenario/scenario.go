package scenario

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

var presets = map[string]orchestrator.Config{
	"smoke": {
		NumGames:    1,
		NumPlayers:  3,
		GameTimeout: 5 * time.Minute,
		RampUp:      0,
	},
	"light": {
		NumGames:    10,
		NumPlayers:  4,
		GameTimeout: 10 * time.Minute,
		RampUp:      30 * time.Second,
	},
	"medium": {
		NumGames:    50,
		NumPlayers:  4,
		GameTimeout: 15 * time.Minute,
		RampUp:      60 * time.Second,
	},
	"heavy": {
		NumGames:    100,
		NumPlayers:  4,
		GameTimeout: 20 * time.Minute,
		RampUp:      120 * time.Second,
	},
	"soak": {
		NumGames:    20,
		NumPlayers:  4,
		GameTimeout: 60 * time.Minute,
		RampUp:      60 * time.Second,
	},
}

// Get returns the config for a named preset.
func Get(name string) (orchestrator.Config, error) {
	cfg, ok := presets[name]
	if !ok {
		return orchestrator.Config{}, fmt.Errorf("unknown preset %q, available: %v", name, List())
	}

	return cfg, nil
}

// List returns sorted preset names.
func List() []string {
	names := make([]string, 0, len(presets))
	for k := range presets {
		names = append(names, k)
	}

	sort.Strings(names)

	return names
}
