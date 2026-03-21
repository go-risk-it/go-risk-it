package db

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/jackc/pgx/v5"
)

// Transactable is the minimal interface a querier must satisfy to participate
// in generic transactions.
type Transactable[Q any] interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	WithTx(tx pgx.Tx) Q
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

	return executeTransaction(querier, ctx, txFunc, transaction)
}

//nolint:nonamedreturns // named returns needed for defer-based commit/rollback
func executeTransaction[Q Transactable[Q], T any](
	querier Q,
	ctx ctx.LogContext,
	txFunc func(Q) (T, error),
	transaction pgx.Tx,
) (result T, err error) {
	ctx.Log().Infow("started transaction")

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
	}()

	result, err = txFunc(querier.WithTx(transaction))

	return result, err
}

func rollback(transaction pgx.Tx, ctx ctx.LogContext) {
	ctx.Log().Infow("rolling back transaction")

	err := transaction.Rollback(ctx)
	if err != nil {
		ctx.Log().Errorf("failed to rollback transaction: %v", err)
	}

	ctx.Log().Infow("rolled back transaction")
}
