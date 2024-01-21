package ws

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

type Message struct {
	PlayerId int
	GameId   int
	Payload  Payload
}

type Payload struct {
	StartRegionId int
	EndRegionId   int
	NumTroops     int
}

func NewUpgrader(logger *zap.SugaredLogger) *websocket.Upgrader {
	u := websocket.Upgrader{
		// Resolve cross-domain problems
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	u.OnOpen(func(c *websocket.Conn) {
		// echo
		logger.Info("OnOpen:", zap.String("remoteAddress", c.RemoteAddr().String()))
		err := c.WriteMessage(1, []byte("Established connection"))
		if err != nil {
			panic(err)
		}
	})

	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		logger.Infow("OnMessage:", "messageType", messageType, "data", string(data))
		err := c.WriteMessage(messageType, []byte("{\"hello\":\"there\"}"))
		if err != nil {
			panic(err)
		}
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		logger.Infow("OnClose:", "remoteAddress", c.RemoteAddr().String(), "error", err)
	})

	return &u
}
