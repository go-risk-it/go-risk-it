package logging

import (
	"context"
	"log"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger(lc fx.Lifecycle) *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("can't initialize zap logger: %v", err)
	}
	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				if err := logger.Sync(); err != nil {
					return err
				}
				return nil
			},
		},
	)
	return logger.Sugar()
}
