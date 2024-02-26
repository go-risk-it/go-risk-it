package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/web/controllers/board"
	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"go.uber.org/zap"
)

type MessageType string

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"data"`
}

type MessageHandler interface {
	OnMessage(
		connection *websocket.Conn,
		messageType websocket.MessageType,
		data []byte,
	)

	OnOpen(
		connection *websocket.Conn,
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

func (m *MessageHandlerImpl) OnOpen(
	connection *websocket.Conn,
) {
	m.log.Info("OnOpen:", zap.String("remoteAddress", connection.RemoteAddr().String()))
	err := connection.WriteMessage(websocket.TextMessage, []byte("Established connection"))

	for i := 0; i < 100; i++ {
		rawResponse, err := buildMessage("boardState", buildRandomBoardState())
		if err != nil {
			m.log.Errorf("unable to build response: %v", err)

			panic(err)
		}

		err = connection.WriteMessage(websocket.TextMessage, rawResponse)
		if err != nil {
			m.log.Errorf("unable to write message: %v", err)

			panic(err)
		}

		time.Sleep(10 * time.Second)
	}

	if err != nil {
		panic(err)
	}
}

func (m *MessageHandlerImpl) OnMessage(
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

	rawResponse, err := buildMessage(responseType, response)
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
	requestMessage Message,
) (interface{}, MessageType, error) {
	var (
		response     interface{}
		responseType MessageType
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
			nilResponseType MessageType
		)

		return nilResponse, nilResponseType, fmt.Errorf("unable to handle message: %w", err)
	}

	return response, responseType, nil
}

// func handleRequest[Request interface{}, Response interface{}](
//	payload json.RawMessage,
//	handler func(Request) (Response, error),
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
//	response, err := handler(request)
//	if err != nil {
//		return nilResponse, fmt.Errorf("unable to handle request: %w", err)
//	}
//
//	return response, nil
//}

func buildRandomBoardState() message.BoardState {
	nBig, err := rand.Int(rand.Reader, big.NewInt(27))
	if err != nil {
		panic(err)
	}

	if n := nBig.Int64(); n%2 == 0 {
		return message.BoardState{
			Regions: []message.Region{
				{
					ID:      "alaska",
					OwnerID: 1,
					Troops:  10,
				},
				{
					ID:      "ukraine",
					OwnerID: 2,
					Troops:  20,
				},
			},
		}
	} else {
		return message.BoardState{
			Regions: []message.Region{
				{
					ID:      "greenland",
					OwnerID: 2,
					Troops:  10,
				},
				{
					ID:      "congo",
					OwnerID: 1,
					Troops:  20,
				},
			},
		}
	}
}

func buildMessage(
	messageType MessageType,
	payload interface{},
) (json.RawMessage, error) {
	var result Message
	result.Type = messageType

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	result.Payload = data

	rawMessage, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %w", err)
	}

	return rawMessage, nil
}
