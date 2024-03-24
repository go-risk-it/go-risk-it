package rest

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewHandler,
			fx.As(new(Handler)),
		),
	),
)
