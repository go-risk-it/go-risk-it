package main

import (
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game"
	"github.com/tomfran/go-risk-it/internal/loggerfx"
	"github.com/tomfran/go-risk-it/internal/web"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		loggerfx.Module,
		game.Module,
		db.Module,
		web.Module,
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
