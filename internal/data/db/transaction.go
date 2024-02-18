package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"google.golang.org/appengine/log"
)

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
	}
}

func (q *Queries) Transact(ctx context.Context, txFunc func(Querier) error) error {
	transaction, err := q.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if panicking := recover(); panicking != nil {
			err := transaction.Rollback(ctx)
			if err != nil {
				log.Errorf(ctx, "failed to rollback transaction: %v", err)
			}

			panic(panicking) // re-throw panic after Rollback
		} else if err != nil {
			err2 := transaction.Rollback(ctx) // err is non-nil; don't change it
			if err2 != nil {
				log.Errorf(ctx, "failed to rollback transaction: %v", err2)
			}
		} else {
			err = transaction.Commit(ctx) // err is nil; if Commit returns error update err
		}
	}()

	err = txFunc(q.WithTx(transaction))

	return err
}
