package signals

import (
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/maniartech/signals"
)

type PlayerConnectedData struct {
	Connection *websocket.Conn
	GameID     int64
}

type PlayerConnectedSignal interface {
	signals.Signal[PlayerConnectedData]
}

func NewPlayerConnectedSignal() PlayerConnectedSignal {
	return signals.New[PlayerConnectedData]()
}
