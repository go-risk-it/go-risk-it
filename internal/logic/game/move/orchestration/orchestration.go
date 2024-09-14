package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
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

var (
	_ DeployOrchestrator  = (*OrchestratorImpl[deploy.Move, *deploy.MoveResult])(nil)
	_ AttackOrchestrator  = (*OrchestratorImpl[attack.Move, *attack.MoveResult])(nil)
	_ ConquerOrchestrator = (*OrchestratorImpl[conquer.Move, *conquer.MoveResult])(nil)
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
			validation.NewService,
			fx.As(new(validation.Service)),
		),
	),
)
