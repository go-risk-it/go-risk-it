package ws

import (
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			connection.NewHandler,
			fx.As(new(connection.Handler)),
		),
		fx.Annotate(
			connection.NewManager,
			fx.As(new(connection.Manager)),
			fx.ParamTags(`group:"fetchers"`),
		),
		fx.Annotate(
			connection.New,
			fx.As(new(connection.Upgrader)),
		),
	),
)
