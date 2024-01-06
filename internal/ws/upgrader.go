package ws

import (
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

func NewUpgrader(logger *zap.SugaredLogger) *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnOpen(func(c *websocket.Conn) {
		// echo
		logger.Info("OnOpen:", zap.String("remoteAddress", c.RemoteAddr().String()))
		c.WriteMessage(1, []byte("Stato iniziale"))
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		// echo
		logger.Infow("OnMessage:", "messageType", messageType, "data", string(data))
		c.WriteMessage(messageType, data)
		for i := 0; i < 100; i++ {
			c.WriteMessage(messageType, []byte("Fottiti coglione"))
		}
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		logger.Infow("OnClose:", "remoteAddress", c.RemoteAddr().String(), "error", err)
	})
	return u
}
