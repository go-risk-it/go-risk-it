package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/dice"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer"
	"go.uber.org/fx"
)

var Module = fx.Options(
	dice.Module,
	performer.Module,
	orchestration.Module,
)
