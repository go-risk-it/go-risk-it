package ws

import (
	"github.com/tomfran/go-risk-it/internal/web/ws/connection/manager"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection/upgrader"
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
			manager.NewManager,
			fx.As(new(manager.Manager)),
			fx.ParamTags(`group:"fetchers"`),
		),
		fx.Annotate(
			upgrader.New,
			fx.As(new(upgrader.Upgrader)),
		),
	),
)
