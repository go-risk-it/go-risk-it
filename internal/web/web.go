package web

import (
	"github.com/tomfran/go-risk-it/internal/web/handlers"
	"github.com/tomfran/go-risk-it/internal/web/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ws.NewUpgrader,
		NewServeMux,
		handlers.NewWebSocketHandler,
	),
)
