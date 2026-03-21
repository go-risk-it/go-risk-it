package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			deploy.NewService,
			fx.As(new(deploy.Service)),
			fx.As(new(service.Service[deploy.Move])),
		),
		fx.Annotate(
			attack.NewService,
			fx.As(new(attack.Service)),
			fx.As(new(service.Service[attack.Move])),
		),
		fx.Annotate(
			conquer.NewService,
			fx.As(new(conquer.Service)),
			fx.As(new(service.Service[conquer.Move])),
		),
		fx.Annotate(
			reinforce.NewService,
			fx.As(new(reinforce.Service)),
			fx.As(new(service.Service[reinforce.Move])),
		),
		fx.Annotate(
			cards.NewService,
			fx.As(new(cards.Service)),
			fx.As(new(service.Service[cards.Move])),
		),
	),
	dice.Module,
	orchestration.Module,
)
