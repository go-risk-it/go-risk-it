package ws_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func lobbyContext(lobbyID int64) ctx.LobbyContext {
	userContext := ctx.WithUserID(
		ctx.WithSpan(
			ctx.WithLog(context.Background(), zap.NewNop().Sugar()),
			noop.Span{},
		),
		"test-user",
	)

	return ctx.WithLobbyID(userContext, lobbyID)
}

func TestManagerImpl_Broadcast_ConcurrentSameLobby(t *testing.T) {
	t.Parallel()

	manager := ws.NewManager(nil)

	const numGoroutines = 100

	lobbyID := int64(42)

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// All goroutines hit Broadcast for the same lobby ID concurrently.
	// Internally this calls playerConnections() which must safely create
	// exactly one PlayerConnections instance.
	for range numGoroutines {
		go func() {
			defer waitGroup.Done()

			lobbyCtx := lobbyContext(lobbyID)
			manager.Broadcast(lobbyCtx, json.RawMessage("{}"))
		}()
	}

	waitGroup.Wait()
}

func TestManagerImpl_Broadcast_ConcurrentDifferentLobbies(t *testing.T) {
	t.Parallel()

	manager := ws.NewManager(nil)

	const numGoroutines = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(numGoroutines)

	// Each goroutine uses a different lobby ID — should create separate instances.
	for idx := range numGoroutines {
		go func() {
			defer waitGroup.Done()

			lobbyCtx := lobbyContext(int64(idx))
			manager.Broadcast(lobbyCtx, json.RawMessage("{}"))
		}()
	}

	waitGroup.Wait()
}

func TestManagerImpl_Broadcast_MixedConcurrent(t *testing.T) {
	t.Parallel()

	manager := ws.NewManager(nil)

	const (
		numLobbies         = 5
		goroutinesPerLobby = 50
	)

	var waitGroup sync.WaitGroup

	waitGroup.Add(numLobbies * goroutinesPerLobby)

	for lobbyIdx := range numLobbies {
		for range goroutinesPerLobby {
			go func() {
				defer waitGroup.Done()

				lobbyCtx := lobbyContext(int64(lobbyIdx))
				manager.Broadcast(lobbyCtx, json.RawMessage("{}"))
			}()
		}
	}

	waitGroup.Wait()
}
