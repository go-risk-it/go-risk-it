package data

import (
	game "github.com/go-risk-it/go-risk-it/internal/game/data"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
)
