package performer

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/dice"
	"go.uber.org/fx"
)

var Module = fx.Options(
	dice.Module,
	fx.Provide(
		fx.Annotate(
			deploy.NewService,
			fx.As(new(deploy.Service)),
		),
		fx.Annotate(
			attack.NewService,
			fx.As(new(attack.Service)),
		),
	),
)
