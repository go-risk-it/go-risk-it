package ws_test

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func testMetrics(t *testing.T) *metrics.Metrics {
	t.Helper()

	m, err := metrics.NewMetrics(metricnoop.Meter{})
	require.NoError(t, err)

	return m
}

func gameContext(gameID int64) ctx.GameContext {
	userContext := ctx.WithUserID(
		ctx.WithSpan(
			ctx.WithLog(context.Background(), zap.NewNop().Sugar()),
			noop.Span{},
		),
		"test-user",
	)

	return ctx.WithGameID(userContext, gameID)
}

func TestManagerImpl_GetConnectedPlayers_ConcurrentSameGame(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		manager := ws.NewManager(nil, nil, nil, testMetrics(t))

		const numGoroutines = 100

		gameID := int64(42)

		// All goroutines hit GetConnectedPlayers for the same game ID concurrently.
		// Internally this calls playerConnections() which must safely create
		// exactly one PlayerConnections instance.
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

		// Each goroutine uses a different game ID — should create separate instances.
		for idx := range numGoroutines {
			go func() {
				gameCtx := gameContext(int64(idx))
				_ = manager.GetConnectedPlayers(gameCtx)
			}()
		}

		synctest.Wait()

		// Verify each game has its own (empty) player list without races.
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
