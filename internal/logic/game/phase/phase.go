package phase

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase/walker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		walker.Module,
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
	),
)
