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
			fx.Annotate(
				assignment.NewAssignmentService,
				fx.As(new(assignment.Service)),
			),
			fx.Annotate(
				board.NewBoardService,
				fx.As(new(board.Service)),
			),
			fx.Annotate(
				region.NewRegionService,
				fx.As(new(region.Service)),
			),
			fx.Annotate(
				player.NewPlayersService,
				fx.As(new(player.Service)),
			),
			fx.Annotate(
				game.NewGameService,
				fx.As(new(game.Service)),
			),
		),
		fx.Invoke(func(engine *nbhttp.Engine) {}),
		//fx.Invoke(func(gs *game.ServiceImpl, di db.DBTX, q *db.Queries) error {
		//	ctx := context.TODO()
		//	// cast to pgxpool.Pool
		//	db := di.(*pgxpool.Pool)
		//	tx, err := db.Begin(ctx)
		//	if err != nil {
		//		return err
		//	}
		//	defer tx.Rollback(ctx)
		//	qtx := q.WithTx(tx)
		//	err = gs.CreateGame(ctx, qtx, &board.Board{
		//		Regions: []board.Region{
		//			{
		//				ExternalReference: 1,
		//				Name:              "Alaska",
		//				ContinentId:       1,
		//			},
		//			{
		//				ExternalReference: 2,
		//				Name:              "Northwest Territory",
		//				ContinentId:       1,
		//			},
		//		},
		//		Continents: []board.Continent{
		//			{
		//				ExternalReference: 1,
		//				Name:              "North America",
		//				BonusTroops:       5,
		//			},
		//		},
		//		Borders: nil,
		//	}, []string{"tom", "fran"})
		//	if err != nil {
		//		panic(err)
		//	}
		//
		//	return tx.Commit(ctx)
		//}),
	).Run()
}
