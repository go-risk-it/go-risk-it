package data

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TxBeginner is the minimal interface for beginning transactions.
type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// TxSupport provides BeginTx for embedding in domain querier structs.
type TxSupport struct {
	DB TxBeginner
}

func (t TxSupport) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return t.DB.BeginTx(ctx, txOptions)
}
