package db

import (
	"context"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Transactable is the minimal interface a querier must satisfy to participate
// in generic transactions.
type Transactable[Q any] interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	WithTx(tx pgx.Tx) Q
}

// Transaction metrics — initialized lazily via sync.Once inside the OTel meter.
var ( //nolint:gochecknoglobals
	txDuration  metric.Float64Histogram
	txRollbacks metric.Int64Counter
)

func init() { //nolint:gochecknoinits
	meter := otel.Meter("db.transaction")

	txDuration, _ = meter.Float64Histogram(
		"db.transaction.duration",
		metric.WithDescription("Duration of database transactions"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(
			0.001,
			0.005,
			0.01,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
			2.5,
			5,
			10,
		),
	)

	txRollbacks, _ = meter.Int64Counter("db.transaction.rollbacks.total",
		metric.WithDescription("Total number of transaction rollbacks"),
	)
}

// InTransaction executes txFunc within a transaction with default isolation (ReadCommitted).
func InTransaction[Q Transactable[Q], T any](
	querier Q,
	ctx ctx.LogContext,
	txFunc func(Q) (T, error),
) (T, error) {
	return InTransactionWithIsolation(querier, ctx, pgx.ReadCommitted, txFunc)
}

// InTransactionWithIsolation executes txFunc within a transaction
// with the specified isolation level. Fully type-safe — no any boxing.
func InTransactionWithIsolation[Q Transactable[Q], T any](
	querier Q,
	ctx ctx.LogContext,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Q) (T, error),
) (T, error) {
	ctx.Log().Infow("starting transaction", "isolation", isolationLevel)

	transaction, err := querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolationLevel})
	if err != nil {
		var zero T

		return zero, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return executeTransaction(querier, ctx, isolationLevel, txFunc, transaction)
}

//nolint:nonamedreturns // named returns needed for defer-based commit/rollback
func executeTransaction[Q Transactable[Q], T any](
	querier Q,
	ctx ctx.LogContext,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Q) (T, error),
	transaction pgx.Tx,
) (result T, err error) {
	ctx.Log().Infow("started transaction")

	txStart := time.Now()
	isoAttr := metric.WithAttributes(attribute.String("isolation", string(isolationLevel)))

	defer func() {
		if panicking := recover(); panicking != nil {
			ctx.Log().Errorw("panic in transaction, rolling back", "panic", panicking)
			rollback(transaction, ctx)

			panic(panicking) // re-throw panic after Rollback
		} else if err != nil {
			ctx.Log().Errorw("error in transaction, rolling back", "err", err)
			rollback(transaction, ctx)
		} else {
			err = transaction.Commit(ctx) // err is nil; if Commit returns error update err
			if err != nil {
				ctx.Log().Errorw("failed to commit transaction", "err", err)
			} else {
				ctx.Log().Infow("committed transaction")
			}
		}

		txDuration.Record(ctx, time.Since(txStart).Seconds(), isoAttr)
	}()

	result, err = txFunc(querier.WithTx(transaction))

	return result, err
}

func rollback(transaction pgx.Tx, ctx ctx.LogContext) {
	ctx.Log().Infow("rolling back transaction")

	txRollbacks.Add(ctx, 1)

	err := transaction.Rollback(ctx)
	if err != nil {
		ctx.Log().Errorf("failed to rollback transaction: %v", err)
	}

	ctx.Log().Infow("rolled back transaction")
}
