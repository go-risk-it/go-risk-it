package gamestate

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
	),
)
