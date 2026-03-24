package slog

import (
	stdslog "log/slog"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"go.uber.org/fx"
)

func NewLogger() *stdslog.Logger {
	var level stdslog.Level
	if os.Getenv(config.EnvironmentKey) == "prod" {
		level = stdslog.LevelInfo
	} else {
		level = stdslog.LevelDebug
	}

	inner := stdslog.NewJSONHandler(os.Stderr, &stdslog.HandlerOptions{
		Level: level,
	})
	handler := NewContextHandler(inner)

	logger := stdslog.New(handler)
	stdslog.SetDefault(logger)

	return logger
}

var Module = fx.Options(
	fx.Provide(NewLogger),
)
