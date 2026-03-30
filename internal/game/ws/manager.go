package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
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

func (m *manager) GetConnectedPlayers(ctx ctx.GameContext) []string {
	connections := m.connections.Get(ctx.GameID())
	if connections == nil {
		return nil
	}

	return connections.GetConnectedPlayers(ctx)
}

func (m *manager) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	connections := m.connections.Get(ctx.GameID())
	if connections == nil {
		slog.DebugContext(ctx, "no connections for game, skipping broadcast")

		return
	}

	connections.Broadcast(ctx, message)
}

func (m *manager) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	slog.InfoContext(ctx, "connecting player to game")

	m.connections.GetOrCreate(ctx.GameID(), func() *ws.PlayerConnections {
		return ws.NewPlayerConnections(m.metrics)
	}).ConnectPlayer(ctx, connection)

	m.bus.Emit(ctx, gameevt.NewPlayerConnected(ctx.GameID(), ctx.UserID(), time.Now()))
}

func (m *manager) WriteMessage(ctx ctx.GameContext, message json.RawMessage) {
	connections := m.connections.Get(ctx.GameID())
	if connections == nil {
		slog.DebugContext(ctx, "no connections for game, skipping write")

		return
	}

	connections.Write(ctx, message)
}

// RemoveGame removes a game's connection tracking from the manager. It decrements
// ActiveConnections by the number of tracked players and deletes the map entry.
// It does NOT close any websocket.Conn — connections are left to close naturally
// or via their own read/write error handling.
// No-op with a debug log when the gameID is not tracked.
func (m *manager) RemoveGame(ctx ctx.GameContext) {
	connections, ok := m.connections.Remove(ctx.GameID())
	if !ok {
		slog.DebugContext(ctx, "RemoveGame called for untracked game, no-op")

		return
	}

	playerCount := connections.PlayerCount()

	m.metrics.ActiveConnections.Add(ctx, -int64(playerCount))

	slog.InfoContext(ctx, "removed game connections",
		"removedPlayers", playerCount,
	)
}
