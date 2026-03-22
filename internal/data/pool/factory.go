package pool

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
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
	connectionPool, err := createPool(cfg, schema, initSQL...)
	if err != nil {
		return nil, err
	}

	log.Infow("created connection pool",
		"schema", schema,
		"maxConns", cfg.MaxConns,
		"minConns", cfg.MinConns,
	)

	poolStatsDone := make(chan struct{})

	lifecycle.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go reportPoolStats(connectionPool, poolStatsDone)

				return nil
			},
			OnStop: func(_ context.Context) error {
				close(poolStatsDone)
				connectionPool.Close()

				return nil
			},
		},
	)

	return connectionPool, nil
}

func createPool(
	cfg config.DatabaseConfig,
	schema string,
	initSQL ...string,
) (*pgxpool.Pool, error) {
	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(cfg.BuildConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s connection string: %w", schema, err)
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}

	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}

	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	connectionPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s connection pool: %w", schema, err)
	}

	for _, sql := range initSQL {
		if _, err := connectionPool.Exec(ctx, sql); err != nil {
			connectionPool.Close()

			return nil, fmt.Errorf("unable to execute init SQL for %s: %w", schema, err)
		}
	}

	return connectionPool, nil
}

// reportPoolStats periodically exports pgxpool stats as OTel metrics.
func reportPoolStats(pool *pgxpool.Pool, done <-chan struct{}) {
	meter := otel.Meter("db.pool")

	activeConns, _ := meter.Int64ObservableGauge("db.pool.active",
		metric.WithDescription("Number of active (in-use) connections"),
	)
	idleConns, _ := meter.Int64ObservableGauge("db.pool.idle",
		metric.WithDescription("Number of idle connections"),
	)
	totalConns, _ := meter.Int64ObservableGauge("db.pool.total",
		metric.WithDescription("Total connections in pool"),
	)
	acquireCount, _ := meter.Int64ObservableCounter("db.pool.acquires.total",
		metric.WithDescription("Cumulative connection acquires"),
	)
	acquireDuration, _ := meter.Float64ObservableGauge("db.pool.acquire.duration",
		metric.WithDescription("Average time to acquire a connection"),
		metric.WithUnit("s"),
	)
	emptyAcquires, _ := meter.Int64ObservableCounter(
		"db.pool.empty_acquires.total",
		metric.WithDescription(
			"Cumulative acquires that waited for a connection (pool was empty)",
		),
	)
	canceledAcquires, _ := meter.Int64ObservableCounter("db.pool.canceled_acquires.total",
		metric.WithDescription("Cumulative acquires canceled while waiting"),
	)
	maxConns, _ := meter.Int64ObservableGauge("db.pool.max_conns",
		metric.WithDescription("Maximum number of connections allowed in the pool"),
	)

	_, _ = meter.RegisterCallback(
		func(_ context.Context, observer metric.Observer) error {
			stat := pool.Stat()
			inUse := int64(stat.AcquiredConns())
			idle := int64(stat.IdleConns())

			observer.ObserveInt64(activeConns, inUse)
			observer.ObserveInt64(idleConns, idle)
			observer.ObserveInt64(totalConns, int64(stat.TotalConns()))
			observer.ObserveInt64(acquireCount, stat.AcquireCount())
			observer.ObserveInt64(emptyAcquires, stat.EmptyAcquireCount())
			observer.ObserveInt64(canceledAcquires, stat.CanceledAcquireCount())
			observer.ObserveInt64(maxConns, int64(stat.MaxConns()))

			if stat.AcquireCount() > 0 {
				avgAcquire := stat.AcquireDuration().Seconds() / float64(stat.AcquireCount())
				observer.ObserveFloat64(acquireDuration, avgAcquire)
			}

			return nil
		},
		activeConns, idleConns, totalConns, acquireCount, acquireDuration,
		emptyAcquires, canceledAcquires, maxConns,
	)

	// Keep goroutine alive until done signal.
	<-done
}
