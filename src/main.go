package main

import (
	"github.com/lesismal/nbio/nbhttp"
	"go-risk-it/handlers"
	"go-risk-it/logging"
	"go-risk-it/nbio"
	"go-risk-it/ws"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			ws.NewUpgrader,
			nbio.NewServeMux,
			nbio.NewNbioConfig,
			nbio.NewEngine,
			handlers.NewWebSocketHandler,
			logging.NewLogger,
		),
		fx.Invoke(func(engine *nbhttp.Engine) {}),
	).Run()
}
