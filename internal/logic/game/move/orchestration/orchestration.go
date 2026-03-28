package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/events"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/timing"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"go.uber.org/fx"
)

// OrchestratorDeps contains the non-generic dependencies shared by all orchestrator constructors.
type OrchestratorDeps struct {
	fx.In

	Querier           db.Querier
	GameService       state.Service
	LoggingService    logging.Service
	MissionService    mission.Service
	ValidationService validation.Service
	Bus               events.Bus
	Metrics           *metrics.Metrics
	GameTiming        *timing.GameTiming
}

type DeployOrchestrator = Orchestrator[deploy.Move, struct{}]

type AttackOrchestrator = Orchestrator[attack.Move, *attack.MoveResult]

type ConquerOrchestrator = Orchestrator[conquer.Move, struct{}]

type ReinforceOrchestrator = Orchestrator[reinforce.Move, struct{}]

type CardsOrchestrator = Orchestrator[cards.Move, *cards.MoveResult]

func NewDeployOrchestrator(
	deps OrchestratorDeps,
	service moveservice.Service[deploy.Move, struct{}],
) DeployOrchestrator {
	return newOrchestratorFromDeps[deploy.Move, struct{}](deps, service)
}

func NewAttackOrchestrator(
	deps OrchestratorDeps,
	service moveservice.Service[attack.Move, *attack.MoveResult],
) AttackOrchestrator {
	return newOrchestratorFromDeps[attack.Move, *attack.MoveResult](deps, service)
}

func NewConquerOrchestrator(
	deps OrchestratorDeps,
	service moveservice.Service[conquer.Move, struct{}],
) ConquerOrchestrator {
	return newOrchestratorFromDeps[conquer.Move, struct{}](deps, service)
}

func NewReinforceOrchestrator(
	deps OrchestratorDeps,
	service moveservice.Service[reinforce.Move, struct{}],
) ReinforceOrchestrator {
	return newOrchestratorFromDeps[reinforce.Move, struct{}](deps, service)
}

func NewCardsOrchestrator(
	deps OrchestratorDeps,
	service moveservice.Service[cards.Move, *cards.MoveResult],
) CardsOrchestrator {
	return newOrchestratorFromDeps[cards.Move, *cards.MoveResult](deps, service)
}

func newOrchestratorFromDeps[T, R any](
	deps OrchestratorDeps,
	service moveservice.Service[T, R],
) Orchestrator[T, R] {
	return NewOrchestrator[T, R](
		deps.Querier,
		service,
		deps.GameService,
		deps.LoggingService,
		deps.MissionService,
		deps.ValidationService,
		deps.Bus,
		deps.Metrics,
		deps.GameTiming,
	)
}

var Module = fx.Options(
	fx.Provide(
		NewDeployOrchestrator,
		NewAttackOrchestrator,
		NewConquerOrchestrator,
		NewReinforceOrchestrator,
		NewCardsOrchestrator,
		validation.New,
		logging.New,
	),
)
