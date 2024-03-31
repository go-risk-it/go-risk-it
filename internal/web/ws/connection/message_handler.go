package connection

import (
	"encoding/json"
	"fmt"

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
	connection *websocket.Conn,
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

	err = m.handleMessage(requestMessage, connection)
	if err != nil {
		m.log.Errorf("unable to handle message: %v", err)

		return
	}
}

func (m *HandlerImpl) handleMessage(
	requestMessage message.Message,
	connection *websocket.Conn,
) error {
	m.log.Infow("Received message", "requestMessage", requestMessage)

	switch requestMessage.Type {
	case message.Subscribe:
		var joinGamePayload message.SubscribePayload

		err := json.Unmarshal(requestMessage.Payload, &joinGamePayload)
		if err != nil {
			return fmt.Errorf("unable to unmarshal json: %w", err)
		}

		m.connectionManager.ConnectPlayer(connection, joinGamePayload.GameID)

		return nil
	default:
		return fmt.Errorf("unknown message type: %s", requestMessage.Type)
	}
}
