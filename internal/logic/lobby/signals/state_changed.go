package signals

import "github.com/maniartech/signals"

type LobbyStateChangedData struct{}

type LobbyStateChangedSignal interface {
	signals.Signal[LobbyStateChangedData]
}

func NewLobbyStateChangedSignal() LobbyStateChangedSignal {
	return signals.New[LobbyStateChangedData]()
}
