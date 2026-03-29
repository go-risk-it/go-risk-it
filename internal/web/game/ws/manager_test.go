package ws_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"testing/synctest"

	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	playerws "github.com/go-risk-it/go-risk-it/internal/web/ws"
	mockplayer "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/player"
	mockstate "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/state"
	mockbus "github.com/go-risk-it/go-risk-it/mocks/internal_/kernel/bus"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"
)

const testPlayerID = "player-1"

func testMetrics(t *testing.T) *metrics.InfraMetrics {
	t.Helper()

	m, err := metrics.NewInfraMetrics(metricnoop.Meter{})
	require.NoError(t, err)

	return m
}

func gameContext(gameID int64) ctx.GameContext {
	userContext := ctx.WithUserID(
		ctx.WithSpan(
			context.Background(),
			noop.Span{},
		),
		"test-user",
	)

	return ctx.WithGameID(userContext, gameID)
}

func defaultGameContext() ctx.GameContext {
	userContext := ctx.WithUserID(
		ctx.WithSpan(
			context.Background(),
			noop.Span{},
		),
		testPlayerID,
	)

	return ctx.WithGameID(userContext, 1)
}

func testWSConn(t *testing.T) *websocket.Conn {
	t.Helper()

	c, _ := net.Pipe()
	t.Cleanup(func() { c.Close() })

	return websocket.NewServerConn(&websocket.Upgrader{}, c, "", false, false)
}

// connectableManager creates a Manager with mocked services that allow ConnectPlayer
// to succeed for the given userID. Returns the manager and cleanup-aware mocks.
func connectableManager(
	t *testing.T,
	metr *metrics.InfraMetrics,
) (ws.Manager, *mockstate.Service, *mockplayer.Service, *mockbus.Bus) {
	t.Helper()

	stateSvc := mockstate.NewService(t)
	playerSvc := mockplayer.NewService(t)
	bus := mockbus.NewBus(t)

	manager := ws.NewManager(stateSvc, playerSvc, bus, metr)

	return manager, stateSvc, playerSvc, bus
}

// expectConnectPlayer sets up mock expectations so that ConnectPlayer succeeds
// for the given userID on the given gameID.
func expectConnectPlayer(
	stateSvc *mockstate.Service,
	playerSvc *mockplayer.Service,
	bus *mockbus.Bus,
) {
	stateSvc.EXPECT().
		GetGameState(mock.Anything).
		Return(&state.Game{}, nil)

	playerSvc.EXPECT().
		GetPlayersState(mock.Anything).
		Return([]sqlc.GetPlayersStateRow{{UserID: testPlayerID}}, nil)

	bus.EXPECT().
		Emit(mock.Anything, mock.Anything).
		Return()
}

// --- Existing concurrency tests ---

func TestManagerImpl_GetConnectedPlayers_ConcurrentSameGame(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager := ws.NewManager(nil, nil, nil, testMetrics(t))

		const numGoroutines = 100

		gameID := int64(42)

		// All goroutines hit GetConnectedPlayers for the same game ID concurrently.
		// getPlayerConnections (read-only) returns nil for untracked games,
		// so this verifies the read path is race-free.
		for range numGoroutines {
			go func() {
				gameCtx := gameContext(gameID)
				_ = manager.GetConnectedPlayers(gameCtx)
			}()
		}

		synctest.Wait()

		// After all goroutines complete, there should be exactly 0 connected
		// players (no one called ConnectPlayer), but no panics or races.
		result := manager.GetConnectedPlayers(gameContext(gameID))
		assert.Empty(t, result)
	})
}

func TestManagerImpl_GetConnectedPlayers_ConcurrentDifferentGames(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager := ws.NewManager(nil, nil, nil, testMetrics(t))

		const numGoroutines = 100

		// Each goroutine uses a different game ID — all return nil (no lazy init).
		for idx := range numGoroutines {
			go func() {
				gameCtx := gameContext(int64(idx))
				_ = manager.GetConnectedPlayers(gameCtx)
			}()
		}

		synctest.Wait()

		// Verify each game returns empty without races.
		for idx := range numGoroutines {
			result := manager.GetConnectedPlayers(gameContext(int64(idx)))
			assert.Empty(t, result, "game %d should have no connected players", idx)
		}
	})
}

func TestManagerImpl_GetConnectedPlayers_MixedConcurrent(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager := ws.NewManager(nil, nil, nil, testMetrics(t))

		const (
			numGames          = 5
			goroutinesPerGame = 50
		)

		for gameIdx := range numGames {
			for range goroutinesPerGame {
				go func() {
					gameCtx := gameContext(int64(gameIdx))
					_ = manager.GetConnectedPlayers(gameCtx)
				}()
			}
		}

		synctest.Wait()

		// Verify each game is independently accessible.
		for gameIdx := range numGames {
			result := manager.GetConnectedPlayers(gameContext(int64(gameIdx)))
			assert.Empty(t, result)
		}
	})
}

