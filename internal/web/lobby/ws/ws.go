package ws

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewManager,
			fx.As(new(Manager)),
		),
	),
)
