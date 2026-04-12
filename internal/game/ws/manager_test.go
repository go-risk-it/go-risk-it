package ws_test

import (
	"context"
	"encoding/json"
	"net"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/ws"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	playerws "github.com/go-risk-it/go-risk-it/internal/web/ws"
	mockbus "github.com/go-risk-it/go-risk-it/mocks/internal_/kernel/bus"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"
)

const testPlayerID = "player-1"

func testMetrics(t *testing.T) *metrics.StateMetrics {
	t.Helper()

	m, err := metrics.NewStateMetrics(metricnoop.Meter{})
	require.NoError(t, err)

	return m
}

func gameContext(gameID int64) ctx.GameContext {
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(
			context.Background(),
			noop.Span{},
		),
		"test-user",
	)

	return ctx.WithGameID(userContext, gameID)
}

func defaultGameContext() ctx.GameContext {
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(
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

// connectableManager creates a Manager with a mock bus publisher that allows ConnectPlayer.
func connectableManager(
	t *testing.T,
	metr *metrics.StateMetrics,
) (ws.Manager, *mockbus.Publisher) {
	t.Helper()

	pub := mockbus.NewPublisher(t)
	manager := ws.NewManager(pub, metr)

	return manager, pub
}

// --- Existing concurrency tests ---

func TestManagerImpl_GetConnectedPlayers_ConcurrentSameGame(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager := ws.NewManager(nil, testMetrics(t))

		const numGoroutines = 100

		gameID := int64(42)

		// All goroutines hit GetConnectedPlayers for the same game ID concurrently.
		// Get (read-only) returns nil for untracked games,
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
		manager := ws.NewManager(nil, testMetrics(t))

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
		manager := ws.NewManager(nil, testMetrics(t))

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

// --- Tests for RemoveGame, PlayerCount, and read-only path behavior ---

func TestManager_RemoveGame_RemovesEntry(t *testing.T) {
	t.Parallel()

	manager, bus := connectableManager(t, testMetrics(t))

	gameCtx := defaultGameContext()

	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	// Connect a player — this creates an entry via GetOrCreate.
	manager.ConnectPlayer(gameCtx, testWSConn(t))

	// Verify the player is tracked.
	players := manager.GetConnectedPlayers(gameCtx)
	require.Len(t, players, 1)
	assert.Equal(t, testPlayerID, players[0])

	// Remove the game.
	manager.RemoveGame(gameCtx)

	// After removal, GetConnectedPlayers returns nil for untracked games.
	result := manager.GetConnectedPlayers(gameCtx)
	assert.Empty(t, result)
}

func TestManager_RemoveGame_NonExistentGame(t *testing.T) {
	t.Parallel()

	manager := ws.NewManager(nil, testMetrics(t))

	// RemoveGame for an unknown game must not panic — just a debug log no-op.
	assert.NotPanics(t, func() {
		manager.RemoveGame(gameContext(999))
	})
}

func TestManager_RemoveGame_ConcurrentWithGetConnectedPlayers(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager, bus := connectableManager(t, testMetrics(t))

		gameCtx := defaultGameContext()

		bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()
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

	manager, bus := connectableManager(t, testMetrics(t))

	gameCtx := defaultGameContext()

	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()
	manager.ConnectPlayer(gameCtx, testWSConn(t))

	// Remove the game.
	manager.RemoveGame(gameCtx)

	// Broadcast after removal must not panic — Get returns nil,
	// and the nil check causes an early return.
	assert.NotPanics(t, func() {
		manager.Broadcast(gameCtx, json.RawMessage(`{"type":"test"}`))
	})
}

func TestManager_WriteMessage_AfterRemoveGame(t *testing.T) {
	t.Parallel()

	manager, bus := connectableManager(t, testMetrics(t))

	gameCtx := defaultGameContext()

	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()
	manager.ConnectPlayer(gameCtx, testWSConn(t))

	// Remove the game.
	manager.RemoveGame(gameCtx)

	// WriteMessage after removal must not panic — Get returns nil,
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
	userCtx1 := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		"user-1",
	)
	userCtx2 := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
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

// --- Presence broadcast tests ---

func gameContextForUser(userID string) ctx.GameContext {
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, 1) // gameID constant from line 364
}

// writableWSConn creates a WS server connection backed by a net.Pipe.
// Uses NewUpgrader() (not bare &Upgrader{}) so the Engine is initialized
// and WriteMessage works without panicking.
// Returns the websocket.Conn for the server side and the raw net.Conn
// for the client/reader side.
func writableWSConn(t *testing.T) (*websocket.Conn, net.Conn) {
	t.Helper()

	server, client := net.Pipe()
	t.Cleanup(func() {
		server.Close()
		client.Close()
	})

	return websocket.NewServerConn(websocket.NewUpgrader(), server, "", false, false), client
}

// readWSMessage reads a single WS frame from the raw pipe and returns
// the payload bytes. It reads the nbio websocket frame format directly.
func readWSMessage(t *testing.T, reader net.Conn) []byte {
	t.Helper()

	// Read up to 4KB — more than enough for a presence message.
	buf := make([]byte, 4096)
	bytesRead, err := reader.Read(buf)
	assert.NoError(t, err) //nolint:testifylint // Called from goroutines - safe with assert
	assert.Positive(
		t,
		bytesRead,
	)

	// nbio writes raw WS frames. Parse the frame to extract the payload.
	// Frame format: [opcode byte] [length byte(s)] [payload]
	frame := buf[:bytesRead]

	// Skip the first byte (FIN + opcode).
	payloadStart := 2
	payloadLen := int(frame[1] & 0x7F)

	if payloadLen == 126 {
		payloadLen = int(frame[2])<<8 | int(frame[3])
		payloadStart = 4
	}

	return frame[payloadStart : payloadStart+payloadLen]
}

func TestManager_ConnectBroadcastsPresence(t *testing.T) {
	t.Parallel()

	manager, bus := connectableManager(t, testMetrics(t))

	// Player A connects first — no one else to receive presence yet.
	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	ctxA := gameContextForUser("player-A")
	connA, clientA := writableWSConn(t)
	manager.ConnectPlayer(ctxA, connA)

	// Player B connects — player A should receive a connected presence signal.
	// ConnectPlayer writes to A's pipe synchronously, so we must read from
	// clientA concurrently to avoid blocking.
	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	received := make(chan []byte, 1)

	go func() {
		received <- readWSMessage(t, clientA)
	}()

	ctxB := gameContextForUser("player-B")
	connB, _ := writableWSConn(t)
	manager.ConnectPlayer(ctxB, connB)

	// Read the presence message from player A's pipe.
	payload := <-received

	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	require.NoError(t, json.Unmarshal(payload, &envelope))
	assert.Equal(t, "playerConnection", envelope.Type)

	var presence struct {
		UserID string `json:"userId"`
		Status string `json:"status"`
	}

	require.NoError(t, json.Unmarshal(envelope.Data, &presence))
	assert.Equal(t, "player-B", presence.UserID)
	assert.Equal(t, "connected", presence.Status)
}

func TestManager_NoSelfPresence(t *testing.T) {
	t.Parallel()

	manager, bus := connectableManager(t, testMetrics(t))

	// Player A connects.
	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	ctxA := gameContextForUser("player-A")
	connA, clientA := writableWSConn(t)
	manager.ConnectPlayer(ctxA, connA)

	// Player B connects — start reading from A concurrently to unblock the
	// synchronous net.Pipe write.
	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	receivedA := make(chan []byte, 1)

	go func() {
		receivedA <- readWSMessage(t, clientA)
	}()

	ctxB := gameContextForUser("player-B")
	connB, clientB := writableWSConn(t)
	manager.ConnectPlayer(ctxB, connB)

	// Player A should have received presence for B.
	payloadA := <-receivedA
	require.Contains(t, string(payloadA), "player-B")

	// Player B should NOT have received any message. Close the server-side
	// connection so a read on clientB returns immediately.
	connB.Close()

	buf := make([]byte, 1)
	_, err := clientB.Read(buf)

	// Expect EOF or closed pipe — no data was written to B.
	assert.Error(t, err)
}

// closableConn wraps a net.Conn and returns net.ErrClosed (instead of
// io.ErrClosedPipe) after Close is called. This matches the behavior of real
// TCP connections and triggers PlayerConnections.Broadcast's cleanup path
// which checks errors.Is(err, net.ErrClosed).
type closableConn struct {
	net.Conn

	closed atomic.Bool
}

func (c *closableConn) Write(b []byte) (int, error) {
	if c.closed.Load() {
		return 0, net.ErrClosed
	}

	return c.Conn.Write(b)
}

func (c *closableConn) Read(b []byte) (int, error) {
	if c.closed.Load() {
		return 0, net.ErrClosed
	}

	return c.Conn.Read(b)
}

func (c *closableConn) Close() error {
	c.closed.Store(true)

	return c.Conn.Close()
}

func TestManager_DisconnectBroadcastsPresence(t *testing.T) {
	t.Parallel()

	manager, bus := connectableManager(t, testMetrics(t))

	// Connect player A with a readable pipe.
	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	ctxA := gameContextForUser("player-A")
	connA, clientA := writableWSConn(t)
	manager.ConnectPlayer(ctxA, connA)

	// Connect player B. Read A's connect-presence concurrently to avoid
	// blocking the synchronous pipe.
	bus.EXPECT().Emit(mock.Anything, mock.Anything).Return()

	drainCh := make(chan []byte, 1)

	go func() {
		drainCh <- readWSMessage(t, clientA)
	}()

	ctxB := gameContextForUser("player-B")

	// Use closableConn so Close() makes writes return net.ErrClosed,
	// matching production TCP behavior and triggering the cleanup path.
	serverB, clientB := net.Pipe()
	t.Cleanup(func() {
		serverB.Close()
		clientB.Close()
	})

	wrappedB := &closableConn{Conn: serverB}
	connB := websocket.NewServerConn(websocket.NewUpgrader(), wrappedB, "", false, false)
	manager.ConnectPlayer(ctxB, connB)

	// Drain the connect-presence message.
	<-drainCh

	// Close player B's connection. Now writes return net.ErrClosed.
	wrappedB.Close()

	// Trigger a broadcast — this will discover B's connection is broken
	// and clean it up, which should send a disconnect presence to A.
	// Read concurrently: A receives the broadcast + disconnect presence.
	messages := make(chan []byte, 2)

	go func() {
		messages <- readWSMessage(t, clientA)
		messages <- readWSMessage(t, clientA)
	}()

	manager.Broadcast(
		gameContextForUser("any"),
		json.RawMessage(`{"type":"test","data":{}}`),
	)

	broadcastPayload := <-messages
	require.Contains(t, string(broadcastPayload), "test")

	disconnectPayload := <-messages

	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	require.NoError(t, json.Unmarshal(disconnectPayload, &envelope))
	assert.Equal(t, "playerConnection", envelope.Type)

	var presence struct {
		UserID string `json:"userId"`
		Status string `json:"status"`
	}

	require.NoError(t, json.Unmarshal(envelope.Data, &presence))
	assert.Equal(t, "player-B", presence.UserID)
	assert.Equal(t, "disconnected", presence.Status)
}