// --- New tests for RemoveGame, PlayerCount, and read-only path behavior ---

func TestManager_RemoveGame_RemovesEntry(t *testing.T) {
	t.Parallel()

	manager, stateSvc, playerSvc, bus := connectableManager(t, testMetrics(t))

	gameCtx := defaultGameContext()

	// Set up mocks so ConnectPlayer succeeds.
	expectConnectPlayer(stateSvc, playerSvc, bus)

	// Connect a player — this creates an entry via getOrCreatePlayerConnections.
	manager.ConnectPlayer(gameCtx, testWSConn(t))

	// Verify the player is tracked.
	players := manager.GetConnectedPlayers(gameCtx)
	require.Len(t, players, 1)
	assert.Equal(t, testPlayerID, players[0])

	// Remove the game.
	manager.RemoveGame(gameCtx)

	// After removal, GetConnectedPlayers uses getPlayerConnections (read-only)
	// which returns nil for untracked games → nil result.
	result := manager.GetConnectedPlayers(gameCtx)
	assert.Empty(t, result)
}

func TestManager_RemoveGame_NonExistentGame(t *testing.T) {
	t.Parallel()

	manager := ws.NewManager(nil, nil, nil, testMetrics(t))

	// RemoveGame for an unknown game must not panic — just a debug log no-op.
	assert.NotPanics(t, func() {
		manager.RemoveGame(gameContext(999))
	})
}

func TestManager_RemoveGame_ConcurrentWithGetConnectedPlayers(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager, stateSvc, playerSvc, bus := connectableManager(t, testMetrics(t))

		gameCtx := defaultGameContext()

		expectConnectPlayer(stateSvc, playerSvc, bus)
		manager.ConnectPlayer(gameCtx, testWSConn(t))

		// Run RemoveGame and GetConnectedPlayers concurrently — must not race.
		go func() {
			manager.RemoveGame(gameCtx)
		}()

		go func() {
			_ = manager.GetConnectedPlayers(gameCtx)
		}()

		synctest.Wait()
	})
}

func TestManager_Broadcast_AfterRemoveGame(t *testing.T) {
	t.Parallel()

	manager, stateSvc, playerSvc, bus := connectableManager(t, testMetrics(t))

	gameCtx := defaultGameContext()

	expectConnectPlayer(stateSvc, playerSvc, bus)
	manager.ConnectPlayer(gameCtx, testWSConn(t))

	// Remove the game.
	manager.RemoveGame(gameCtx)

	// Broadcast after removal must not panic — getPlayerConnections returns nil,
	// and the nil check causes an early return.
	assert.NotPanics(t, func() {
		manager.Broadcast(gameCtx, json.RawMessage(`{"type":"test"}`))
	})
}

func TestManager_WriteMessage_AfterRemoveGame(t *testing.T) {
	t.Parallel()

	manager, stateSvc, playerSvc, bus := connectableManager(t, testMetrics(t))

	gameCtx := defaultGameContext()

	expectConnectPlayer(stateSvc, playerSvc, bus)
	manager.ConnectPlayer(gameCtx, testWSConn(t))

	// Remove the game.
	manager.RemoveGame(gameCtx)

	// WriteMessage after removal must not panic — getPlayerConnections returns nil,
	// and the nil check causes an early return.
	assert.NotPanics(t, func() {
		manager.WriteMessage(gameCtx, json.RawMessage(`{"type":"test"}`))
	})
}

func TestManager_PlayerCount(t *testing.T) {
	t.Parallel()

	metr := testMetrics(t)

	// PlayerCount is on PlayerConnections, not Manager, so test directly.
	playerConns := playerws.NewPlayerConnections(metr)

	// Empty connections.
	assert.Equal(t, 0, playerConns.PlayerCount())

	// Connect two players with different user IDs.
	userCtx1 := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
		"user-1",
	)
	userCtx2 := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
		"user-2",
	)

	rawConn1, _ := net.Pipe()
	t.Cleanup(func() { rawConn1.Close() })

	rawConn2, _ := net.Pipe()
	t.Cleanup(func() { rawConn2.Close() })

	conn1 := websocket.NewServerConn(&websocket.Upgrader{}, rawConn1, "", false, false)
	conn2 := websocket.NewServerConn(&websocket.Upgrader{}, rawConn2, "", false, false)

	playerConns.ConnectPlayer(userCtx1, conn1)
	assert.Equal(t, 1, playerConns.PlayerCount())

	playerConns.ConnectPlayer(userCtx2, conn2)
	assert.Equal(t, 2, playerConns.PlayerCount())
}
