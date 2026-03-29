package data

import (
	"github.com/go-risk-it/go-risk-it/internal/data/lobby"
	game "github.com/go-risk-it/go-risk-it/internal/game/data"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	lobby.Module,
)
