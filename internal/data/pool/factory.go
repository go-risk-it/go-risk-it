package pool

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewConnectionPool creates a pgxpool.Pool, optionally executes initSQL statements,
// and registers a shutdown hook.
func NewConnectionPool(
	lifecycle fx.Lifecycle,
	log *zap.SugaredLogger,
	cfg config.DatabaseConfig,
	schema string,
	initSQL ...string,
) (*pgxpool.Pool, error) {
	ctx := context.Background()

	connectionPool, err := pgxpool.New(ctx, cfg.BuildConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to create %s connection pool: %w", schema, err)
	}

	for _, sql := range initSQL {
		if _, err := connectionPool.Exec(ctx, sql); err != nil {
			connectionPool.Close()

			return nil, fmt.Errorf("unable to execute init SQL for %s: %w", schema, err)
		}
	}

	log.Infow("created connection pool", "schema", schema)

	lifecycle.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				connectionPool.Close()

				return nil
			},
		},
	)

	return connectionPool, nil
}
