package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
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
		observe.Info(context.Background(), "chaos enabled",
			attribute.Float64("disconnect_rate_pct", cfg.Chaos.DisconnectRate*100),
			attribute.Float64("slow_move_rate_pct", cfg.Chaos.SlowMoveRate*100),
			attribute.Float64("error_move_rate_pct", cfg.Chaos.ErrorMoveRate*100),
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
		log.Fatal(err) //nolint:gocritic // defers cleaned up by process exit
	}
}

func logPresetInfo(cfg *Config) {
	ctx := context.Background()

	switch {
	case cfg.Run.staircaseCfg != nil:
		observe.Info(ctx, "using staircase preset",
			attribute.String("preset", cfg.Run.Preset),
			attribute.String("steps", fmt.Sprintf("%v", cfg.Run.staircaseCfg.Steps)),
			attribute.String("hold", cfg.Run.staircaseCfg.HoldDuration.String()),
		)
	case cfg.Run.rampCfg != nil:
		observe.Info(ctx, "using ramp preset",
			attribute.String("preset", cfg.Run.Preset),
			attribute.Int("games_per_min", cfg.Run.rampCfg.GamesPerMinute),
			attribute.Int("max_games", cfg.Run.rampCfg.MaxGames),
			attribute.Float64("error_threshold_pct", cfg.Run.rampCfg.ErrorThreshold*100),
		)
	case cfg.Run.Preset != "":
		observe.Info(ctx, "using preset",
			attribute.String("preset", cfg.Run.Preset),
			attribute.Int("num_games", cfg.Run.Batch.NumGames),
			attribute.Int("numPlayers", cfg.Game.NumPlayers),
			attribute.String("timeout", cfg.Game.GameTimeout.String()),
			attribute.String("ramp_up", cfg.Run.Batch.RampUp.String()),
		)
	}
}
