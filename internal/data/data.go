package data

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	lobby.Module,
)
