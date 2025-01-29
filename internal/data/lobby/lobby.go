package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/pool"
	"go.uber.org/fx"
)

var Module = fx.Options(
	pool.Module,
	fx.Provide(
		db.New,
	),
)
