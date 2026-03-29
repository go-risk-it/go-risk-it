package mission

import (
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission/checker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			checker.NewTwoContinentsChecker,
			fx.As(new(checker.MissionChecker)),
			fx.ResultTags(`group:"mission_checkers"`),
		),
		fx.Annotate(
			checker.NewTwoContinentsPlusOneChecker,
			fx.As(new(checker.MissionChecker)),
			fx.ResultTags(`group:"mission_checkers"`),
		),
		fx.Annotate(
			checker.NewEighteenTerritoriesChecker,
			fx.As(new(checker.MissionChecker)),
			fx.ResultTags(`group:"mission_checkers"`),
		),
		fx.Annotate(
			checker.NewTwentyFourTerritoriesChecker,
			fx.As(new(checker.MissionChecker)),
			fx.ResultTags(`group:"mission_checkers"`),
		),
		fx.Annotate(
			checker.NewEliminatePlayerChecker,
			fx.As(new(checker.MissionChecker)),
			fx.ResultTags(`group:"mission_checkers"`),
		),
		fx.Annotate(
			checker.NewRegistry,
			fx.ParamTags(`group:"mission_checkers"`),
		),
		New,
	),
)
