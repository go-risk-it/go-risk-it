package db

import (
	"context"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/tomfran/go-risk-it/internal/data/pool"
)

func (q *Queries) WithTx(tx pool.Transaction) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
		log:     q.log,
		db:      q.db,
	}
}

func (q *Queries) ExecuteInTransaction(ctx context.Context, txFunc func(Querier) error) error {
	q.log.Infow("starting transaction")

	transaction, err := q.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	q.log.Infow("started transaction")

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

	err = txFunc(q.WithTx(transaction))

	return err
}

func (q *Queries) rollback(transaction pgx.Tx, ctx context.Context) {
	err := transaction.Rollback(ctx)
	if err != nil {
		q.log.Errorf("failed to rollback transaction: %v", err)
	}
}
