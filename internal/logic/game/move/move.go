package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer"
	"go.uber.org/fx"
)

var Module = fx.Options(
	performer.Module,
	orchestration.Module,
)
