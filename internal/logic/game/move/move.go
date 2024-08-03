package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			attack.NewService,
			fx.As(new(attack.Service)),
		),
		fx.Annotate(
			cards.NewService,
			fx.As(new(cards.Service)),
		),
		fx.Annotate(
			deploy.NewService,
			fx.As(new(deploy.Service)),
		),
		fx.Annotate(
			conquer.NewService,
			fx.As(new(conquer.Service)),
		),
	),
	dice.Module,
	orchestration.Module,
)
