package connection

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"go.uber.org/zap"
)

type Manager interface {
	ConnectPlayer(connection *websocket.Conn, gameID int64)
	DisconnectPlayer(connection *websocket.Conn, gameID int64)
	Broadcast(gameID int64, message json.RawMessage)
}

type ManagerImpl struct {
	log             *zap.SugaredLogger
	gameConnections map[int64][]*websocket.Conn
	fetchers        []fetchers.Fetcher
}

func NewManager(
	fetchers []fetchers.Fetcher,
	log *zap.SugaredLogger,
) *ManagerImpl {
	return &ManagerImpl{
		log:             log,
		gameConnections: make(map[int64][]*websocket.Conn),
		fetchers:        fetchers,
	}
}

func (m *ManagerImpl) Broadcast(gameID int64, message json.RawMessage) {
	m.log.Infof(
		"broadcasting message to %d players for game %d",
		len(m.gameConnections[gameID]),
		gameID,
	)

	for i := range m.gameConnections[gameID] {
		err := m.gameConnections[gameID][i].WriteMessage(websocket.TextMessage, message)
		if err != nil {
			m.log.Errorw("unable to write message", "error", err)
		}
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

	NStates := len(m.fetchers)
	stateChannel := make(chan json.RawMessage, NStates)

	for _, fetcher := range m.fetchers {
		go fetcher.FetchState(ctx, 1, stateChannel)
	}

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

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)

	return append(ret, s[index+1:]...)
}
