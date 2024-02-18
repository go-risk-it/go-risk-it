package db

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/data/pool"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	"go.uber.org/zap"
)

type Querier interface {
	sqlc.Querier
	ExecuteInTransaction(ctx context.Context, txFunc func(Querier) error) error
}

var _ Querier = (*Queries)(nil)

type Queries struct {
	*sqlc.Queries
	log *zap.SugaredLogger
	db  pool.DB
}

func New(db pool.DB, log *zap.SugaredLogger) Querier {
	return &Queries{
		Queries: sqlc.New(db),
		log:     log,
		db:      db,
	}
}
