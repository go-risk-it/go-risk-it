package config

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
)

// GameConfig groups all game-specific configuration sections.
type GameConfig struct {
	Dice             DiceConfig             `koanf:"dice"`
	Regionassignment RegionassignmentConfig `koanf:"regionassignment"`
	History          HistoryConfig          `koanf:"history"`
}

// Result provides individual config types as fx dependencies.
type Result struct {
	fx.Out

	DiceConfig             DiceConfig
	RegionassignmentConfig RegionassignmentConfig
	HistoryConfig          HistoryConfig
}

// NewGameConfig loads game configuration from the "game" section of the
// koanf instance provided by kernel/config.
func NewGameConfig(k *koanf.Koanf) (Result, error) {
	var cfg GameConfig
	if err := k.Unmarshal("game", &cfg); err != nil {
		return Result{}, fmt.Errorf("failed to unmarshal game config: %w", err)
	}

	return Result{
		DiceConfig:             cfg.Dice,
		RegionassignmentConfig: cfg.Regionassignment,
		HistoryConfig:          cfg.History,
	}, nil
}

// Module provides game configuration types via fx.
var Module = fx.Options(
	fx.Provide(NewGameConfig),
)
