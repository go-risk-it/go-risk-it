package region

import (
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region/assignment"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewService,
		assignment.NewAssignmentService,
	),
)
