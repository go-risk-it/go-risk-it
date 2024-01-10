package main

import (
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/game"
	"github.com/tomfran/go-risk-it/internal/game/player"
	"github.com/tomfran/go-risk-it/internal/game/region"
	"github.com/tomfran/go-risk-it/internal/game/region/assignment"
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
			player.NewPlayersService,
			assignment.NewAssignmentService,
			board.NewBoardService,
			region.NewRegionService,
			game.NewGameService,
		),
		fx.Invoke(func(engine *nbhttp.Engine) {}),
		fx.Invoke(func(gs *game.Service) {
			err := gs.CreateGame(&board.Board{
				Regions: []board.Region{
					{
						ExternalReference: 1,
						Name:              "Alaska",
						ContinentId:       1,
					},
					{
						ExternalReference: 2,
						Name:              "Northwest Territory",
						ContinentId:       1,
					},
				},
				Continents: []board.Continent{
					{
						ExternalReference: 1,
						Name:              "North America",
						BonusTroops:       5,
					},
				},
				Borders: nil,
			}, []string{"tom", "fran"})
			if err != nil {
				panic(err)
			}
		}),
	).Run()
}
