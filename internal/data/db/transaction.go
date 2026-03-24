package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Transactable is the minimal interface a querier must satisfy to participate
// in generic transactions.
type Transactable[Q any] interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	WithTx(tx pgx.Tx) Q
}

const (
	maxRetries           = 3
	serializationFailure = "40001"
	deadlockDetected     = "40P01"
	retryBackoffBase     = 10 * time.Millisecond
)

// InTransaction executes txFunc within a transaction with default isolation (ReadCommitted).
// Pass nil for txMetrics if metrics are not available.
func InTransaction[Q Transactable[Q], T any](
	querier Q,
	ctx context.Context,
	txMetrics *metrics.Metrics,
	txFunc func(Q) (T, error),
) (T, error) {
	return InTransactionWithIsolation(querier, ctx, txMetrics, pgx.ReadCommitted, txFunc)
}

// InTransactionWithIsolation executes txFunc within a transaction
// with the specified isolation level. Retries on serialization failures
// (PostgreSQL 40001) and deadlocks (40P01) up to maxRetries times.
// Pass nil for txMetrics if metrics are not available.
func InTransactionWithIsolation[Q Transactable[Q], T any](
	querier Q,
	ctx context.Context,
	txMetrics *metrics.Metrics,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Q) (T, error),
) (T, error) {
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			slog.WarnContext(ctx, "retrying transaction",
				"attempt", attempt+1,
				"maxRetries", maxRetries,
				"lastError", lastErr,
			)

			if txMetrics != nil {
				txMetrics.TransactionRetries.Add(ctx, 1)
			}

			time.Sleep(retryBackoffBase * time.Duration(1<<attempt))
		}

		slog.InfoContext(ctx, "starting transaction",
			"isolation", isolationLevel,
			"attempt", attempt+1,
		)

		transaction, err := querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolationLevel})
		if err != nil {
			var zero T

			return zero, fmt.Errorf("failed to begin transaction: %w", err)
		}

		result, err := executeTransaction(
			querier, ctx, txMetrics, isolationLevel, txFunc, transaction,
		)
		if err != nil && isRetryable(err) {
			lastErr = err

			continue
		}

		return result, err
	}

	var zero T

	return zero, fmt.Errorf("transaction failed after %d attempts: %w", maxRetries, lastErr)
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == serializationFailure || pgErr.Code == deadlockDetected
	}

	return false
}

//nolint:nonamedreturns // named returns needed for defer-based commit/rollback
func executeTransaction[Q Transactable[Q], T any](
	querier Q,
	ctx context.Context,
	txMetrics *metrics.Metrics,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Q) (T, error),
	transaction pgx.Tx,
) (result T, err error) {
	_, span := otel.GetTracerProvider().Tracer("go-risk-it-db").Start(
		ctx, "db.transaction",
		trace.WithAttributes(
			attribute.String("isolation", string(isolationLevel)),
		),
	)
	defer span.End()

	slog.InfoContext(ctx, "started transaction")

	txStart := time.Now()
	isoAttr := metric.WithAttributes(attribute.String("isolation", string(isolationLevel)))

	defer func() {
		if panicking := recover(); panicking != nil {
			slog.ErrorContext(ctx, "panic in transaction, rolling back", "panic", panicking)
			rollback(transaction, ctx, txMetrics)

			panic(panicking) // re-throw panic after Rollback
		} else if err != nil {
			slog.ErrorContext(ctx, "error in transaction, rolling back", "error", err)
			rollback(transaction, ctx, txMetrics)
		} else {
			err = transaction.Commit(ctx) // err is nil; if Commit returns error update err
			if err != nil {
				slog.ErrorContext(ctx, "failed to commit transaction", "error", err)
			} else {
				slog.InfoContext(ctx, "committed transaction")
			}
		}

		if txMetrics != nil {
			txMetrics.TransactionDuration.Record(
				ctx, time.Since(txStart).Seconds(), isoAttr,
			)
		}
	}()

	result, err = txFunc(querier.WithTx(transaction))

	return result, err
}

func rollback(
	transaction pgx.Tx,
	ctx context.Context,
	txMetrics *metrics.Metrics,
) {
	slog.InfoContext(ctx, "rolling back transaction")

	if txMetrics != nil {
		txMetrics.TransactionRollbacks.Add(ctx, 1)
	}

	err := transaction.Rollback(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to rollback transaction", "error", err)
	}

	slog.InfoContext(ctx, "rolled back transaction")
}
