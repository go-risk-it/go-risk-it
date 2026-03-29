package slog

import (
	stdslog "log/slog"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
)

func NewLogger(loggerProvider *sdklog.LoggerProvider) *stdslog.Logger {
	var level stdslog.Level
	if os.Getenv(config.EnvironmentKey) == "prod" {
		level = stdslog.LevelInfo
	} else {
		level = stdslog.LevelDebug
	}

	inner := otelslog.NewHandler(
		"go-risk-it",
		otelslog.WithLoggerProvider(loggerProvider),
	)

	handler := NewContextHandler(inner, level)

	logger := stdslog.New(handler)
	stdslog.SetDefault(logger)

	return logger
}

var Module = fx.Options(
	fx.Provide(NewLogger),
)
