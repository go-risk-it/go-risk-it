package logic

import (
	game "github.com/go-risk-it/go-risk-it/internal/game/logic"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	lobby.Module,
)
