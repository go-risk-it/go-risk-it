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
			NewCorsMiddleware,
			fx.As(new(CorsMiddleware)),
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
			NewLobbyMiddleware,
			fx.As(new(LobbyMiddleware)),
		),
		fx.Annotate(
			NewLogMiddleware,
			fx.As(new(LogMiddleware)),
		),
		fx.Annotate(
			NewOTelMiddleware,
			fx.As(new(OTelMiddleware)),
		),
	),
)
