package connection

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/lib/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type gameConnections struct {
	upgradablerwmutex.UpgradableRWMutex
	gameConnections map[int64]*playerConnections
}

func newGameConnections() *gameConnections {
	return &gameConnections{
		gameConnections: make(map[int64]*playerConnections),
	}
}

func (g *gameConnections) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	playerConnections, ok := g.gameConnections[ctx.GameID()]
	if !ok {
		ctx.Log().Errorw("no connections for given game")

		return
	}

	playerConnections.Broadcast(ctx, message)
}

func (g *gameConnections) Write(ctx ctx.GameContext, message json.RawMessage) {
	g.RLock()
	defer g.RUnlock()

	playerConnections, ok := g.gameConnections[ctx.GameID()]
	if !ok {
		ctx.Log().Errorw("no connections for given game")

		return
	}

	playerConnections.Write(ctx, message)
}

func (g *gameConnections) PlayerConnections(gameID int64) *playerConnections {
	g.UpgradableRLock()
	defer g.UpgradableRUnlock()

	connections, ok := g.gameConnections[gameID]
	if !ok {
		connections = newPlayerConnections()

		g.UpgradeWLock()
		g.gameConnections[gameID] = connections
	}

	return connections
}

func (g *gameConnections) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	g.PlayerConnections(ctx.GameID()).ConnectPlayer(ctx, connection)
}
