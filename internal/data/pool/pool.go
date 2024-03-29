package pool

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomfran/go-risk-it/internal/config"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
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
		buildConnectionString(config),
	)
	if err != nil {
		log.Fatal(os.Stderr, "Unable to create connection pool: %v\n", err)
		panic("Unable to create connection pool")
	}

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

func buildConnectionString(config config.DatabaseConfig) string {
	hostPort := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))

	return fmt.Sprintf(
		"postgresql://%s/%s?user=%s&password=%s",
		hostPort,
		config.Name,
		config.User,
		config.Password,
	)
}

type Transaction interface {
	pgx.Tx
}

type DB interface {
	sqlc.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewConnectionPool,
			fx.As(new(sqlc.DBTX)),
			fx.As(new(DB)),
		),
	),
)
