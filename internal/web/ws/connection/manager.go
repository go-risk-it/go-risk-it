package connection

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Manager interface {
	GetConnectedPlayers(ctx ctx.GameContext) []string
	ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn)
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
	WriteMessage(ctx ctx.GameContext, message json.RawMessage)
}

type ManagerImpl struct {
	upgradablerwmutex.UpgradableRWMutex
	gameStateService      state.Service
	playerService         player.Service
	gameConnections       map[int64]*playerConnections
	playerConnectedSignal signals.PlayerConnectedSignal
}

func (m *ManagerImpl) GetConnectedPlayers(ctx ctx.GameContext) []string {
	return m.playerConnections(ctx).GetConnectedPlayers(ctx)
}

var _ Manager = (*ManagerImpl)(nil)

func NewManager(
	gameStateService state.Service,
	playerService player.Service,
	playerConnectedSignal signals.PlayerConnectedSignal,
) *ManagerImpl {
	return &ManagerImpl{
		gameStateService:      gameStateService,
		playerService:         playerService,
		gameConnections:       make(map[int64]*playerConnections),
		playerConnectedSignal: playerConnectedSignal,
	}
}

func (m *ManagerImpl) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	m.playerConnections(ctx).Broadcast(ctx, message)
}

func (m *ManagerImpl) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	ctx.Log().Infow("connecting player")

	if err := m.validateConnectionAttempt(ctx); err != nil {
		ctx.Log().Debugw("failed to validate connection attempt", "error", err)

		err = connection.WriteClose(1003, "failed to validate connection attempt")
		if err != nil {
			ctx.Log().Errorw("failed to close websocket connection", "error", err)

			return
		}

		return
	}

	m.playerConnections(ctx).ConnectPlayer(ctx, connection)

	m.playerConnectedSignal.Emit(ctx, signals.PlayerConnectedData{})
}

func (m *ManagerImpl) validateConnectionAttempt(ctx ctx.GameContext) error {
	gameState, err := m.gameStateService.GetGameState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	ctx.Log().Debugw("game state", "state", gameState)

	players, err := m.playerService.GetPlayersState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	if !userIsParticipatingInGame(ctx, players) {
		return errors.New("user not in game")
	}

	return nil
}

func userIsParticipatingInGame(ctx ctx.GameContext, players []sqlc.GetPlayersStateRow) bool {
	for _, player := range players {
		if player.UserID == ctx.UserID() {
			return true
		}
	}

	return false
}

func (m *ManagerImpl) WriteMessage(ctx ctx.GameContext, message json.RawMessage) {
	m.playerConnections(ctx).Write(ctx, message)
}

func (m *ManagerImpl) playerConnections(ctx ctx.GameContext) *playerConnections {
	m.UpgradableRLock()
	defer m.UpgradableRUnlock()

	connections, ok := m.gameConnections[ctx.GameID()]
	if !ok {
		connections = newPlayerConnections()

		m.UpgradeWLock()
		m.gameConnections[ctx.GameID()] = connections
	}

	return connections
}
