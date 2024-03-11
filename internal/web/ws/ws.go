package ws

import (
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			message.NewHandler,
			fx.As(new(message.Handler)),
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
