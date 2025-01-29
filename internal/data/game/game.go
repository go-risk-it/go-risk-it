package game

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/pool"
	"go.uber.org/fx"
)

var Module = fx.Options(
	pool.Module,
	fx.Provide(
		db.New,
	),
)
