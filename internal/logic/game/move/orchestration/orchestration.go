package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/timing"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"go.uber.org/fx"
)

type DeployOrchestrator = Orchestrator[deploy.Move]

type AttackOrchestrator = Orchestrator[attack.Move]

type ConquerOrchestrator = Orchestrator[conquer.Move]

type ReinforceOrchestrator = Orchestrator[reinforce.Move]

type CardsOrchestrator = Orchestrator[cards.Move]

func NewDeployOrchestrator(
	querier db.Querier,
	service moveservice.Service[deploy.Move],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) DeployOrchestrator {
	return NewOrchestrator[deploy.Move](
		querier,
		service,
		gameService,
		loggingService,
		missionService,
		validationService,
		gameStateChangedSignal,
		metrics,
		gameTiming,
	)
}

func NewAttackOrchestrator(
	querier db.Querier,
	service moveservice.Service[attack.Move],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) AttackOrchestrator {
	return NewOrchestrator[attack.Move](
		querier,
		service,
		gameService,
		loggingService,
		missionService,
		validationService,
		gameStateChangedSignal,
		metrics,
		gameTiming,
	)
}

func NewConquerOrchestrator(
	querier db.Querier,
	service moveservice.Service[conquer.Move],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) ConquerOrchestrator {
	return NewOrchestrator[conquer.Move](
		querier,
		service,
		gameService,
		loggingService,
		missionService,
		validationService,
		gameStateChangedSignal,
		metrics,
		gameTiming,
	)
}

func NewReinforceOrchestrator(
	querier db.Querier,
	service moveservice.Service[reinforce.Move],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) ReinforceOrchestrator {
	return NewOrchestrator[reinforce.Move](
		querier,
		service,
		gameService,
		loggingService,
		missionService,
		validationService,
		gameStateChangedSignal,
		metrics,
		gameTiming,
	)
}

func NewCardsOrchestrator(
	querier db.Querier,
	service moveservice.Service[cards.Move],
	gameService state.Service,
	loggingService logging.Service,
	missionService mission.Service,
	validationService validation.Service,
	gameStateChangedSignal signals.GameStateChangedSignal,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) CardsOrchestrator {
	return NewOrchestrator[cards.Move](
		querier,
		service,
		gameService,
		loggingService,
		missionService,
		validationService,
		gameStateChangedSignal,
		metrics,
		gameTiming,
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
