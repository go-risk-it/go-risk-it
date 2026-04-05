package move

import (
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/reinforce"
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
