package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := ParseFlags()
	if err := cfg.ApplyPreset(); err != nil {
		log.Fatal(err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	if cfg.Chaos.Enabled() {
		log.Printf(
			"chaos enabled: disconnect=%.0f%% slow_move=%.0f%% error_move=%.0f%%",
			cfg.Chaos.DisconnectRate*100,
			cfg.Chaos.SlowMoveRate*100,
			cfg.Chaos.ErrorMoveRate*100,
		)
	}

	logPresetInfo(cfg)

	app, err := NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func logPresetInfo(cfg *Config) {
	if cfg.staircaseCfg != nil {
		log.Printf(
			"using staircase preset %q: steps=%v, hold=%v",
			cfg.Preset, cfg.staircaseCfg.Steps, cfg.staircaseCfg.HoldDuration,
		)
	} else if cfg.rampCfg != nil {
		log.Printf(
			"using ramp preset %q: %d games/min, max %d, threshold %.0f%%",
			cfg.Preset, cfg.rampCfg.GamesPerMinute, cfg.rampCfg.MaxGames,
			cfg.rampCfg.ErrorThreshold*100,
		)
	} else if cfg.Preset != "" {
		log.Printf(
			"using preset %q: %d games, %d players, timeout=%v, ramp=%v",
			cfg.Preset, cfg.NumGames, cfg.Game.NumPlayers, cfg.Game.GameTimeout, cfg.RampUp,
		)
	}
}
