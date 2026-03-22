package nbio

import (
	"context"
	"fmt"

	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewEngine(lc fx.Lifecycle, config nbhttp.Config, log *zap.SugaredLogger) *nbhttp.Engine {
	engine := nbhttp.NewEngine(config)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting engine...")

			err := engine.Start()
			if err != nil {
				return fmt.Errorf("nbio.Start failed: %w", err)
			}

			log.Info("Engine started!")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down engine...")

			err := engine.Shutdown(ctx)
			if err != nil {
				return fmt.Errorf("failure during engine shutdown: %w", err)
			}

			log.Info("Engine shut down successfully")

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
