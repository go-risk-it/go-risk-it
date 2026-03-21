package scenario

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

// Scenario bundles an orchestrator config with optional chaos injection config.
type Scenario struct {
	Config      orchestrator.Config
	ChaosConfig chaos.Config
}

var presets = map[string]Scenario{
	"smoke": {
		Config: orchestrator.Config{
			NumGames:    1,
			NumPlayers:  3,
			GameTimeout: 5 * time.Minute,
			RampUp:      0,
		},
	},
	"light": {
		Config: orchestrator.Config{
			NumGames:    10,
			NumPlayers:  4,
			GameTimeout: 10 * time.Minute,
			RampUp:      30 * time.Second,
		},
	},
	"medium": {
		Config: orchestrator.Config{
			NumGames:    50,
			NumPlayers:  4,
			GameTimeout: 15 * time.Minute,
			RampUp:      60 * time.Second,
		},
	},
	"heavy": {
		Config: orchestrator.Config{
			NumGames:    100,
			NumPlayers:  4,
			GameTimeout: 20 * time.Minute,
			RampUp:      120 * time.Second,
		},
	},
	"soak": {
		Config: orchestrator.Config{
			NumGames:    20,
			NumPlayers:  4,
			GameTimeout: 60 * time.Minute,
			RampUp:      60 * time.Second,
		},
	},
	"chaos-light": {
		Config: orchestrator.Config{
			NumGames:    5,
			NumPlayers:  4,
			GameTimeout: 10 * time.Minute,
			RampUp:      0,
		},
		ChaosConfig: chaos.Config{
			DisconnectRate: 0.05,
			SlowMoveRate:   0.10,
			SlowMoveDelay:  1 * time.Second,
			ErrorMoveRate:  0.02,
			ReconnectDelay: 2 * time.Second,
		},
	},
	"chaos-heavy": {
		Config: orchestrator.Config{
			NumGames:    20,
			NumPlayers:  4,
			GameTimeout: 15 * time.Minute,
			RampUp:      0,
		},
		ChaosConfig: chaos.Config{
			DisconnectRate: 0.15,
			SlowMoveRate:   0.20,
			SlowMoveDelay:  3 * time.Second,
			ErrorMoveRate:  0.05,
			ReconnectDelay: 5 * time.Second,
		},
	},
}

// Get returns the scenario for a named preset.
func Get(name string) (Scenario, error) {
	s, ok := presets[name]
	if !ok {
		return Scenario{}, fmt.Errorf("unknown preset %q, available: %v", name, List())
	}

	return s, nil
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
