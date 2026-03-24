package middleware

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

type Middleware interface {
	Wrap(routeToWrap route.Route) route.Route
}

var Module = fx.Options(
	fx.Provide(
		NewAuthMiddleware,
		NewCorsMiddleware,
		NewWebsocketAuthMiddleware,
		NewGameMiddleware,
		NewLobbyMiddleware,
		NewLogMiddleware,
		NewOTelMiddleware,
	),
)
