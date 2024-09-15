package connection

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Manager interface {
	ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn)
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
}

type ManagerImpl struct {
	gameConnections       *gameConnections
	playerConnectedSignal signals.PlayerConnectedSignal
}

var _ Manager = (*ManagerImpl)(nil)

func NewManager(playerConnectedSignal signals.PlayerConnectedSignal) *ManagerImpl {
	return &ManagerImpl{
		gameConnections:       newGameConnections(),
		playerConnectedSignal: playerConnectedSignal,
	}
}

func (m *ManagerImpl) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	m.gameConnections.Broadcast(ctx, message)
}

func (m *ManagerImpl) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	m.gameConnections.ConnectPlayer(ctx, connection)

	m.playerConnectedSignal.Emit(ctx, signals.PlayerConnectedData{
		Connection: connection,
		GameID:     ctx.GameID(),
	})
}
