package advancement

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"go.uber.org/fx"
)

type AttackAdvancer interface {
	Service[attack.Move, *attack.MoveResult]
}

type CardsAdvancer interface {
	Service[cards.Move, *cards.MoveResult]
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService[attack.Move, *attack.MoveResult],
			fx.As(new(AttackAdvancer)),
		),
		fx.Annotate(
			NewService[cards.Move, *cards.MoveResult],
			fx.As(new(CardsAdvancer)),
		),
	),
)
