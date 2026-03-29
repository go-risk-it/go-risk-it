package advancement

import (
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
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
	bus eventbus.Bus,
	metrics *metrics.InfraMetrics,
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
	bus eventbus.Bus,
	metrics *metrics.InfraMetrics,
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
	bus eventbus.Bus,
	metrics *metrics.InfraMetrics,
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
