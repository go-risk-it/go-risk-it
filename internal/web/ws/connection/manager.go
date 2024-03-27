package connection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/signals"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"go.uber.org/zap"
)

type Manager interface {
	ConnectPlayer(connection *websocket.Conn, gameID int64)
	DisconnectPlayer(connection *websocket.Conn, gameID int64)
	Broadcast(gameID int64, message json.RawMessage)
}

type ManagerImpl struct {
	log                   *zap.SugaredLogger
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
		gameConnections:       make(map[int64][]*websocket.Conn),
		fetchers:              fetchers,
		playerConnectedSignal: playerConnectedSignal,
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
	m.log.Infow(
		"Disconnecting player",
		"remoteAddress", connection.RemoteAddr().String())

	gameConnections := m.gameConnections[gameID]

	index, err := findIndexToRemove(connection, gameConnections)
	if err != nil {
		m.log.Errorw(err.Error())

		return
	}

	m.gameConnections[gameID] = removeIndex(gameConnections, index)
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

	m.gameConnections[gameID] = append(m.gameConnections[gameID], connection)
	m.playerConnectedSignal.Emit(context.Background(), signals.PlayerConnectedData{
		Connection: connection,
		GameID:     gameID,
	})
}

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)

	return append(ret, s[index+1:]...)
}
