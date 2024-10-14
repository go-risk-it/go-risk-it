package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	"go.uber.org/fx"
)

type DeployOrchestrator interface {
	Orchestrator[deploy.Move, *deploy.MoveResult]
}

type AttackOrchestrator interface {
	Orchestrator[attack.Move, *attack.MoveResult]
}

type ConquerOrchestrator interface {
	Orchestrator[conquer.Move, *conquer.MoveResult]
}

type ReinforceOrchestrator interface {
	Orchestrator[reinforce.Move, *reinforce.MoveResult]
}

type CardsOrchestrator interface {
	Orchestrator[cards.Move, *cards.MoveResult]
}

var (
	_ DeployOrchestrator    = (*OrchestratorImpl[deploy.Move, *deploy.MoveResult])(nil)
	_ AttackOrchestrator    = (*OrchestratorImpl[attack.Move, *attack.MoveResult])(nil)
	_ ConquerOrchestrator   = (*OrchestratorImpl[conquer.Move, *conquer.MoveResult])(nil)
	_ ReinforceOrchestrator = (*OrchestratorImpl[reinforce.Move, *reinforce.MoveResult])(nil)
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewOrchestrator[deploy.Move, *deploy.MoveResult],
			fx.As(new(DeployOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[attack.Move, *attack.MoveResult],
			fx.As(new(AttackOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[conquer.Move, *conquer.MoveResult],
			fx.As(new(ConquerOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[reinforce.Move, *reinforce.MoveResult],
			fx.As(new(ReinforceOrchestrator)),
		),
		fx.Annotate(
			NewOrchestrator[cards.Move, *cards.MoveResult],
			fx.As(new(CardsOrchestrator)),
		),

		fx.Annotate(
			validation.NewService,
			fx.As(new(validation.Service)),
		),
	),
)
