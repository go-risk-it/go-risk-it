package data

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var errGetPath = errors.New("failed to get path")

func GetQuerier(ctx ctx.LogContext) (db.Querier, error) {
	connStr, err := setupPostgresTestcontainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup postgres testcontainer: %w", err)
	}

	// migrate
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errGetPath
	}

	pathToMigrationFiles := filepath.Dir(path) + "/sqlc/migrations"

	mig, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate data: %w", err)
	}

	defer mig.Close()
	err = mig.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	// create pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	return db.New(pool, zap.NewNop().Sugar()), nil
}

func setupPostgresTestcontainer(ctx ctx.LogContext) (string, error) {
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
		return "", fmt.Errorf("failed to run testcontainer: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	if err != nil {
		return "", fmt.Errorf("failed to get connection string: %w", err)
	}

	return connStr, nil
}
