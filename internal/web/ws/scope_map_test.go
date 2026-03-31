package ws_test

import (
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

func scopeMapTestMetrics(t *testing.T) *metrics.StateMetrics {
	t.Helper()

	m, err := metrics.NewStateMetrics(metricnoop.Meter{})
	require.NoError(t, err)

	return m
}

func TestScopeMap_Get_ReturnsNilForMissingKey(t *testing.T) {
	t.Parallel()

	scopeMap := ws.NewScopeMap[int64]()

	result := scopeMap.Get(42)

	assert.Nil(t, result)
}

func TestScopeMap_GetOrCreate_CreatesOnce(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scopeMap := ws.NewScopeMap[int64]()
		metr := scopeMapTestMetrics(t)

		var callCount atomic.Int32

		factory := func() *ws.PlayerConnections {
			callCount.Add(1)

			return ws.NewPlayerConnections(metr)
		}

		const numGoroutines = 100

		results := make([]*ws.PlayerConnections, numGoroutines)

		for i := range numGoroutines {
			go func() {
				results[i] = scopeMap.GetOrCreate(1, factory)
			}()
		}

		synctest.Wait()

		// Factory must have been called exactly once.
		assert.Equal(t, int32(1), callCount.Load())

		// All goroutines must have received the same instance.
		for i := 1; i < numGoroutines; i++ {
			assert.Same(t, results[0], results[i], "goroutine %d got a different instance", i)
		}
	})
}

func TestScopeMap_GetOrCreate_DifferentKeys(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scopeMap := ws.NewScopeMap[int64]()
		metr := scopeMapTestMetrics(t)

		var callCount atomic.Int32

		factory := func() *ws.PlayerConnections {
			callCount.Add(1)

			return ws.NewPlayerConnections(metr)
		}

		const numKeys = 50

		results := make([]*ws.PlayerConnections, numKeys)

		for idx := range numKeys {
			go func() {
				results[idx] = scopeMap.GetOrCreate(int64(idx), factory)
			}()
		}

		synctest.Wait()

		// Factory called once per unique key.
		assert.Equal(t, int32(numKeys), callCount.Load())

		// Each key has a distinct instance.
		seen := make(map[*ws.PlayerConnections]bool, numKeys)
		for idx, pc := range results {
			require.NotNil(t, pc, "key %d returned nil", idx)
			assert.False(t, seen[pc], "key %d shares instance with another key", idx)
			seen[pc] = true
		}
	})
}

func TestScopeMap_Remove_ReturnsEntryAndTrue(t *testing.T) {
	t.Parallel()

	scopeMap := ws.NewScopeMap[int64]()
	metr := scopeMapTestMetrics(t)

	created := scopeMap.GetOrCreate(1, func() *ws.PlayerConnections {
		return ws.NewPlayerConnections(metr)
	})

	removed, ok := scopeMap.Remove(1)

	assert.True(t, ok)
	assert.Same(t, created, removed)

	// After removal, Get returns nil.
	assert.Nil(t, scopeMap.Get(1))
}

func TestScopeMap_Remove_ReturnsFalseForMissingKey(t *testing.T) {
	t.Parallel()

	scopeMap := ws.NewScopeMap[int64]()

	removed, ok := scopeMap.Remove(999)

	assert.False(t, ok)
	assert.Nil(t, removed)
}

func TestScopeMap_ConcurrentGetRemove(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scopeMap := ws.NewScopeMap[int64]()
		metr := scopeMapTestMetrics(t)

		// Pre-populate a key.
		scopeMap.GetOrCreate(1, func() *ws.PlayerConnections {
			return ws.NewPlayerConnections(metr)
		})

		const numReaders = 50

		// Concurrent Gets and a single Remove must not race.
		for range numReaders {
			go func() {
				_ = scopeMap.Get(1)
			}()
		}

		go func() {
			scopeMap.Remove(1)
		}()

		synctest.Wait()

		// After everything settles, key is gone.
		assert.Nil(t, scopeMap.Get(1))
	})
}

func TestScopeMap_MixedOperations(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scopeMap := ws.NewScopeMap[int64]()
		metr := scopeMapTestMetrics(t)

		factory := func() *ws.PlayerConnections {
			return ws.NewPlayerConnections(metr)
		}

		const numKeys = 10

		// Phase 1: concurrent GetOrCreate for numKeys keys.
		for idx := range numKeys {
			go func() {
				scopeMap.GetOrCreate(int64(idx), factory)
			}()
		}

		synctest.Wait()

		// Phase 2: concurrent Get, GetOrCreate, and Remove across keys.
		for idx := range numKeys {
			go func() {
				_ = scopeMap.Get(int64(idx))
			}()

			go func() {
				scopeMap.GetOrCreate(int64(idx), factory)
			}()

			if idx%2 == 0 {
				go func() {
					scopeMap.Remove(int64(idx))
				}()
			}
		}

		synctest.Wait()

		// Odd keys should still be present (never removed).
		for idx := 1; idx < numKeys; idx += 2 {
			assert.NotNil(t, scopeMap.Get(int64(idx)), "odd key %d should still exist", idx)
		}
	})
}
