package db

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/jackc/pgx/v5"
)

// InTransaction executes txFunc within a transaction with default isolation (ReadCommitted).
// This is a type-safe generic wrapper around Querier.ExecuteInTransaction.
func InTransaction[T any](
	q Querier,
	ctx ctx.LogContext,
	txFunc func(Querier) (T, error),
) (T, error) {
	return InTransactionWithIsolation(q, ctx, pgx.ReadCommitted, txFunc)
}

// InTransactionWithIsolation executes txFunc within a transaction
// with the specified isolation level.
// This is a type-safe generic wrapper around Querier.ExecuteInTransactionWithIsolation.
func InTransactionWithIsolation[T any](
	q Querier,
	ctx ctx.LogContext,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Querier) (T, error),
) (T, error) {
	result, err := q.ExecuteInTransactionWithIsolation(
		ctx,
		isolationLevel,
		func(q Querier) (any, error) {
			return txFunc(q)
		},
	)
	if err != nil {
		var zero T

		return zero, err
	}

	//nolint:forcetypeassert // guaranteed by txFunc signature
	return result.(T), nil
}
