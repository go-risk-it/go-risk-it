package ws

import (
	"github.com/tomfran/go-risk-it/internal/web/ws/handler"
	"github.com/tomfran/go-risk-it/internal/web/ws/upgrader"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			handler.New,
			fx.As(new(handler.MessageHandler)),
		),
		upgrader.New,
	),
)
