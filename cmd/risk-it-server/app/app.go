package app

import (
	"github.com/go-risk-it/go-risk-it/internal/game"
	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/go-risk-it/go-risk-it/internal/kernel/logger"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	riskslog "github.com/go-risk-it/go-risk-it/internal/kernel/slog"
	"github.com/go-risk-it/go-risk-it/internal/lobby"
	"github.com/go-risk-it/go-risk-it/internal/web"
	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/fx"
)

var Module = fx.Options(
	config.Module,
	riskslog.Module,
	logger.Module,
	metrics.Module,
	web.Module,
	game.Module,
	lobby.Module,
	fx.Invoke(func(engine *nbhttp.Engine) {}),
)
