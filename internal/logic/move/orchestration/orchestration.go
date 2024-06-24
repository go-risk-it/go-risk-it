package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration/validation"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			phase.NewService,
			fx.As(new(phase.Service)),
		),
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
		fx.Annotate(
			validation.NewService,
			fx.As(new(validation.Service)),
		),
	),
)
