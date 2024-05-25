package pool

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

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
		buildConnectionString(config),
	)
	if err != nil {
		panic("Unable to create connection pool")
	}

	log.Infow("created connection pool", "connstr", buildConnectionString(config))

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

	result := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		config.User,
		config.Password,
		hostPort,
		config.Name,
	)

	if config.DisableSSL {
		result += "?sslmode=disable"
	}

	return result
}

func executeMigrations(
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
) error {
	log.Infow("preparing to execute migrations", "connstr", buildConnectionString(config))

	migr, err := migrate.New("file://migrations", buildConnectionString(config))
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
