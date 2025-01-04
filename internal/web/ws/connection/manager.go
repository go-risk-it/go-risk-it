package connection

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Manager interface {
	GetConnectedPlayers(ctx ctx.GameContext) []string
	ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn)
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
	WriteMessage(ctx ctx.GameContext, message json.RawMessage)
}

type ManagerImpl struct {
	upgradablerwmutex.UpgradableRWMutex
	gameConnections       map[int64]*playerConnections
	playerConnectedSignal signals.PlayerConnectedSignal
}

func (m *ManagerImpl) GetConnectedPlayers(ctx ctx.GameContext) []string {
	return m.playerConnections(ctx).GetConnectedPlayers(ctx)
}

var _ Manager = (*ManagerImpl)(nil)

func NewManager(playerConnectedSignal signals.PlayerConnectedSignal) *ManagerImpl {
	return &ManagerImpl{
		gameConnections:       make(map[int64]*playerConnections),
		playerConnectedSignal: playerConnectedSignal,
	}
}

func (m *ManagerImpl) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	m.playerConnections(ctx).Broadcast(ctx, message)
}

func (m *ManagerImpl) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	m.playerConnections(ctx).ConnectPlayer(ctx, connection)

	m.playerConnectedSignal.Emit(ctx, signals.PlayerConnectedData{})
}

func (m *ManagerImpl) WriteMessage(ctx ctx.GameContext, message json.RawMessage) {
	m.playerConnections(ctx).Write(ctx, message)
}

func (m *ManagerImpl) playerConnections(ctx ctx.GameContext) *playerConnections {
	m.UpgradableRLock()
	defer m.UpgradableRUnlock()

	connections, ok := m.gameConnections[ctx.GameID()]
	if !ok {
		connections = newPlayerConnections()

		m.UpgradeWLock()
		m.gameConnections[ctx.GameID()] = connections
	}

	return connections
}
