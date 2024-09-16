package connection

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	upgradable_rw_mutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type gameConnections struct {
	upgradable_rw_mutex.UpgradableRWMutex
	gameConnections map[int64]*playerConnections
}

func newGameConnections() *gameConnections {
	return &gameConnections{
		gameConnections: make(map[int64]*playerConnections),
	}
}

func (g *gameConnections) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	g.PlayerConnections(ctx).Broadcast(ctx, message)
}

func (g *gameConnections) Write(ctx ctx.GameContext, message json.RawMessage) {
	g.PlayerConnections(ctx).Write(ctx, message)
}

func (g *gameConnections) PlayerConnections(ctx ctx.GameContext) *playerConnections {
	g.UpgradableRLock()
	defer g.UpgradableRUnlock()

	connections, ok := g.gameConnections[ctx.GameID()]
	if !ok {
		connections = newPlayerConnections()

		g.UpgradeWLock()
		g.gameConnections[ctx.GameID()] = connections
	}

	return connections
}

func (g *gameConnections) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	g.PlayerConnections(ctx).ConnectPlayer(ctx, connection)
}
