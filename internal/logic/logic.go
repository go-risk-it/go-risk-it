package logic

import (
	game "github.com/go-risk-it/go-risk-it/internal/game/logic"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
)
