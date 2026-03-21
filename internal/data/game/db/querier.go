package db

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/pool"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/jackc/pgx/v5"
)

type Querier interface {
	sqlc.Querier
	db.Transactable[Querier]
}

var (
	_ Querier                  = (*Queries)(nil)
	_ db.Transactable[Querier] = (*Queries)(nil)
)

type Queries struct {
	*sqlc.Queries

	db pool.DB
}

func New(db pool.DB) Querier {
	return &Queries{
		Queries: sqlc.New(db),
		db:      db,
	}
}

func (q *Queries) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return q.db.BeginTx(ctx, txOptions)
}

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
		db:      q.db,
	}
}
