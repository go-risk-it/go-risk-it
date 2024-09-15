package connection

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

type Handler interface {
	OnMessage(
		connection *websocket.Conn,
		messageType websocket.MessageType,
		data []byte,
	)
}

type HandlerImpl struct {
	log               *zap.SugaredLogger
	connectionManager Manager
}

func NewHandler(
	log *zap.SugaredLogger,
	connectionManager Manager,
) *HandlerImpl {
	return &HandlerImpl{
		log:               log,
		connectionManager: connectionManager,
	}
}

func (m *HandlerImpl) OnMessage(
	_ *websocket.Conn,
	messageType websocket.MessageType,
	data []byte,
) {
	var requestMessage message.Message

	m.log.Infow("Received message", "messageType", messageType, "data", data)

	err := json.Unmarshal(data, &requestMessage)
	if err != nil {
		m.log.Info("Unable to unmarshal json: %v", err)

		return
	}
}
