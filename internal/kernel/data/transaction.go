package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
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
	txMetrics *metrics.StateMetrics,
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
	txMetrics *metrics.StateMetrics,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Q) (T, error),
) (T, error) {
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			observe.Warn(ctx, "retrying transaction",
				attribute.Int("attempt", attempt+1),
				attribute.Int("max_retries", maxRetries),
				attribute.String("last_error", lastErr.Error()),
			)

			if txMetrics != nil {
				txMetrics.TransactionRetries.Add(ctx, 1)
			}

			time.Sleep(retryBackoffBase * time.Duration(1<<attempt))
		}

		transaction, err := querier.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolationLevel})
		if err != nil {
			var zero T

			return zero, fmt.Errorf("failed to begin transaction: %w", err)
		}

		result, err := executeTransaction(
			querier, ctx, isolationLevel, txFunc, transaction,
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
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Q) (T, error),
	transaction pgx.Tx,
) (result T, err error) {
	spanCtx, done := observe.Span(ctx, "db.transaction",
		attribute.String("isolation", string(isolationLevel)),
	)

	defer func() {
		span := trace.SpanFromContext(spanCtx)

		if panicking := recover(); panicking != nil {
			observe.Error(
				ctx,
				fmt.Errorf("panic: %v", panicking),
				"panic in transaction, rolling back",
			)
			rollback(transaction, ctx)

			span.SetAttributes(attribute.String("db_outcome", "rollback"))
			done(fmt.Errorf("panic: %v", panicking))

			panic(panicking) // re-throw panic after Rollback
		} else if err != nil {
			observe.Error(ctx, err, "error in transaction, rolling back")
			rollback(transaction, ctx)

			span.SetAttributes(attribute.String("db_outcome", "rollback"))
			done(err)
		} else {
			err = transaction.Commit(ctx) // err is nil; if Commit returns error update err
			if err != nil {
				observe.Error(ctx, err, "failed to commit transaction")

				span.SetAttributes(attribute.String("db_outcome", "rollback"))
				done(err)
			} else {
				span.SetAttributes(attribute.String("db_outcome", "commit"))
				done(nil)
			}
		}
	}()

	result, err = txFunc(querier.WithTx(transaction))

	return result, err
}

func rollback(
	transaction pgx.Tx,
	ctx context.Context,
) {
	err := transaction.Rollback(ctx)
	if err != nil {
		observe.Error(ctx, err, "failed to rollback transaction")
	}
}
