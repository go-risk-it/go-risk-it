package main

import (
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/handlers"
	"github.com/tomfran/go-risk-it/internal/logging"
	"github.com/tomfran/go-risk-it/internal/nbio"
	"github.com/tomfran/go-risk-it/internal/ws"
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
