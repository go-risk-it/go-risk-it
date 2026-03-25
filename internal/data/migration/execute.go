package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

func Execute(
	config config.DatabaseConfig,
	schema string,
) error {
	slog.Info("preparing to execute migrations", "schema", schema)

	connStr := config.BuildConnectionString()

	if err := createSchema(connStr, schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{SchemaName: schema})
	if err != nil {
		return fmt.Errorf("error creating pgx migrate driver: %w", err)
	}

	migr, err := migrate.NewWithDatabaseInstance("file://migrations/"+schema, config.Name, driver)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %w", err)
	}

	slog.Info("executing migrations", "schema", schema)

	if err := migr.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		slog.Warn("failed to run migrations", "error", err, "schema", schema)

		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("successfully ran migrations", "schema", schema)

	return nil
}

func createSchema(connStr, schema string) error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to DB for schema creation: %w", err)
	}

	defer func(conn *pgx.Conn, ctx context.Context) {
		if err := conn.Close(ctx); err != nil {
			slog.Error("failed to close connection", "error", err)
		}
	}(conn, ctx)

	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schema, err)
	}

	return nil
}
