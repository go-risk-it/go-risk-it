package game

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/go-risk-it/go-risk-it/internal/kernel/data/migration"
	poolfactory "github.com/go-risk-it/go-risk-it/internal/kernel/data/pool"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func executeMigrations(
	config config.DatabaseConfig,
) error {
	return migration.Execute(config, "game")
}

func NewConnectionPool(
	lifecycle fx.Lifecycle,
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {
	return poolfactory.NewConnectionPool(lifecycle, cfg, "game")
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewConnectionPool,
			fx.As(new(sqlc.DBTX)),
			fx.As(new(db.DB)),
		),
	),
	fx.Invoke(
		executeMigrations,
	),
	fx.Provide(
		db.New,
	),
)
