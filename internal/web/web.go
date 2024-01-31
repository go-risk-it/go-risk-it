package web

import (
	"github.com/tomfran/go-risk-it/internal/web/controllers"
	"github.com/tomfran/go-risk-it/internal/web/handlers"
	"github.com/tomfran/go-risk-it/internal/web/nbio"
	"github.com/tomfran/go-risk-it/internal/web/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	nbio.Module,
	controllers.Module,
	ws.Module,
	fx.Provide(
		NewServeMux,
		handlers.NewWebSocketHandler,
	),
)
