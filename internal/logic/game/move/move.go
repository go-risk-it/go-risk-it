package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		deploy.NewService,
		attack.NewService,
		conquer.NewService,
		reinforce.NewService,
		cards.NewService,
	),
	dice.Module,
	orchestration.Module,
)
