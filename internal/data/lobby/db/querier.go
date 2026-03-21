package db

import (
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/pool"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
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
	db.TxSupport
}

func New(d pool.DB) Querier {
	return &Queries{
		Queries:   sqlc.New(d),
		TxSupport: db.TxSupport{DB: d},
	}
}

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries:   q.Queries.WithTx(tx),
		TxSupport: q.TxSupport,
	}
}
