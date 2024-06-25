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
		fx.Annotate(
			NewAuthMiddleware,
			fx.As(new(AuthMiddleware)),
		),
		fx.Annotate(
			NewWebsocketAuthMiddleware,
			fx.As(new(WebsocketHeaderConversionMiddleware)),
		),
		fx.Annotate(
			NewGameMiddleware,
			fx.As(new(GameMiddleware)),
		),
		fx.Annotate(
			NewLogMiddleware,
			fx.As(new(LogMiddleware)),
		),
	),
)
