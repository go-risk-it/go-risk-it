package main

import (
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/players"
	"github.com/tomfran/go-risk-it/internal/handlers"
	"github.com/tomfran/go-risk-it/internal/logging"
	"github.com/tomfran/go-risk-it/internal/nbio"
	"github.com/tomfran/go-risk-it/internal/ws"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			fx.Annotate(
				db.NewConnectionPool,
				fx.As(new(db.DBTX)),
			),
			db.New,
			ws.NewUpgrader,
			nbio.NewServeMux,
			nbio.NewNbioConfig,
			nbio.NewEngine,
			handlers.NewWebSocketHandler,
			logging.NewLogger,
			players.NewPlayersService,
		),
		fx.Invoke(func(engine *nbhttp.Engine) {}),
	).Run()
}
