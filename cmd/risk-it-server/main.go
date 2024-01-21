package main

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/loggerfx"
	"github.com/tomfran/go-risk-it/internal/logic"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/web"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		loggerfx.Module,
		logic.Module,
		db.Module,
		web.Module,
		fx.Invoke(func(engine *nbhttp.Engine) {}),
		fx.Invoke(func(gs game.Service, di db.DB, q *db.Queries, log *zap.SugaredLogger) error {
			ctx := context.TODO()
			dbPool := di.(*pgxpool.Pool)
			tx, err := dbPool.Begin(ctx)
			if err != nil {
				return err
			}
			defer func(tx pgx.Tx, ctx context.Context) {
				err := tx.Rollback(ctx)
				if err != nil {
					log.Info(err)
				}
			}(tx, ctx)
			qtx := q.WithTx(tx)
			err = gs.CreateGame(ctx, qtx, &board.Board{
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

			return tx.Commit(ctx)
		}),
	).Run()
}
