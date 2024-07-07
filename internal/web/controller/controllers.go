package controller

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewBoardController,
			fx.As(new(BoardController)),
		),
		fx.Annotate(
			NewGameController,
			fx.As(new(GameController)),
		),
		fx.Annotate(
			NewMoveController,
			fx.As(new(MoveController)),
		),
		fx.Annotate(
			NewPlayerController,
			fx.As(new(PlayerController)),
		),
		fx.Annotate(
			NewPhaseController,
			fx.As(new(PhaseController)),
		),
	),
)
