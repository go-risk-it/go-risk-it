package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/go-risk-it/go-risk-it/internal/kernel/data/migration"
	poolfactory "github.com/go-risk-it/go-risk-it/internal/kernel/data/pool"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func executeMigrations(
	config config.DatabaseConfig,
) error {
	return migration.Execute(config, "lobby")
}

func NewConnectionPool(
	lifecycle fx.Lifecycle,
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {
	return poolfactory.NewConnectionPool(lifecycle, cfg, "lobby", "SET search_path TO lobby;")
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
