package loggerfx

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger(lifecycle fx.Lifecycle) *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	lifecycle.Append(
		fx.Hook{
			OnStart: nil,
			OnStop: func(ctx context.Context) error {
				if err := logger.Sync(); err != nil {
					return fmt.Errorf("failed to sync logger: %w", err)
				}

				return nil
			},
		},
	)

	return logger.Sugar()
}

// Module provided to fx.
var Module = fx.Options(
	fx.Provide(NewLogger),
)
