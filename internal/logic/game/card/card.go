package card

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			New,
			fx.As(new(Service)),
		),
	),
)
