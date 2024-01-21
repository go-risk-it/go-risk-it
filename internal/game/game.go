package game

import (
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/game"
	"github.com/tomfran/go-risk-it/internal/game/player"
	"github.com/tomfran/go-risk-it/internal/game/region"
	"github.com/tomfran/go-risk-it/internal/game/region/assignment"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			assignment.NewAssignmentService,
			fx.As(new(assignment.Service)),
		),
		fx.Annotate(
			board.NewBoardService,
			fx.As(new(board.Service)),
		),
		fx.Annotate(
			region.NewRegionService,
			fx.As(new(region.Service)),
		),
		fx.Annotate(
			player.NewPlayersService,
			fx.As(new(player.Service)),
		),
		fx.Annotate(
			game.NewGameService,
			fx.As(new(game.Service)),
		),
	),
)
