package ws

import (
	"encoding/json"
	"errors"
	"net"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type PlayerConnections struct {
	upgradablerwmutex.UpgradableRWMutex
	playerConnections map[string]*websocket.Conn
}

func NewPlayerConnections() *PlayerConnections {
	return &PlayerConnections{
		playerConnections: make(map[string]*websocket.Conn),
	}
}

func (p *PlayerConnections) Broadcast(ctx ctx.UserContext, message json.RawMessage) {
	p.UpgradableRLock()
	defer p.UpgradableRUnlock()

	if len(p.playerConnections) == 0 {
		ctx.Log().Warnw("no connections for given game")

		return
	}

	ctx.Log().Infof("broadcasting message to %d players", len(p.playerConnections))

	toCleanup := make([]string, 0)

	for player, connection := range p.playerConnections {
		err := connection.WriteMessage(websocket.TextMessage, message)
		if err != nil && errors.Is(err, net.ErrClosed) {
			ctx.Log().Debugw("unable to write message because connection is closed")

			toCleanup = append(toCleanup, player)
		}
	}

	p.cleanUpConnections(ctx, toCleanup)
}

func (p *PlayerConnections) Write(ctx ctx.UserContext, message json.RawMessage) {
	p.UpgradableRLock()
	defer p.UpgradableRUnlock()

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

	p.UpgradeWLock()

	for _, player := range toCleanup {
		delete(p.playerConnections, player)
	}

	ctx.Log().Debugw("cleaned up connections", "users", toCleanup)
}

func (p *PlayerConnections) ConnectPlayer(ctx ctx.UserContext, connection *websocket.Conn) {
	ctx.Log().Infow(
		"Connecting player",
		"remoteAddress", connection.RemoteAddr().String())

	p.Lock()
	defer p.Unlock()

	if p.playerConnections[ctx.UserID()] != nil {
		ctx.Log().Warnw("player already connected, overwriting")
	}

	p.playerConnections[ctx.UserID()] = connection
	ctx.Log().Infow("Connected player", "currentConnections", len(p.playerConnections))
}

func (p *PlayerConnections) GetConnectedPlayers(ctx ctx.UserContext) []string {
	p.RLock()
	defer p.RUnlock()

	result := make([]string, 0)
	for player := range p.playerConnections {
		result = append(result, player)
	}

	ctx.Log().Debugw("found connected players", "players", result, "count", len(result))

	return result
}
