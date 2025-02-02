package logic

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	lobby.Module,
)
