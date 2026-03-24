package game

import (
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/data/migration"
	poolfactory "github.com/go-risk-it/go-risk-it/internal/data/pool"
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

func NewConnectionPool(
	lifecycle fx.Lifecycle,
	log *zap.SugaredLogger,
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {
	return poolfactory.NewConnectionPool(lifecycle, log, cfg, "game")
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
