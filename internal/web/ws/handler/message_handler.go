package handler

import (
	"encoding/json"
	"fmt"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/web/controllers/board"
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
	log             *zap.SugaredLogger
	gameController  game.Controller
	boardController board.Controller
}

func New(
	log *zap.SugaredLogger,
	gameController game.Controller,
	boardController board.Controller,
) *MessageHandlerImpl {
	return &MessageHandlerImpl{
		log:             log,
		gameController:  gameController,
		boardController: boardController,
	}
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

	response, responseType, err := m.handleMessage(requestMessage)
	if err != nil {
		m.log.Errorf("unable to handle message: %v", err)

		return
	}

	rawResponse, err := buildResponseMessage(response, responseType)
	if err != nil {
		m.log.Errorf("unable to build response: %v", err)

		return
	}

	m.log.Infow("Sending response", "rawResponse", rawResponse)

	err = connection.WriteMessage(websocket.TextMessage, rawResponse)
	if err != nil {
		m.log.Errorf("unable to write response: %v", err)

		return
	}
}

func (m *MessageHandlerImpl) handleMessage(
	requestMessage message.RequestMessage,
) (interface{}, message.ResponseType, error) {
	var (
		response     interface{}
		responseType message.ResponseType
		err          error
	)

	switch requestMessage.Type {
	case message.GameStateRequestType:
		response, err = handleRequest(requestMessage.Payload, m.gameController.GetGameState)
		responseType = message.GameStateResponseType
	case message.BoardStateRequestType:
		response, err = handleRequest(requestMessage.Payload, m.boardController.GetBoardState)
		responseType = message.BoardStateResponseType
	case message.FullStateRequestType:
		response, err = handleRequest(requestMessage.Payload, m.gameController.GetFullState)
		responseType = message.FullStateResponseType
	}

	if err != nil {
		var (
			nilResponse     interface{}
			nilResponseType message.ResponseType
		)

		return nilResponse, nilResponseType, fmt.Errorf("unable to handle message: %w", err)
	}

	return response, responseType, nil
}

func handleRequest[Request interface{}, Response interface{}](
	payload json.RawMessage,
	handler func(Request) (Response, error),
) (Response, error) {
	var (
		request     Request
		nilResponse Response
	)

	err := json.Unmarshal(payload, &request)
	if err != nil {
		return nilResponse, fmt.Errorf("unable to unmarshal json: %w", err)
	}

	response, err := handler(request)
	if err != nil {
		return nilResponse, fmt.Errorf("unable to handle request: %w", err)
	}

	return response, nil
}

func buildResponseMessage(
	payload interface{},
	messageType message.ResponseType,
) (json.RawMessage, error) {
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