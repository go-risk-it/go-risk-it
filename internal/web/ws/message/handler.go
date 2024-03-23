package message

import (
	"encoding/json"
	"fmt"

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
	log *zap.SugaredLogger
}

func NewHandler(
	log *zap.SugaredLogger,
) *HandlerImpl {
	return &HandlerImpl{
		log: log,
	}
}

func (m *HandlerImpl) OnMessage(
	connection *websocket.Conn,
	messageType websocket.MessageType,
	data []byte,
) {
	var requestMessage Message

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

	rawResponse, err := BuildMessage(responseType, response)
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

func (m *HandlerImpl) handleMessage(
	requestMessage Message,
) (interface{}, Type, error) {
	var (
		response     interface{}
		responseType Type
		err          error
	)

	m.log.Infow("Received message", "requestMessage", requestMessage)

	responseType = "dummyType"

	// switch requestMessage.Type {
	// case message.GameStateType:
	//	response, err = handleRequest(requestMessage.Payload, m.gameController.GetGameState)
	//	responseType = message.GameStateResponseType
	// case message.BoardStateType:
	//	response, err = handleRequest(requestMessage.Payload, m.boardController.GetBoardState)
	//	responseType = message.BoardStateResponseType
	// case message.FullStateType:
	//	response, err = handleRequest(requestMessage.Payload, m.gameController.GetFullState)
	//	responseType = message.FullStateResponseType
	//}

	if err != nil {
		var (
			nilResponse     interface{}
			nilResponseType Type
		)

		return nilResponse, nilResponseType, fmt.Errorf("unable to handle message: %w", err)
	}

	return response, responseType, nil
}

// func handleRequest[Request interface{}, Response interface{}](
//	payload json.RawMessage,
//	handlers func(Request) (Response, error),
// ) (Response, error) {
//	var (
//		request     Request
//		nilResponse Response
//	)
//
//	err := json.Unmarshal(payload, &request)
//	if err != nil {
//		return nilResponse, fmt.Errorf("unable to unmarshal json: %w", err)
//	}
//
//	response, err := handlers(request)
//	if err != nil {
//		return nilResponse, fmt.Errorf("unable to handle request: %w", err)
//	}
//
//	return response, nil
//}
