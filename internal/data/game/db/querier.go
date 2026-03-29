package db

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/data"
	"github.com/jackc/pgx/v5"
)

type Transaction interface {
	pgx.Tx
}

type DB interface {
	sqlc.DBTX
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Querier interface {
	sqlc.Querier
	data.Transactable[Querier]
}

var (
	_ Querier                    = (*Queries)(nil)
	_ data.Transactable[Querier] = (*Queries)(nil)
)

type Queries struct {
	*sqlc.Queries
	data.TxSupport
}

func New(d DB) Querier {
	return &Queries{
		Queries:   sqlc.New(d),
		TxSupport: data.TxSupport{DB: d},
	}
}

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries:   q.Queries.WithTx(tx),
		TxSupport: q.TxSupport,
	}
}
