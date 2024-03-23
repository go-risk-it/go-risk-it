package logic

import (
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/move"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"github.com/tomfran/go-risk-it/internal/logic/region"
	"github.com/tomfran/go-risk-it/internal/logic/region/assignment"
	"github.com/tomfran/go-risk-it/internal/signals"
	"go.uber.org/fx"
)

var Module = fx.Options(
	signals.Module,
	fx.Provide(
		fx.Annotate(
			assignment.NewAssignmentService,
			fx.As(new(assignment.Service)),
		),
		fx.Annotate(
			board.NewService,
			fx.As(new(board.Service)),
		),
		fx.Annotate(
			region.NewService,
			fx.As(new(region.Service)),
		),
		fx.Annotate(
			player.NewService,
			fx.As(new(player.Service)),
		),
		fx.Annotate(
			game.NewService,
			fx.As(new(game.Service)),
		),
		fx.Annotate(
			move.NewService,
			fx.As(new(move.Service)),
		),
	),
)
