package controllers

import (
	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		game.New,
	),
)
