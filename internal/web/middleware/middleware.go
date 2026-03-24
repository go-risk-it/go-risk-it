package middleware

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewAuthMiddleware,
		NewCorsMiddleware,
		NewLogMiddleware,
		NewOTelMiddleware,
	),
)
