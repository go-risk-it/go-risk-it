package signals

import (
	"github.com/maniartech/signals"
)

type PlayerConnectedData struct{}

type PlayerConnectedSignal interface {
	signals.Signal[PlayerConnectedData]
}

func NewPlayerConnectedSignal() PlayerConnectedSignal {
	return signals.New[PlayerConnectedData]()
}
