package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

type Manager interface {
	ConnectPlayer(connection *websocket.Conn, gameID int64)
	DisconnectPlayer(connection *websocket.Conn, gameID int64)
	Broadcast(gameID int64, message json.RawMessage)
}

type ManagerImpl struct {
	log                   *zap.SugaredLogger
	connectionRWMutex     *sync.RWMutex
	gameConnections       map[int64][]*websocket.Conn
	fetchers              []fetchers.Fetcher
	playerConnectedSignal signals.PlayerConnectedSignal
}

func NewManager(
	fetchers []fetchers.Fetcher,
	log *zap.SugaredLogger,
	playerConnectedSignal signals.PlayerConnectedSignal,
) *ManagerImpl {
	return &ManagerImpl{
		log:                   log,
		connectionRWMutex:     &sync.RWMutex{},
		gameConnections:       make(map[int64][]*websocket.Conn),
		fetchers:              fetchers,
		playerConnectedSignal: playerConnectedSignal,
	}
}

func (m *ManagerImpl) Broadcast(gameID int64, message json.RawMessage) {
	m.connectionRWMutex.RLock()
	connections, ok := m.gameConnections[gameID]
	m.connectionRWMutex.RUnlock()

	if !ok {
		m.log.Errorw("no connections for given game", "gameId", gameID)

		return
	}

	m.log.Infof(
		"broadcasting message to %d players for game %d",
		len(connections),
		gameID,
	)

	for i := range connections {
		err := connections[i].WriteMessage(websocket.TextMessage, message)
		if err != nil {
			m.log.Errorw("unable to write message", "error", err)
		}
	}
}

func (m *ManagerImpl) DisconnectPlayer(connection *websocket.Conn, gameID int64) {
	m.log.Infow(
		"Disconnecting player",
		"remoteAddress", connection.RemoteAddr().String())

	m.connectionRWMutex.RLock()
	gameConnections := m.gameConnections[gameID]
	m.connectionRWMutex.RUnlock()

	index, err := findIndexToRemove(connection, gameConnections)
	if err != nil {
		m.log.Errorw(err.Error())

		return
	}

	m.connectionRWMutex.Lock()
	m.gameConnections[gameID] = removeIndex(gameConnections, index)
	m.log.Infow("Disconnected player", "currentConnections", len(m.gameConnections[gameID]))
	m.connectionRWMutex.Unlock()
}

func findIndexToRemove(
	connection *websocket.Conn,
	gameConnections []*websocket.Conn,
) (int, error) {
	for i := range gameConnections {
		if gameConnections[i].RemoteAddr() == connection.RemoteAddr() {
			return i, nil
		}
	}

	return -1, fmt.Errorf("unable to find index to remove")
}

func (m *ManagerImpl) ConnectPlayer(connection *websocket.Conn, gameID int64) {
	m.log.Infow(
		"Connecting player",
		"remoteAddress", connection.RemoteAddr().String(),
		"gameID", gameID)

	m.connectionRWMutex.Lock()
	m.gameConnections[gameID] = append(m.gameConnections[gameID], connection)
	m.connectionRWMutex.Unlock()

	m.playerConnectedSignal.Emit(context.Background(), signals.PlayerConnectedData{
		Connection: connection,
		GameID:     gameID,
	})
	m.log.Infow("Connected player", "currentConnections", len(m.gameConnections[gameID]))
}

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)

	return append(ret, s[index+1:]...)
}
