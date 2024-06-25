package signals

import (
	"github.com/maniartech/signals"
)

type BoardStateChangedData struct {
	GameID int64
}

type BoardStateChangedSignal interface {
	signals.Signal[BoardStateChangedData]
}

func NewBoardStateChangedSignal() BoardStateChangedSignal {
	return signals.New[BoardStateChangedData]()
}
