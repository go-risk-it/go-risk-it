package connection

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/lib/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type playerConnections struct {
	upgradablerwmutex.UpgradableRWMutex
	playerConnections map[string]*websocket.Conn
}

func newPlayerConnections() *playerConnections {
	return &playerConnections{
		playerConnections: make(map[string]*websocket.Conn),
	}
}

func (p *playerConnections) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	if len(p.playerConnections) == 0 {
		ctx.Log().Warnw("no connections for given game")

		return
	}

	ctx.Log().Infof("broadcasting message to %d players", len(p.playerConnections))

	for i := range p.playerConnections {
		err := p.playerConnections[i].WriteMessage(websocket.TextMessage, message)
		if err != nil {
			ctx.Log().Errorw("unable to write message", "error", err)
		}
	}
}

func (p *playerConnections) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
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
