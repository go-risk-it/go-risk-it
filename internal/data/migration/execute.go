package migration

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func Execute(
	log *zap.SugaredLogger,
	config config.DatabaseConfig,
	schema string,
) error {
	log.Infow("preparing to execute migrations")

	connStr := config.BuildConnectionString(schema)
	if err := createSchema(log, connStr, schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	migr, err := migrate.New("file://migrations/"+schema, connStr)
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

func createSchema(log *zap.SugaredLogger, connStr, schema string) error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to DB for schema creation: %w", err)
	}

	defer func(conn *pgx.Conn, ctx context.Context) {
		if err := conn.Close(ctx); err != nil {
			log.Errorw("failed to close connection", "error", err)
		}
	}(conn, ctx)

	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schema, err)
	}

	return nil
}
