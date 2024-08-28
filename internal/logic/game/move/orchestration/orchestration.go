package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	"go.uber.org/fx"
)

type DeployOrchestrator interface {
	Orchestator[deploy.Move, *deploy.MoveResult]
}

type AttackOrchestrator interface {
	Orchestator[attack.Move, *attack.MoveResult]
}

type ConquerOrchestrator interface {
	Orchestator[conquer.Move, *conquer.MoveResult]
}

var (
	_ DeployOrchestrator  = (*OrchestatorImpl[deploy.Move, *deploy.MoveResult])(nil)
	_ AttackOrchestrator  = (*OrchestatorImpl[attack.Move, *attack.MoveResult])(nil)
	_ ConquerOrchestrator = (*OrchestatorImpl[conquer.Move, *conquer.MoveResult])(nil)
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
