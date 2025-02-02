package signals

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewLobbyStateChangedSignal,
		NewPlayerConnectedSignal,
	),
)
