package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	"go.uber.org/fx"
)

type DeployOrchestrator interface {
	Orchestrator[deploy.Move]
}

type AttackOrchestrator interface {
	Orchestrator[attack.Move]
}

type ConquerOrchestrator interface {
	Orchestrator[conquer.Move]
}

type ReinforceOrchestrator interface {
	Orchestrator[reinforce.Move]
}

type CardsOrchestrator interface {
	Orchestrator[cards.Move]
}

var (
	_ DeployOrchestrator    = (*OrchestratorImpl[deploy.Move])(nil)
	_ AttackOrchestrator    = (*OrchestratorImpl[attack.Move])(nil)
	_ ConquerOrchestrator   = (*OrchestratorImpl[conquer.Move])(nil)
	_ ReinforceOrchestrator = (*OrchestratorImpl[reinforce.Move])(nil)
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewOrchestrator[deploy.Move],
			fx.As(new(DeployOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[attack.Move],
			fx.As(new(AttackOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[conquer.Move],
			fx.As(new(ConquerOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[reinforce.Move],
			fx.As(new(ReinforceOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[cards.Move],
			fx.As(new(CardsOrchestrator)),
		),
		fx.Annotate(
			validation.New,
			fx.As(new(validation.Service)),
		),
		fx.Annotate(
			logging.New,
			fx.As(new(logging.Service)),
		),
	),
)
