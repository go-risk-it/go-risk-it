package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/kernel/upgradablerw_mutex"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type manager struct {
	mu upgradablerwmutex.UpgradableRWMutex

	gameStateService state.Service
	playerService    player.Service
	gameConnections  map[int64]*ws.PlayerConnections
	bus              eventbus.Bus
	metrics          *metrics.Metrics
}

func (m *manager) GetConnectedPlayers(ctx ctx.GameContext) []string {
	connections := m.getPlayerConnections(ctx)
	if connections == nil {
		return nil
	}

	return connections.GetConnectedPlayers(ctx)
}

func NewManager(
	gameStateService state.Service,
	playerService player.Service,
	bus eventbus.Bus,
	metrics *metrics.Metrics,
) Manager {
	return &manager{
		gameStateService: gameStateService,
		playerService:    playerService,
		gameConnections:  make(map[int64]*ws.PlayerConnections),
		bus:              bus,
		metrics:          metrics,
	}
}

func (m *manager) Broadcast(ctx ctx.GameContext, message json.RawMessage) {
	connections := m.getPlayerConnections(ctx)
	if connections == nil {
		slog.DebugContext(ctx, "no connections for game, skipping broadcast")

		return
	}

	connections.Broadcast(ctx, message)
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

	m.getOrCreatePlayerConnections(ctx).ConnectPlayer(ctx, connection)

	m.bus.Emit(ctx, gameevt.NewPlayerConnected(ctx.GameID(), ctx.UserID(), time.Now()))
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
	connections := m.getPlayerConnections(ctx)
	if connections == nil {
		slog.DebugContext(ctx, "no connections for game, skipping write")

		return
	}

	connections.Write(ctx, message)
}

// getPlayerConnections returns the PlayerConnections for the given game, or nil
// if the game is not tracked. This is the read-only path: it never creates a new
// entry, so callers must handle a nil return (typically by returning early/no-op).
// Uses UpgradableRLock (not RLock) for race-detector compatibility with Lock()
// used by RemoveGame — UpgradableRWMutex's RLock lacks race annotations.
func (m *manager) getPlayerConnections(ctx ctx.GameContext) *ws.PlayerConnections {
	m.mu.UpgradableRLock()
	defer m.mu.UpgradableRUnlock()

	return m.gameConnections[ctx.GameID()]
}

// getOrCreatePlayerConnections returns the PlayerConnections for the given game,
// creating one if it does not exist. This is the only code path that inserts into
// gameConnections — used exclusively by ConnectPlayer.
func (m *manager) getOrCreatePlayerConnections(ctx ctx.GameContext) *ws.PlayerConnections {
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

// RemoveGame removes a game's connection tracking from the manager. It decrements
// ActiveConnections by the number of tracked players and deletes the map entry.
// It does NOT close any websocket.Conn — connections are left to close naturally
// or via their own read/write error handling.
// No-op with a debug log when the gameID is not tracked.
func (m *manager) RemoveGame(ctx ctx.GameContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connections, ok := m.gameConnections[ctx.GameID()]
	if !ok {
		slog.DebugContext(ctx, "RemoveGame called for untracked game, no-op")

		return
	}

	playerCount := connections.PlayerCount()

	delete(m.gameConnections, ctx.GameID())

	m.metrics.ActiveConnections.Add(ctx, -int64(playerCount))

	slog.InfoContext(ctx, "removed game connections",
		"removedPlayers", playerCount,
	)
}
