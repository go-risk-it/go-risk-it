package pool

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/data/migration"
	poolfactory "github.com/go-risk-it/go-risk-it/internal/data/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func executeMigrations(
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
) error {
	return migration.Execute(log, config, "game")
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
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {
	return poolfactory.NewConnectionPool(lifecycle, log, cfg, "game")
}
