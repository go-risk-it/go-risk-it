package main

import (
	"context"
	"errors"
	"fmt"

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

var (
	errPoolCast = errors.New("cannot cast db pool")
	gameBoard   = &board.Board{
		Regions: []board.Region{
			{
				ExternalReference: 1,
				Name:              "Alaska",
				ContinentID:       1,
			},
			{
				ExternalReference: 2,
				Name:              "Northwest Territory",
				ContinentID:       1,
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
	}
)

func main() {
	fx.New(
		loggerfx.Module,
		logic.Module,
		db.Module,
		web.Module,
		fx.Invoke(func(engine *nbhttp.Engine) {}),
		fx.Invoke(
			func(service game.Service, di db.DB, que *db.Queries, log *zap.SugaredLogger) error {
				ctx := context.TODO()
				dbPool, ok := di.(*pgxpool.Pool)
				if !ok {
					return errPoolCast
				}

				transaction, err := dbPool.Begin(ctx)
				if err != nil {
					return fmt.Errorf("transaction begin failed: %w", err)
				}

				defer func(tx pgx.Tx, ctx context.Context) {
					err := tx.Rollback(ctx)
					if err != nil {
						log.Info(err)
					}
				}(transaction, ctx)

				qtx := que.WithTx(transaction)
				err = service.CreateGame(ctx, qtx, gameBoard, []string{"tom", "fran"})
				if err != nil {
					return fmt.Errorf("failed to create game: %w", err)
				}

				err = transaction.Commit(ctx)
				if err != nil {
					return fmt.Errorf("transaction commit failed: %w", err)
				}

				return nil
			},
		),
	).Run()
}
