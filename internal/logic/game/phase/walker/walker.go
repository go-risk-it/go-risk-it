package walker

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"go.uber.org/fx"
)

type PhaseWalker interface {
	Walk(ctx ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error)
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
		fx.Annotate(
			NewDeployPhaseWalker,
			fx.As(new(DeployPhaseWalker)),
		),
		fx.Annotate(
			NewAttackPhaseWalker,
			fx.As(new(AttackPhaseWalker)),
		),
	),
)
