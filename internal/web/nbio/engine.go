package nbio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/fx"
)

func NewEngine(lc fx.Lifecycle, config nbhttp.Config) *nbhttp.Engine {
	engine := nbhttp.NewEngine(config)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			slog.Info("Starting engine...")

			err := engine.Start()
			if err != nil {
				return fmt.Errorf("nbio.Start failed: %w", err)
			}

			slog.Info("Engine started!")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Shutting down engine...")

			err := engine.Shutdown(ctx)
			if err != nil {
				return fmt.Errorf("failure during engine shutdown: %w", err)
			}

			slog.Info("Engine shut down successfully")

			return nil
		},
	})

	return engine
}

var Module = fx.Options(
	fx.Provide(
		newNbioConfig,
		NewEngine,
	),
)
