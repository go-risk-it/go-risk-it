package nbio

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewEngine(lc fx.Lifecycle, config nbhttp.Config, log *zap.SugaredLogger) *nbhttp.Engine {
	engine := nbhttp.NewEngine(config)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Starting engine...")
				err := engine.Start()
				if err != nil {
					panic("nbio.Start failed")
				}
				log.Info("Engine started!")

				interrupt := make(chan os.Signal, 1)
				signal.Notify(interrupt, os.Interrupt)
				<-interrupt

				_, cancel := context.WithTimeout(ctx, time.Second*3)
				defer cancel()
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := engine.Shutdown(ctx)
			if err != nil {
				return fmt.Errorf("failure during engine shutdown: %w", err)
			}

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
