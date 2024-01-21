package web

import (
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/web/handlers"
	"github.com/tomfran/go-risk-it/internal/web/nbio"
	"github.com/tomfran/go-risk-it/internal/web/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	nbio.Module,
	fx.Provide(
		ws.NewUpgrader,
		NewServeMux,
		handlers.NewWebSocketHandler,
	),
	fx.Invoke(func(engine *nbhttp.Engine) {}),
)
