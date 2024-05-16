package middleware

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewAuthMiddleware,
			fx.As(new(AuthMiddleware)),
		),
	),
)
