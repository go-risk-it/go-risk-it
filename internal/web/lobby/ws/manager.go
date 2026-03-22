package ws

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
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
	mu upgradablerwmutex.UpgradableRWMutex

	lobbyConnections      map[int64]*ws.PlayerConnections
	playerConnectedSignal signals.PlayerConnectedSignal
	metrics               *metrics.Metrics
}

var _ Manager = (*ManagerImpl)(nil)

func NewManager(
	playerConnectedSignal signals.PlayerConnectedSignal,
	metrics *metrics.Metrics,
) *ManagerImpl {
	return &ManagerImpl{
		lobbyConnections:      make(map[int64]*ws.PlayerConnections),
		playerConnectedSignal: playerConnectedSignal,
		metrics:               metrics,
	}
}

func (m *ManagerImpl) ConnectPlayer(ctx ctx.LobbyContext, connection *websocket.Conn) {
	ctx.Log().Info("connecting player to lobby")

	m.playerConnections(ctx).ConnectPlayer(ctx, connection)

	m.playerConnectedSignal.Emit(ctx, signals.PlayerConnectedData{})
}

func (m *ManagerImpl) Broadcast(ctx ctx.LobbyContext, message json.RawMessage) {
	m.playerConnections(ctx).Broadcast(ctx, message)
}

func (m *ManagerImpl) WriteMessage(ctx ctx.LobbyContext, message json.RawMessage) {
	m.playerConnections(ctx).Write(ctx, message)
}

func (m *ManagerImpl) playerConnections(ctx ctx.LobbyContext) *ws.PlayerConnections {
	m.mu.UpgradableRLock()
	defer m.mu.UpgradableRUnlock()

	connections, ok := m.lobbyConnections[ctx.LobbyID()]
	if !ok {
		m.mu.UpgradeWLock()

		// Re-check after acquiring write lock — another goroutine may have inserted.
		if existing, exists := m.lobbyConnections[ctx.LobbyID()]; exists {
			return existing
		}

		connections = ws.NewPlayerConnections(m.metrics)
		m.lobbyConnections[ctx.LobbyID()] = connections
	}

	return connections
}
