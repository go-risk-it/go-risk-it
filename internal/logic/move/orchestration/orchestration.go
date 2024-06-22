package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration/phase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			phase.NewService,
			fx.As(new(phase.Service)),
		),
		fx.Annotate(
			orchestration.NewService,
			fx.As(new(orchestration.Service)),
		),
	),
)
