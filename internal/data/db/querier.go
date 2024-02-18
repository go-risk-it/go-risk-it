package db

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/data/pool"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
)

type Querier interface {
	sqlc.Querier
	Transact(ctx context.Context, txFunc func(Querier) error) error
}

var _ Querier = (*Queries)(nil)

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
