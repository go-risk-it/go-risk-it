package logging

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"log"
)

func NewLogger(lc fx.Lifecycle) *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("can't initialize zap logger: %v", err)
	}
	lc.Append(fx.Hook{OnStop: func(ctx context.Context) error {
		defer logger.Sync()
		return nil
	}})
	return logger.Sugar()
}
