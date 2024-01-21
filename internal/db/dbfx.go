package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/fx"
)

type DB interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewConnectionPool,
			fx.As(new(DBTX)),
			fx.As(new(DB)),
		),
		New,
	),
)
