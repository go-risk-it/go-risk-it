package ws

import (
	"encoding/json"
	"fmt"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/api/game/message"
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
	var requestMessage message.RequestMessage

	m.log.Infow("Received message", "messageType", messageType, "data", data)

	err := json.Unmarshal(data, &requestMessage)
	if err != nil {
		m.log.Info("Unable to unmarshal json: %v", err)

		return
	}

	switch requestMessage.Type {
	case message.GameStateRequestType:
		response, err := handleMessage(requestMessage.Payload, m.gameController.GetGameState)
		if err != nil {
			m.log.Info("Unable to handle message: %v", err)
		}

		rawResponse, err := buildResponseMessage(response, message.GameStateResponseType)
		if err != nil {
			m.log.Errorf("unable to build response: %v", err)

			return
		}

		err = connection.WriteMessage(websocket.BinaryMessage, rawResponse)
		if err != nil {
			m.log.Errorf("unable to write response: %v", err)

			return
		}
	}
}

func handleMessage[Request interface{}, Response interface{}](
	payload json.RawMessage,
	handleRequest func(Request) (Response, error),
) (Response, error) {
	var (
		request     Request
		nilResponse Response
	)

	err := json.Unmarshal(payload, &request)
	if err != nil {
		return nilResponse, fmt.Errorf("unable to unmarshal json: %w", err)
	}

	response, err := handleRequest(request)
	if err != nil {
		return nilResponse, fmt.Errorf("unable to handle request: %w", err)
	}

	return response, nil
}

func buildResponseMessage(payload interface{}, messageType message.ResponseType) ([]byte, error) {
	var responseMessage message.ResponseMessage
	responseMessage.Type = messageType

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	responseMessage.Payload = data

	result, err := json.Marshal(responseMessage)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return result, nil
}
