package connection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

type Manager interface {
	ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn)
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
	DisconnectPlayer(connection *websocket.Conn, gameID int64)
}

type ManagerImpl struct {
	log                   *zap.SugaredLogger
	gameConnections       map[int64][]*websocket.Conn
	fetchers              []fetcher.Fetcher
	playerConnectedSignal signals.PlayerConnectedSignal
}

var _ Manager = (*ManagerImpl)(nil)

func NewManager(
	fetchers []fetcher.Fetcher,
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

func (m *ManagerImpl) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	gameConnections, ok := m.gameConnections[ctx.GameID()]
	if !ok {
		ctx.Log().Errorw("no connections for given game")

		return
	}

	if len(gameConnections) == 0 {
		ctx.Log().Warnw("no connections for given game")

		return
	}

	ctx.Log().Infof("broadcasting message to %d players", len(gameConnections))

	for i := range gameConnections {
		err := gameConnections[i].WriteMessage(websocket.TextMessage, message)
		if err != nil {
			ctx.Log().Errorw("unable to write message", "error", err)
		}
	}
}

func (m *ManagerImpl) DisconnectPlayer(connection *websocket.Conn, gameID int64) {
	m.log.Infow(
		"Disconnecting player",
		"remoteAddress", connection.RemoteAddr().String())

	gameConnections, ok := m.gameConnections[gameID]
	if !ok {
		m.log.Warnw("no connections for given game", "gameId", gameID)

		return
	}

	addresses := make([]string, 0, len(gameConnections))

	for i := range gameConnections {
		addresses = append(addresses, gameConnections[i].RemoteAddr().String())
	}

	m.log.Debugw("Found connections", "connections", addresses)

	index, err := findIndexToRemove(connection, gameConnections)
	if err != nil {
		m.log.Errorw(err.Error())

		return
	}

	m.gameConnections[gameID] = removeIndex(m.gameConnections[gameID], index)
	m.log.Infow("Disconnected player", "currentConnections", len(m.gameConnections[gameID]))
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

func (m *ManagerImpl) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	ctx.Log().Infow(
		"Connecting player",
		"remoteAddress", connection.RemoteAddr().String())

	m.gameConnections[ctx.GameID()] = append(m.gameConnections[ctx.GameID()], connection)

	m.playerConnectedSignal.Emit(context.Background(), signals.PlayerConnectedData{
		Connection: connection,
		GameID:     ctx.GameID(),
	})
	ctx.Log().Infow("Connected player", "currentConnections", len(m.gameConnections[ctx.GameID()]))
}

func removeIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)

	return append(ret, s[index+1:]...)
}
