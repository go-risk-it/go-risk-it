package logic

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	signals.Module,
)
