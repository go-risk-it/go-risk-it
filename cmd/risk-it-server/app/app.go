package app

import (
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data"
	"github.com/go-risk-it/go-risk-it/internal/logic"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/rand"
	riskslog "github.com/go-risk-it/go-risk-it/internal/slog"
	"github.com/go-risk-it/go-risk-it/internal/web"
	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/fx"
)

var Module = fx.Options(
	config.Module,
	riskslog.Module,
	metrics.Module,
	logic.Module,
	data.Module,
	web.Module,
	rand.Module,
	fx.Invoke(func(engine *nbhttp.Engine) {}),
)
