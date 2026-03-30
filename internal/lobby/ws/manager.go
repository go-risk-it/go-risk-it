package ws

import (
	"encoding/json"
	"log/slog"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type manager struct {
	connections *ws.ScopeMap[int64]
	bus         eventbus.Bus
	metrics     *metrics.InfraMetrics
}

func NewManager(
	bus eventbus.Bus,
	metrics *metrics.InfraMetrics,
) Manager {
	return &manager{
		connections: ws.NewScopeMap[int64](),
		bus:         bus,
		metrics:     metrics,
	}
}

func (m *manager) ConnectPlayer(ctx ctx.LobbyContext, connection *websocket.Conn) {
	slog.InfoContext(ctx, "connecting player to lobby")

	m.connections.GetOrCreate(ctx.LobbyID(), func() *ws.PlayerConnections {
		return ws.NewPlayerConnections(m.metrics)
	}).ConnectPlayer(ctx, connection)

	m.bus.Emit(ctx, lobbyevt.NewLobbyPlayerConnected(ctx.LobbyID(), ctx.UserID()))
}

func (m *manager) Broadcast(ctx ctx.LobbyContext, message json.RawMessage) {
	connections := m.connections.GetOrCreate(ctx.LobbyID(), func() *ws.PlayerConnections {
		return ws.NewPlayerConnections(m.metrics)
	})

	connections.Broadcast(ctx, message)
}

func (m *manager) WriteMessage(ctx ctx.LobbyContext, message json.RawMessage) {
	connections := m.connections.GetOrCreate(ctx.LobbyID(), func() *ws.PlayerConnections {
		return ws.NewPlayerConnections(m.metrics)
	})

	connections.Write(ctx, message)
}
