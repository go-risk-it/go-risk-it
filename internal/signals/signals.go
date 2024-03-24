package signals

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewBoardStateChangedSignal,
		NewGameStateChangedSignal,
		NewPlayerStateChangedSignal,
		NewPlayerConnectedSignal,
	),
)
