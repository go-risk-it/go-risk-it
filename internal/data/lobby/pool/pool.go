package pool

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/data/migration"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func executeMigrations(
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
) error {
	return migration.Execute(log, config, "lobby")
}

type Transaction interface {
	pgx.Tx
}

type DB interface {
	sqlc.DBTX
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewConnectionPool,
			fx.As(new(sqlc.DBTX)),
			fx.As(new(DB)),
		),
	),
	fx.Invoke(
		executeMigrations,
	),
)

func NewConnectionPool(
	lifecycle fx.Lifecycle,
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
) *pgxpool.Pool {
	pool, err := pgxpool.New(
		context.Background(),
		config.BuildConnectionString("lobby"),
	)
	if err != nil {
		panic("Unable to create connection pool")
	}

	log.Infow("created connection pool", "schema", "lobby")

	lifecycle.Append(
		fx.Hook{
			OnStart: nil,
			OnStop: func(ctx context.Context) error {
				pool.Close()

				return nil
			},
		},
	)

	return pool
}
