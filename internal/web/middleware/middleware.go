package middleware

import (
	"go.uber.org/fx"
)

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
