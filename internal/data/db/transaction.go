package db

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/pool"
	pgx "github.com/jackc/pgx/v5"
)

func (q *Queries) WithTx(tx pool.Transaction) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
		log:     q.log,
		db:      q.db,
	}
}

func (q *Queries) ExecuteInTransaction(
	ctx context.Context,
	txFunc func(Querier) (interface{}, error),
) (interface{}, error) {
	q.log.Infow("starting transaction")

	transaction, err := q.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return q.executeTransaction(ctx, txFunc, transaction)
}

func (q *Queries) ExecuteInTransactionWithIsolation(
	ctx context.Context,
	txFunc func(Querier) (interface{}, error),
	isolationLevel pgx.TxIsoLevel,
) (interface{}, error) {
	q.log.Infow("starting transaction", "isolation", isolationLevel)

	transaction, err := q.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolationLevel})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return q.executeTransaction(ctx, txFunc, transaction)
}

func (q *Queries) executeTransaction(
	ctx context.Context,
	txFunc func(Querier) (interface{}, error),
	transaction pgx.Tx,
) (interface{}, error) {
	q.log.Infow("started transaction")

	var err error

	defer func() {
		if panicking := recover(); panicking != nil {
			q.log.Errorw("panic in transaction, rolling back", "panic", panicking)
			q.rollback(transaction, ctx)

			panic(panicking) // re-throw panic after Rollback
		} else if err != nil {
			q.rollback(transaction, ctx)
		} else {
			err = transaction.Commit(ctx) // err is nil; if Commit returns error update err
			if err != nil {
				q.log.Errorw("failed to commit transaction", "err", err)
			} else {
				q.log.Infow("committed transaction")
			}
		}
	}()

	result, err := txFunc(q.WithTx(transaction))

	return result, err
}

func (q *Queries) rollback(transaction pgx.Tx, ctx context.Context) {
	q.log.Infow("rolling back transaction")

	err := transaction.Rollback(ctx)
	if err != nil {
		q.log.Errorf("failed to rollback transaction: %v", err)
	}

	q.log.Infow("rolled back transaction")
}
