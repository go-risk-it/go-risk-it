package signals

import "github.com/maniartech/signals"

type PlayerStateChangedData struct {
	GameID int64
}

type PlayerStateChangedSignal interface {
	signals.Signal[PlayerStateChangedData]
}

func NewPlayerStateChangedSignal() PlayerStateChangedSignal {
	return signals.New[PlayerStateChangedData]()
}
