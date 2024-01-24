package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GetQuerier(ctx context.Context) (Querier, error) {
	connStr, err := setupPostgresTestcontainer(ctx)
	if err != nil {
		return nil, err
	}

	// migrate
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get path")
	}

	pathToMigrationFiles := filepath.Dir(path) + "/migrations"
	m, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), connStr)
	if err != nil {
		return nil, err
	}

	defer m.Close()
	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	// create pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return New(pool), nil
}

func setupPostgresTestcontainer(ctx context.Context) (string, error) {
	dbName := "risk-it"
	dbUser := "postgres"
	dbPassword := "password"

	// build container
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:latest"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return "", err
	}
	return container.ConnectionString(ctx, "sslmode=disable", "application_name=test")
}
