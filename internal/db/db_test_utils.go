package db

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

func createContainerConnectionPool() (*pgxpool.Pool, error) {
	ctx := context.Background()

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
		return nil, err
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	if err != nil {
		return nil, err
	}

	// migrate
	m, err := migrate.New("file://migrations", connStr)
	if err != nil {
		return nil, err
	}

	defer m.Close()
	err = m.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	// create pool
	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
