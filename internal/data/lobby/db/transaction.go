package db

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/pool"
	"github.com/jackc/pgx/v5"
)

func (q *Queries) WithTx(tx pool.Transaction) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
		db:      q.db,
	}
}

// ExecuteInTransaction with default isolation level (ReadCommitted).
func (q *Queries) ExecuteInTransaction(
	ctx ctx.LogContext,
	txFunc func(Querier) (interface{}, error),
) (interface{}, error) {
	return q.ExecuteInTransactionWithIsolation(ctx, pgx.ReadCommitted, txFunc)
}

func (q *Queries) ExecuteInTransactionWithIsolation(
	ctx ctx.LogContext,
	isolationLevel pgx.TxIsoLevel,
	txFunc func(Querier) (interface{}, error),
) (interface{}, error) {
	ctx.Log().Infow("starting transaction", "isolation", isolationLevel)

	transaction, err := q.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolationLevel})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return q.executeTransaction(ctx, txFunc, transaction)
}

func (q *Queries) executeTransaction(
	ctx ctx.LogContext,
	txFunc func(Querier) (interface{}, error),
	transaction pgx.Tx,
) (interface{}, error) {
	ctx.Log().Infow("started transaction")

	var err error

	defer func() {
		if panicking := recover(); panicking != nil {
			ctx.Log().Errorw("panic in transaction, rolling back", "panic", panicking)
			q.rollback(transaction, ctx)

			panic(panicking) // re-throw panic after Rollback
		} else if err != nil {
			ctx.Log().Errorw("error in transaction, rolling back", "err", err)
			q.rollback(transaction, ctx)
		} else {
			err = transaction.Commit(ctx) // err is nil; if Commit returns error update err
			if err != nil {
				ctx.Log().Errorw("failed to commit transaction", "err", err)
			} else {
				ctx.Log().Infow("committed transaction")
			}
		}
	}()

	result, err := txFunc(q.WithTx(transaction))

	return result, err
}

func (q *Queries) rollback(transaction pgx.Tx, ctx ctx.LogContext) {
	ctx.Log().Infow("rolling back transaction")

	err := transaction.Rollback(ctx)
	if err != nil {
		ctx.Log().Errorf("failed to rollback transaction: %v", err)
	}

	ctx.Log().Infow("rolled back transaction")
}
