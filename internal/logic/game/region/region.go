package region

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region/assignment"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
		fx.Annotate(
			assignment.NewAssignmentService,
			fx.As(new(assignment.Service)),
		),
	),
)
