package phase

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
		fx.Annotate(
			NewWalker,
			fx.As(new(Walker)),
		),
	),
)
