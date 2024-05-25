package db

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/data/pool"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Querier interface {
	sqlc.Querier
	ExecuteInTransaction(
		ctx context.Context,
		txFunc func(Querier) (interface{}, error)) (interface{}, error)
	ExecuteInTransactionWithIsolation(
		ctx context.Context,
		isolationLevel pgx.TxIsoLevel,
		txFunc func(Querier) (interface{}, error),
	) (interface{}, error)
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
