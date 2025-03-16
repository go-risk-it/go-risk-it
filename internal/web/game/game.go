package game

import (
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	controller.Module,
	fetcher.Module,
	rest.Module,
	signals.Module,
	ws.Module,
)
