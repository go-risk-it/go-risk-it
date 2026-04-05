package handlers_test

import (
	"context"
	"sync"
	"testing"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// Shared test constants.
const (
	testGameID   = int64(42)
	testAttacker = "attacker-user"
	testDefender = "defender-user"
	testTurn     = int64(7)
)

//nolint:gochecknoglobals // test-only
var fixedTime = time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)

// gameCtx creates a minimal GameContext carrying testGameID and testAttacker
// as the userID. Used by handler tests that emit events via a reentrantBus.
func gameCtx() gamectx.GameContext {
	return gamectx.WithGameID(
		kernelctx.WithUserID(
			kernelctx.WithSpan(context.Background(), noop.Span{}),
			testAttacker,
		),
		testGameID,
	)
}

// reentrantBus is a synchronous test bus that dispatches handlers inline on Emit.
// This allows handler→bus→handler re-entrant chains (e.g., detector emits derived
// headline events) to run deterministically in a single goroutine.
type reentrantBus struct {
	mu       sync.Mutex
	handlers map[string][]eventbus.Handler
	events   []eventbus.Event
}

func newReentrantBus() *reentrantBus {
	return &reentrantBus{
		handlers: make(map[string][]eventbus.Handler),
	}
}

func (b *reentrantBus) OnType(eventType string, handler eventbus.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *reentrantBus) OnAll(handler eventbus.Handler) {
	// Not used in tests, but satisfies the interface.
	_ = handler
}

func (b *reentrantBus) Emit(ctx context.Context, event eventbus.Event) {
	b.mu.Lock()
	b.events = append(b.events, event)
	handlers := append([]eventbus.Handler(nil), b.handlers[event.EventType()]...)
	b.mu.Unlock()

	for _, h := range handlers {
		h(ctx, event)
	}
}

func (b *reentrantBus) allEvents() []eventbus.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	return append([]eventbus.Event(nil), b.events...)
}

var (
	_ eventbus.Subscriber = (*reentrantBus)(nil)
	_ eventbus.Publisher  = (*reentrantBus)(nil)
)

// eventsOfType filters the bus's recorded events by concrete type.
func eventsOfType[E eventbus.Event](bus *reentrantBus) []E {
	var result []E

	for _, evt := range bus.allEvents() {
		if typed, ok := evt.(E); ok {
			result = append(result, typed)
		}
	}

	return result
}

// testContinents builds a board.Continents from a map of continent name → region IDs.
func testContinents(t *testing.T, defs map[string][]string) board.Continents {
	t.Helper()

	var regions []board.RegionDto

	continentDtos := make([]board.ContinentDto, 0, len(defs))

	for name, regionIDs := range defs {
		continentDtos = append(continentDtos, board.ContinentDto{
			ExternalReference: name,
			Name:              name,
			BonusTroops:       0,
		})

		for _, r := range regionIDs {
			regions = append(regions, board.RegionDto{
				ExternalReference: r,
				Name:              r,
				Continent:         name,
			})
		}
	}

	continents, err := board.NewContinents(&board.BoardDto{
		Regions:    regions,
		Continents: continentDtos,
	})
	require.NoError(t, err)

	return continents
}

// defaultContinents returns a standard 2-continent test map.
func defaultContinents() map[string][]string {
	return map[string][]string{
		"europe": {"france", "germany", "italy"},
		"asia":   {"china", "japan"},
	}
}
