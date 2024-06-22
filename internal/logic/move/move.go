package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/validation"
	"go.uber.org/fx"
)

var Module = fx.Options(
	orchestration.Module,
	fx.Provide(
		fx.Annotate(
			deploy.NewService,
			fx.As(new(deploy.Service)),
		),
		fx.Annotate(
			attack.NewService,
			fx.As(new(attack.Service)),
		),
		fx.Annotate(
			validation.NewService,
			fx.As(new(validation.Service)),
		),
	),
)
