package ws_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"testing/synctest"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ws"
	mockbus "github.com/go-risk-it/go-risk-it/mocks/internal_/kernel/bus"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"
)

func testMetrics(t *testing.T) *metrics.InfraMetrics {
	t.Helper()

	m, err := metrics.NewInfraMetrics(metricnoop.Meter{})
	require.NoError(t, err)

	return m
}

func lobbyContext(lobbyID int64) ctx.LobbyContext {
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(
			context.Background(),
			noop.Span{},
		),
		"test-user",
	)

	return ctx.WithLobbyID(userContext, lobbyID)
}

// testWebsocketConn creates a minimal websocket.Conn backed by a net.Pipe for
// testing purposes. The remote end is closed immediately — the connection is only
// used so PlayerConnections.ConnectPlayer can call RemoteAddr() without panicking.
func testWebsocketConn(t *testing.T) *websocket.Conn {
	t.Helper()

	server, client := net.Pipe()
	t.Cleanup(func() {
		server.Close()
		client.Close()
	})

	wsConn := &websocket.Conn{Conn: server}

	return wsConn
}

func TestManagerImpl_ConnectPlayer_EmitsLobbyPlayerConnected(t *testing.T) {
	t.Parallel()

	bus := mockbus.NewBus(t)
	manager := ws.NewManager(bus, testMetrics(t))

	lobbyCtx := lobbyContext(int64(42))

	bus.EXPECT().
		Emit(lobbyCtx, mock.MatchedBy(func(e eventbus.Event) bool {
			evt, ok := e.(*lobbyevt.LobbyPlayerConnected)

			return ok && evt.LobbyID() == int64(42) && evt.UserID() == "test-user"
		})).
		Return()

	manager.ConnectPlayer(lobbyCtx, testWebsocketConn(t))
}

func TestManagerImpl_Broadcast_ConcurrentSameLobby(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		bus := eventbus.NewTestBus()
		manager := ws.NewManager(bus, testMetrics(t))

		const numGoroutines = 100

		lobbyID := int64(42)

		// All goroutines hit Broadcast for the same lobby ID concurrently.
		// Internally this calls playerConnections() which must safely create
		// exactly one PlayerConnections instance.
		for range numGoroutines {
			go func() {
				lobbyCtx := lobbyContext(lobbyID)
				manager.Broadcast(lobbyCtx, json.RawMessage("{}"))
			}()
		}

		synctest.Wait()
	})
}

func TestManagerImpl_Broadcast_ConcurrentDifferentLobbies(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		bus := eventbus.NewTestBus()
		manager := ws.NewManager(bus, testMetrics(t))

		const numGoroutines = 100

		// Each goroutine uses a different lobby ID — should create separate instances.
		for idx := range numGoroutines {
			go func() {
				lobbyCtx := lobbyContext(int64(idx))
				manager.Broadcast(lobbyCtx, json.RawMessage("{}"))
			}()
		}

		synctest.Wait()
	})
}

func TestManagerImpl_Broadcast_MixedConcurrent(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		bus := eventbus.NewTestBus()
		manager := ws.NewManager(bus, testMetrics(t))

		const (
			numLobbies         = 5
			goroutinesPerLobby = 50
		)

		for lobbyIdx := range numLobbies {
			for range goroutinesPerLobby {
				go func() {
					lobbyCtx := lobbyContext(int64(lobbyIdx))
					manager.Broadcast(lobbyCtx, json.RawMessage("{}"))
				}()
			}
		}

		synctest.Wait()
	})
}
