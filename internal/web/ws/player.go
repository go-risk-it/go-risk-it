package ws

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type PlayerConnections struct {
	mu upgradablerwmutex.UpgradableRWMutex

	playerConnections map[string]*websocket.Conn
	metrics           *metrics.Metrics
}

func NewPlayerConnections(m *metrics.Metrics) *PlayerConnections {
	return &PlayerConnections{
		playerConnections: make(map[string]*websocket.Conn),
		metrics:           m,
	}
}

func (p *PlayerConnections) Broadcast(ctx ctx.UserContext, message json.RawMessage) {
	start := time.Now()

	p.mu.UpgradableRLock()
	defer p.mu.UpgradableRUnlock()

	if len(p.playerConnections) == 0 {
		ctx.Log().Warnw("no connections for given game")

		return
	}

	ctx.Log().Infof("broadcasting message to %d players", len(p.playerConnections))

	toCleanup := make([]string, 0)
	sent := 0

	for player, connection := range p.playerConnections {
		err := connection.WriteMessage(websocket.TextMessage, message)
		if err != nil && errors.Is(err, net.ErrClosed) {
			ctx.Log().Debugw("unable to write message because connection is closed")

			toCleanup = append(toCleanup, player)
			p.metrics.BroadcastErrors.Add(ctx, 1)
		} else {
			sent++
		}
	}

	p.metrics.MessagesSent.Add(ctx, int64(sent))
	p.metrics.BroadcastDuration.Record(ctx, time.Since(start).Seconds())

	p.cleanUpConnections(ctx, toCleanup)
}

func (p *PlayerConnections) Write(ctx ctx.UserContext, message json.RawMessage) {
	p.mu.UpgradableRLock()
	defer p.mu.UpgradableRUnlock()

	if len(p.playerConnections) == 0 {
		ctx.Log().Warnw("no connections for given game")

		return
	}

	connection, ok := p.playerConnections[ctx.UserID()]
	if !ok {
		ctx.Log().Warnw("no connection for given player")

		return
	}

	ctx.Log().Info("writing message to player", "message", string(message))

	err := connection.WriteMessage(websocket.TextMessage, message)
	if err != nil && errors.Is(err, net.ErrClosed) {
		ctx.Log().Debugw("unable to write message because connection is closed")

		p.cleanUpConnections(ctx, []string{ctx.UserID()})
	}
}

func (p *PlayerConnections) cleanUpConnections(ctx ctx.UserContext, toCleanup []string) {
	if len(toCleanup) == 0 {
		return
	}

	ctx.Log().Debugw("cleaning up connections", "users", toCleanup)

	p.mu.UpgradeWLock()

	for _, player := range toCleanup {
		delete(p.playerConnections, player)
	}

	p.metrics.ActiveConnections.Add(ctx, -int64(len(toCleanup)))

	ctx.Log().Debugw("cleaned up connections", "users", toCleanup)
}

func (p *PlayerConnections) ConnectPlayer(ctx ctx.UserContext, connection *websocket.Conn) {
	ctx.Log().Infow(
		"Connecting player",
		"remoteAddress", connection.RemoteAddr().String())

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.playerConnections[ctx.UserID()] != nil {
		ctx.Log().Warnw("player already connected, overwriting")
	} else {
		p.metrics.ActiveConnections.Add(ctx, 1)
	}

	p.playerConnections[ctx.UserID()] = connection
	ctx.Log().Infow("Connected player", "currentConnections", len(p.playerConnections))
}

func (p *PlayerConnections) GetConnectedPlayers(ctx ctx.UserContext) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]string, 0, len(p.playerConnections))
	for player := range p.playerConnections {
		result = append(result, player)
	}

	ctx.Log().Debugw("found connected players", "players", result, "count", len(result))

	return result
}
