package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/web/controllers/board"
	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"github.com/tomfran/go-risk-it/internal/web/controllers/player"
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
	log              *zap.SugaredLogger
	boardController  board.Controller
	gameController   game.Controller
	playerController player.Controller
}

func New(
	log *zap.SugaredLogger,
	boardController board.Controller,
	gameController game.Controller,
	playerController player.Controller,
) *MessageHandlerImpl {
	return &MessageHandlerImpl{
		log:              log,
		boardController:  boardController,
		gameController:   gameController,
		playerController: playerController,
	}
}

func (m *MessageHandlerImpl) OnOpen(
	connection *websocket.Conn,
) {
	m.log.Info("OnOpen:", zap.String("remoteAddress", connection.RemoteAddr().String()))

	err := connection.WriteMessage(websocket.TextMessage, []byte("Established connection"))
	if err != nil {
		m.log.Errorf("unable to write response: %v", err)

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const NStates = 3

	stateChannel := make(chan json.RawMessage, NStates)
	go func() {
		err := getState(ctx, m.gameController.GetGameState, "gameState", stateChannel)
		if err != nil {
			m.log.Errorf("unable to get game state: %v", err)
		}
	}()

	go func() {
		err := getState(ctx, m.boardController.GetBoardState, "boardState", stateChannel)
		if err != nil {
			m.log.Errorf("unable to get board state: %v", err)
		}
	}()

	go func() {
		err := getState(ctx, m.playerController.GetPlayerState, "playerState", stateChannel)
		if err != nil {
			m.log.Errorf("unable to get player state: %v", err)
		}
	}()

	for i := 0; i < NStates; i++ {
		select {
		case state := <-stateChannel:
			err := connection.WriteMessage(websocket.TextMessage, state)
			if err != nil {
				m.log.Errorf("unable to write response: %v", err)
			}
		case <-ctx.Done():
			m.log.Errorf("unable to get all states: %v", ctx.Err())

			return
		}
	}
}

func getState[T interface{}](
	ctx context.Context,
	fetcher func(context.Context, int64) (T, error),
	messageType MessageType,
	channel chan json.RawMessage,
) error {
	state, err := fetcher(ctx, 1)
	if err != nil {
		return fmt.Errorf("unable to fetch state for %s: %w", messageType, err)
	}

	rawResponse, err := buildMessage(messageType, state)
	if err != nil {
		return fmt.Errorf("unable to build response: %w", err)
	}

	channel <- rawResponse

	return nil
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
