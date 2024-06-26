package pool

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewConnectionPool(
	lifecycle fx.Lifecycle,
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
) *pgxpool.Pool {
	pool, err := pgxpool.New(
		context.Background(),
		config.BuildConnectionString(),
	)
	if err != nil {
		panic("Unable to create connection pool")
	}

	log.Infow("created connection pool")

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

func executeMigrations(
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
) error {
	log.Infow("preparing to execute migrations")

	migr, err := migrate.New("file://migrations", config.BuildConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to DB for migrations: %w", err)
	}

	log.Infow("executing migrations")

	if err := migr.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Infow("successfully ran migrations")

	return nil
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
	fx.Invoke(executeMigrations),
)
