package advancement

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	"go.uber.org/fx"
)

type AttackAdvancer interface {
	Service[attack.Move]
}

type CardsAdvancer interface {
	Service[cards.Move]
}

type ReinforceAdvancer interface {
	Service[reinforce.Move]
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService[attack.Move],
			fx.As(new(AttackAdvancer)),
		),
		fx.Annotate(
			NewService[cards.Move],
			fx.As(new(CardsAdvancer)),
		),
		fx.Annotate(
			NewService[reinforce.Move],
			fx.As(new(ReinforceAdvancer)),
		),
	),
)
