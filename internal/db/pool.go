package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
)

func NewConnectionPool(lc fx.Lifecycle, logger *zap.SugaredLogger) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), "postgresql://localhost:5432/risk-it?user=postgres&password=password")
	if err != nil {
		logger.Fatal(os.Stderr, "Unable to create connection pool: %v\n", err)
		panic("Unable to create connection pool")
	}
	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				pool.Close()
				return nil
			},
		},
	)
	return pool
}
