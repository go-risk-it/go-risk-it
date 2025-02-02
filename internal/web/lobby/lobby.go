package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	controller.Module,
	fetcher.Module,
	rest.Module,
	signals.Module,
	ws.Module,
)
