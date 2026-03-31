package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/kernel/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type PlayerConnections struct {
	mu upgradablerwmutex.UpgradableRWMutex

	playerConnections map[string]*websocket.Conn
	metrics           *metrics.StateMetrics
}

func NewPlayerConnections(m *metrics.StateMetrics) *PlayerConnections {
	return &PlayerConnections{
		playerConnections: make(map[string]*websocket.Conn),
		metrics:           m,
	}
}

func (p *PlayerConnections) Broadcast(ctx ctx.UserContext, message json.RawMessage) {
	spanCtx, done := observe.Span(ctx, "ws.broadcast")
	defer done(nil)

	p.mu.UpgradableRLock()
	defer p.mu.UpgradableRUnlock()

	if len(p.playerConnections) == 0 {
		observe.Warn(ctx, "no connections for given game")

		return
	}

	toCleanup := make([]string, 0)
	sent := 0

	for player, connection := range p.playerConnections {
		err := connection.WriteMessage(websocket.TextMessage, message)
		if err != nil && errors.Is(err, net.ErrClosed) {
			toCleanup = append(toCleanup, player)
		} else {
			sent++
		}
	}

	span := trace.SpanFromContext(spanCtx)
	span.SetAttributes(attribute.Int("ws_fanout", len(p.playerConnections)))

	if len(toCleanup) > 0 {
		span.SetStatus(codes.Error, fmt.Sprintf("%d broadcast errors", len(toCleanup)))
	}

	p.cleanUpConnections(ctx, toCleanup)
}

func (p *PlayerConnections) Write(ctx ctx.UserContext, message json.RawMessage) {
	p.mu.UpgradableRLock()
	defer p.mu.UpgradableRUnlock()

	if len(p.playerConnections) == 0 {
		observe.Warn(ctx, "no connections for given game")

		return
	}

	connection, ok := p.playerConnections[ctx.UserID()]
	if !ok {
		observe.Warn(ctx, "no connection for given player")

		return
	}

	err := connection.WriteMessage(websocket.TextMessage, message)
	if err != nil && errors.Is(err, net.ErrClosed) {
		p.cleanUpConnections(ctx, []string{ctx.UserID()})
	}
}

func (p *PlayerConnections) cleanUpConnections(ctx ctx.UserContext, toCleanup []string) {
	if len(toCleanup) == 0 {
		return
	}

	p.mu.UpgradeWLock()

	for _, player := range toCleanup {
		delete(p.playerConnections, player)
	}

	p.metrics.ActiveConnections.Add(ctx, -int64(len(toCleanup)))
}

func (p *PlayerConnections) ConnectPlayer(ctx ctx.UserContext, connection *websocket.Conn) {
	observe.Info(ctx, "connecting player",
		attribute.String("remote_address", connection.RemoteAddr().String()))

	p.mu.Lock()
	defer p.mu.Unlock()

	if existing := p.playerConnections[ctx.UserID()]; existing != nil {
		observe.Warn(ctx, "player already connected, closing old connection")

		_ = existing.Close()
	} else {
		p.metrics.ActiveConnections.Add(ctx, 1)
	}

	p.playerConnections[ctx.UserID()] = connection
	observe.Info(ctx, "connected player",
		attribute.Int("current_connections", len(p.playerConnections)))
}

func (p *PlayerConnections) GetConnectedPlayers(ctx ctx.UserContext) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]string, 0, len(p.playerConnections))
	for player := range p.playerConnections {
		result = append(result, player)
	}

	return result
}

// PlayerCount returns the number of active player connections.
// Used by Manager.RemoveGame to adjust the ActiveConnections metric
// without exposing the internal map.
func (p *PlayerConnections) PlayerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.playerConnections)
}
