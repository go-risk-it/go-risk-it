package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"go.uber.org/fx"
)

// OrchestratorDeps contains the non-generic dependencies shared by all orchestrator constructors.
type OrchestratorDeps struct {
	fx.In

	Querier           db.Querier
	GameService       state.Service
	LoggingService    LoggingService
	MissionService    mission.Service
	ValidationService ValidationService
	Bus               bus.Publisher
	StateMetrics      *metrics.StateMetrics
	GameMetrics       *gamemetrics.GameMetrics
	GameTiming        *gamemetrics.GameTiming
}

type orchestrator[T, R any] struct {
	querier           db.Querier
	service           moveservice.Service[T, R]
	gameService       state.Service
	loggingService    LoggingService
	missionService    mission.Service
	validationService ValidationService
	bus               bus.Publisher
	stateMetrics      *metrics.StateMetrics
	gameMetrics       *gamemetrics.GameMetrics
	gameTiming        *gamemetrics.GameTiming
}

var _ Orchestrator[any, any] = (*orchestrator[any, any])(nil)

func NewOrchestrator[T, R any](
	querier db.Querier,
	service moveservice.Service[T, R],
	gameService state.Service,
	loggingService LoggingService,
	missionService mission.Service,
	validationService ValidationService,
	bus bus.Publisher,
	stateMetrics *metrics.StateMetrics,
	gameMetrics *gamemetrics.GameMetrics,
	gameTiming *gamemetrics.GameTiming,
) Orchestrator[T, R] {
	return &orchestrator[T, R]{
		querier:           querier,
		service:           service,
		gameService:       gameService,
		loggingService:    loggingService,
		missionService:    missionService,
		validationService: validationService,
		bus:               bus,
		stateMetrics:      stateMetrics,
		gameMetrics:       gameMetrics,
		gameTiming:        gameTiming,
	}
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
		deps.StateMetrics,
		deps.GameMetrics,
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
		NewValidationService,
		NewLoggingService,
	),
)
