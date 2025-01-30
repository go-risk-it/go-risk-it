package ws

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			New,
			fx.As(new(Upgrader)),
		),
	),
)
