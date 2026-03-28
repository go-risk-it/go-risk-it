package advancement

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/events"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"go.uber.org/fx"
)

type AttackAdvancer = Service[attack.Move, *attack.MoveResult]

type CardsAdvancer = Service[cards.Move, *cards.MoveResult]

type ReinforceAdvancer = Service[reinforce.Move, struct{}]

func NewAttackAdvancer(
	gameState state.Service,
	querier db.Querier,
	moveService moveservice.Service[attack.Move, *attack.MoveResult],
	validationService validation.Service,
	bus events.Bus,
	metrics *metrics.Metrics,
) AttackAdvancer {
	return NewService[attack.Move, *attack.MoveResult](
		gameState,
		querier,
		moveService,
		validationService,
		bus,
		metrics,
	)
}

func NewCardsAdvancer(
	gameState state.Service,
	querier db.Querier,
	moveService moveservice.Service[cards.Move, *cards.MoveResult],
	validationService validation.Service,
	bus events.Bus,
	metrics *metrics.Metrics,
) CardsAdvancer {
	return NewService[cards.Move, *cards.MoveResult](
		gameState,
		querier,
		moveService,
		validationService,
		bus,
		metrics,
	)
}

func NewReinforceAdvancer(
	gameState state.Service,
	querier db.Querier,
	moveService moveservice.Service[reinforce.Move, struct{}],
	validationService validation.Service,
	bus events.Bus,
	metrics *metrics.Metrics,
) ReinforceAdvancer {
	return NewService[reinforce.Move, struct{}](
		gameState,
		querier,
		moveService,
		validationService,
		bus,
		metrics,
	)
}

var Module = fx.Options(
	fx.Provide(
		NewAttackAdvancer,
		NewCardsAdvancer,
		NewReinforceAdvancer,
	),
)
