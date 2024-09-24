package signals

import "github.com/maniartech/signals"

type GameStateChangedData struct{}

type GameStateChangedSignal interface {
	signals.Signal[GameStateChangedData]
}

func NewGameStateChangedSignal() GameStateChangedSignal {
	return signals.New[GameStateChangedData]()
}
