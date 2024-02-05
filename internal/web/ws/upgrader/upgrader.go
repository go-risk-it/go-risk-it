package upgrader

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/web/ws/handler"
	"go.uber.org/zap"
)

func New(log *zap.SugaredLogger, messageHandler handler.MessageHandler) *websocket.Upgrader {
	//exhaustruct:ignore
	upgrader := websocket.Upgrader{
		// Resolve cross-domain problems
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	upgrader.OnOpen(func(c *websocket.Conn) {
		// echo
		log.Info("OnOpen:", zap.String("remoteAddress", c.RemoteAddr().String()))
		err := c.WriteMessage(1, []byte("Established connection"))
		if err != nil {
			panic(err)
		}
	})

	upgrader.OnMessage(messageHandler.OnMessage)

	upgrader.OnClose(func(c *websocket.Conn, err error) {
		log.Infow("OnClose:", "remoteAddress", c.RemoteAddr().String(), "error", err)
	})

	return &upgrader
}
