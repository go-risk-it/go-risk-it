package app

import (
	"github.com/go-risk-it/go-risk-it/internal/data"
	"github.com/go-risk-it/go-risk-it/internal/game"
	"github.com/go-risk-it/go-risk-it/internal/game/rand"
	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	riskslog "github.com/go-risk-it/go-risk-it/internal/kernel/slog"
	"github.com/go-risk-it/go-risk-it/internal/lobby"
	lobbydata "github.com/go-risk-it/go-risk-it/internal/lobby/data"
	lobbylogic "github.com/go-risk-it/go-risk-it/internal/lobby/logic"
	"github.com/go-risk-it/go-risk-it/internal/logic"
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
	lobbylogic.Module,
	lobbydata.Module,
	web.Module,
	game.Module,
	lobby.Module,
	rand.Module,
	fx.Invoke(func(engine *nbhttp.Engine) {}),
)
