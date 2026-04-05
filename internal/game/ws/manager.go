package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.opentelemetry.io/otel/attribute"
)

type manager struct {
	connections *ws.ScopeMap[int64]
	bus         eventbus.Publisher
	metrics     *metrics.StateMetrics
}

func NewManager(
	bus eventbus.Publisher,
	metrics *metrics.StateMetrics,
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

	return connections.GetConnectedPlayers()
}

func (m *manager) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	connections := m.connections.Get(ctx.GameID())
	if connections == nil {
		return
	}

	connections.Broadcast(ctx, message)
}

func (m *manager) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	observe.Info(ctx, "connecting player to game")

	playerConns := m.connections.GetOrCreate(ctx.GameID(), func() *ws.PlayerConnections {
		conns := ws.NewPlayerConnections(m.metrics)
		conns.SetOnDisconnect(m.makeDisconnectHandler(ctx))

		return conns
	})

	playerConns.ConnectPlayer(ctx, connection)

	msg := buildPresenceMessage(ctx.UserID(), messaging.Connected)
	if msg != nil {
		playerConns.BroadcastOthers(ctx.UserID(), msg)
	}

	m.bus.Emit(ctx, gameevt.NewPlayerConnected(ctx.GameID(), ctx.UserID(), time.Now()))
}

func (m *manager) WriteMessage(ctx ctx.GameContext, message json.RawMessage) {
	connections := m.connections.Get(ctx.GameID())
	if connections == nil {
		return
	}

	connections.Write(ctx, message)
}

// RemoveGame removes a game's connection tracking from the manager. It decrements
// ActiveConnections by the number of tracked players and deletes the map entry.
// It does NOT close any websocket.Conn — connections are left to close naturally
// or via their own read/write error handling.
// No-op when the gameID is not tracked.
func (m *manager) RemoveGame(ctx ctx.GameContext) {
	connections, ok := m.connections.Remove(ctx.GameID())
	if !ok {
		return
	}

	playerCount := connections.PlayerCount()

	m.metrics.ActiveConnections.Add(ctx, -int64(playerCount))

	observe.Info(ctx, "removed game connections",
		attribute.Int("removed_players", playerCount),
	)
}

// makeDisconnectHandler returns a DisconnectFunc that broadcasts a disconnect
// presence signal to remaining players. Called under the PlayerConnections
// write lock — writeToOthers is safe because it iterates the map directly.
func (m *manager) makeDisconnectHandler(
	ctx ctx.GameContext,
) ws.DisconnectFunc {
	return func(removedUserID string, writeToOthers func(message []byte)) {
		observe.Info(ctx, "player disconnected, broadcasting presence",
			attribute.String("disconnected_user", removedUserID),
		)

		msg := buildPresenceMessage(removedUserID, messaging.Disconnected)
		if msg != nil {
			writeToOthers(msg)
		}
	}
}

// buildPresenceMessage constructs a playerConnection envelope. Returns nil on
// marshal failure (logged, not propagated — presence is best-effort).
func buildPresenceMessage(userID string, status messaging.ConnectionStatus) json.RawMessage {
	msg, err := messaging.BuildMessage(messaging.PlayerConnectionType, messaging.PresencePayload{
		UserID: userID,
		Status: status,
	})
	if err != nil {
		slog.Error("failed to build presence message",
			"error", err,
			"userId", userID,
			"status", status,
		)

		return nil
	}

	return msg
}
