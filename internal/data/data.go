package data

import (
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/pool"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"go.uber.org/fx"
)

var Module = fx.Options(
	pool.Module,
	fx.Provide(
		db.New,
		sqlc.New,
	),
)
