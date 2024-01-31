package ws

import (
	"encoding/json"
	"fmt"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/api"
	game2 "github.com/tomfran/go-risk-it/internal/api/game"
	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"go.uber.org/zap"
)

type MessageHandler interface {
	OnMessage(
		connection *websocket.Conn,
		messageType websocket.MessageType,
		data []byte,
	)
}

type MessageHandlerImpl struct {
	log            *zap.SugaredLogger
	gameController game.Controller
}

func New(log *zap.SugaredLogger, gameController game.Controller) *MessageHandlerImpl {
	return &MessageHandlerImpl{log: log, gameController: gameController}
}

func (m *MessageHandlerImpl) OnMessage(
	connection *websocket.Conn,
	messageType websocket.MessageType,
	data []byte,
) {
	var message api.Message

	m.log.Infow("Received message", "messageType", messageType, "data", data)

	err := json.Unmarshal(data, &message)
	if err != nil {
		m.log.Info("Unable to unmarshal json: %v", err)

		return
	}

	switch message.Type {
	case api.GameStateRequestType:
		var request game2.GameStateRequest

		err := json.Unmarshal(message.Payload, &request)
		if err != nil {
			m.log.Infow("Unable to unmarshal json: %v", "error", err)

			return
		}

		response := m.gameController.GetGameState(request)

		responseByteMessage, err := buildResponseMessage(response, api.GameStateResponseType)
		if err != nil {
			m.log.Errorw("Unable to build response message: %v", "error", err)

			return
		}

		m.log.Infow("Sending response:", "response", response)

		err = connection.WriteMessage(websocket.BinaryMessage, responseByteMessage)
		if err != nil {
			m.log.Errorw("Unable to write message: %v", "error", err)

			return
		}
	case api.GameStateResponseType:
		m.log.Infow("Received response:", "response", message)
	}
}

func buildResponseMessage(payload interface{}, messageType api.Type) ([]byte, error) {
	var message api.Message
	message.Type = messageType

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	message.Payload = data

	result, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return result, nil
}
