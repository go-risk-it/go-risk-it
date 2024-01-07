package ws

import (
	"encoding/json"
	"fmt"
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
	u := websocket.NewUpgrader()
	u.OnOpen(func(c *websocket.Conn) {
		// echo
		logger.Info("OnOpen:", zap.String("remoteAddress", c.RemoteAddr().String()))
		c.WriteMessage(1, []byte("Established connection"))
	})

	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		logger.Infow("OnMessage:", "messageType", messageType, "data", string(data))

		var request Message

		err := json.Unmarshal(data, &request)
		if err != nil {
			logger.Error("Unable to unmarshal JSON due to %s", err)
		}

		fmt.Println("Received message: ", request)

		payload, err := json.Marshal(Message{
			PlayerId: 10,
			GameId:   20,
			Payload: Payload{
				StartRegionId: 100,
				EndRegionId:   200,
				NumTroops:     99,
			},
		})
		if err != nil {
			logger.Error("Unable to marshal JSON due to %s", err)
		}
		c.WriteMessage(messageType, payload)
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		logger.Infow("OnClose:", "remoteAddress", c.RemoteAddr().String(), "error", err)
	})

	return u
}
