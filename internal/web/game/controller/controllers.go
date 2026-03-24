package controller

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewBoardController,
		NewGameController,
		NewMoveController,
		NewPlayerController,
		NewPhaseController,
		NewAdvancementController,
		NewCardController,
		NewMoveLogController,
		NewMissionController,
	),
)
