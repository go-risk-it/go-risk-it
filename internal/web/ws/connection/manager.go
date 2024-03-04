package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/web/controllers/board"
	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"github.com/tomfran/go-risk-it/internal/web/controllers/player"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
	"go.uber.org/zap"
)

type Manager interface {
	ConnectPlayer(connection *websocket.Conn, gameID int64)
	DisconnectPlayer(connection *websocket.Conn, gameID int64)
}

type ManagerImpl struct {
	log              *zap.SugaredLogger
	gameConnections  map[int64][]*websocket.Conn
	boardController  board.Controller
	gameController   game.Controller
	playerController player.Controller
}

func NewManager(log *zap.SugaredLogger,
	boardController board.Controller,
	gameController game.Controller,
	playerController player.Controller,
) *ManagerImpl {
	return &ManagerImpl{
		log:              log,
		gameConnections:  make(map[int64][]*websocket.Conn),
		boardController:  boardController,
		gameController:   gameController,
		playerController: playerController,
	}
}

func (m *ManagerImpl) DisconnectPlayer(connection *websocket.Conn, gameID int64) {
	m.log.Info(
		"Disconnecting player",
		zap.String("remoteAddress", connection.RemoteAddr().String()),
	)

	gameConnections := m.gameConnections[gameID]

	for i := range gameConnections {
		if gameConnections[i] == connection {
			removeIndex(gameConnections, i)
		}
	}
}

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)

	return append(ret, s[index+1:]...)
}

func (m *ManagerImpl) ConnectPlayer(connection *websocket.Conn, gameID int64) {
	m.log.Info(
		"Connecting player",
		zap.String("remoteAddress", connection.RemoteAddr().String()),
		zap.Int64("gameID", gameID),
	)

	m.gameConnections[gameID] = append(m.gameConnections[gameID], connection)
	go m.sendRelevantState(connection)
}

func (m *ManagerImpl) sendRelevantState(connection *websocket.Conn) {
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
	messageType message.Type,
	channel chan json.RawMessage,
) error {
	state, err := fetcher(ctx, 1)
	if err != nil {
		return fmt.Errorf("unable to fetch state for %s: %w", messageType, err)
	}

	rawResponse, err := message.BuildMessage(messageType, state)
	if err != nil {
		return fmt.Errorf("unable to build response: %w", err)
	}

	channel <- rawResponse

	return nil
}
