package ws

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Manager interface {
	ConnectPlayer(ctx ctx.LobbyContext, connection *websocket.Conn)
	Broadcast(ctx ctx.LobbyContext, message json.RawMessage)
	WriteMessage(ctx ctx.LobbyContext, message json.RawMessage)
}

type ManagerImpl struct {
	upgradablerwmutex.UpgradableRWMutex
	lobbyConnections map[int64]*ws.PlayerConnections
}

var _ Manager = (*ManagerImpl)(nil)

func NewManager() *ManagerImpl {
	return &ManagerImpl{
		lobbyConnections: make(map[int64]*ws.PlayerConnections),
	}
}

func (m *ManagerImpl) ConnectPlayer(ctx ctx.LobbyContext, connection *websocket.Conn) {
	panic("implement me")
}

func (m *ManagerImpl) Broadcast(ctx ctx.LobbyContext, message json.RawMessage) {
	panic("implement me")
}

func (m *ManagerImpl) WriteMessage(ctx ctx.LobbyContext, message json.RawMessage) {
	panic("implement me")
}
