package controllers

import (
	"github.com/tomfran/go-risk-it/internal/web/controllers/board"
	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"github.com/tomfran/go-risk-it/internal/web/controllers/player"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			game.New,
			fx.As(new(game.Controller)),
		),
		fx.Annotate(
			board.New,
			fx.As(new(board.Controller)),
		),
		fx.Annotate(
			player.New,
			fx.As(new(player.Controller)),
		),
	),
)
