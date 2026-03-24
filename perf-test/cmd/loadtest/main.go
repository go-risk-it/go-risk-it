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
	if cfg.Run.staircaseCfg != nil {
		log.Printf(
			"using staircase preset %q: steps=%v, hold=%v",
			cfg.Run.Preset, cfg.Run.staircaseCfg.Steps, cfg.Run.staircaseCfg.HoldDuration,
		)
	} else if cfg.Run.rampCfg != nil {
		log.Printf(
			"using ramp preset %q: %d games/min, max %d, threshold %.0f%%",
			cfg.Run.Preset, cfg.Run.rampCfg.GamesPerMinute, cfg.Run.rampCfg.MaxGames,
			cfg.Run.rampCfg.ErrorThreshold*100,
		)
	} else if cfg.Run.Preset != "" {
		log.Printf(
			"using preset %q: %d games, %d players, timeout=%v, ramp=%v",
			cfg.Run.Preset, cfg.Run.Batch.NumGames, cfg.Game.NumPlayers,
			cfg.Game.GameTimeout, cfg.Run.Batch.RampUp,
		)
	}
}
