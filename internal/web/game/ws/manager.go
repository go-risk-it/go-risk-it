package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/upgradablerw_mutex"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Manager interface {
	GetConnectedPlayers(ctx ctx.GameContext) []string
	ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn)
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
	WriteMessage(ctx ctx.GameContext, message json.RawMessage)
}

type manager struct {
	mu upgradablerwmutex.UpgradableRWMutex

	gameStateService      state.Service
	playerService         player.Service
	gameConnections       map[int64]*ws.PlayerConnections
	playerConnectedSignal signals.PlayerConnectedSignal
	metrics               *metrics.Metrics
}

func (m *manager) GetConnectedPlayers(ctx ctx.GameContext) []string {
	return m.playerConnections(ctx).GetConnectedPlayers(ctx)
}

var _ Manager = (*manager)(nil)

func NewManager(
	gameStateService state.Service,
	playerService player.Service,
	playerConnectedSignal signals.PlayerConnectedSignal,
	metrics *metrics.Metrics,
) Manager {
	return &manager{
		gameStateService:      gameStateService,
		playerService:         playerService,
		gameConnections:       make(map[int64]*ws.PlayerConnections),
		playerConnectedSignal: playerConnectedSignal,
		metrics:               metrics,
	}
}

func (m *manager) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	m.playerConnections(ctx).Broadcast(ctx, message)
}

func (m *manager) ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn) {
	slog.InfoContext(ctx, "connecting player to game")

	if err := m.validateConnectionAttempt(ctx); err != nil {
		slog.DebugContext(ctx, "failed to validate connection attempt", "error", err)

		err = connection.WriteClose(1003, "failed to validate connection attempt")
		if err != nil {
			slog.ErrorContext(ctx, "failed to close websocket connection", "error", err)

			return
		}

		return
	}

	m.playerConnections(ctx).ConnectPlayer(ctx, connection)

	m.playerConnectedSignal.Emit(ctx, signals.PlayerConnectedData{})
}

func (m *manager) validateConnectionAttempt(ctx ctx.GameContext) error {
	gameState, err := m.gameStateService.GetGameState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	slog.DebugContext(ctx, "game state", "state", gameState)

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

func (m *manager) WriteMessage(ctx ctx.GameContext, message json.RawMessage) {
	m.playerConnections(ctx).Write(ctx, message)
}

func (m *manager) playerConnections(ctx ctx.GameContext) *ws.PlayerConnections {
	m.mu.UpgradableRLock()
	defer m.mu.UpgradableRUnlock()

	connections, ok := m.gameConnections[ctx.GameID()]
	if !ok {
		m.mu.UpgradeWLock()

		// Re-check after acquiring write lock — another goroutine may have inserted.
		if existing, exists := m.gameConnections[ctx.GameID()]; exists {
			return existing
		}

		connections = ws.NewPlayerConnections(m.metrics)
		m.gameConnections[ctx.GameID()] = connections
	}

	return connections
}
