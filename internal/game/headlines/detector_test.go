package headlines_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/headlines"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	mockboard "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/board"
	mocksnapshot "github.com/go-risk-it/go-risk-it/mocks/internal_/game/snapshot"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	testGameID   = int64(42)
	testAttacker = "attacker"
	testDefender = "defender"
	testTurn     = int64(5)
)

var fixedTime = time.Date(2026, 3, 27, 14, 0, 0, 0, time.UTC)

// reentrantBus is a synchronous Bus implementation that supports re-entrant Emit
// calls. Unlike eventbus.TestBus, it does NOT hold a lock during handler dispatch,
// so handlers can safely call Emit on the same bus. This is needed because the
// headline detector emits derived events from within a handler.
type reentrantBus struct {
	mu     sync.Mutex
	events []eventbus.Event
	allH   []eventbus.Handler
	typedH map[string][]eventbus.Handler
}

var (
	_ eventbus.Publisher  = (*reentrantBus)(nil)
	_ eventbus.Subscriber = (*reentrantBus)(nil)
)

func newReentrantBus() *reentrantBus {
	return &reentrantBus{
		typedH: make(map[string][]eventbus.Handler),
	}
}

func (b *reentrantBus) Emit(ctx context.Context, event eventbus.Event) {
	if event == nil {
		panic("events: Emit called with nil event")
	}

	// Snapshot handlers under lock, then dispatch without lock held.
	b.mu.Lock()
	allHandlers := make([]eventbus.Handler, len(b.allH))
	copy(allHandlers, b.allH)

	typedHandlers := make([]eventbus.Handler, len(b.typedH[event.EventType()]))
	copy(typedHandlers, b.typedH[event.EventType()])
	b.mu.Unlock()

	for _, h := range allHandlers {
		h(ctx, event)
	}

	for _, h := range typedHandlers {
		h(ctx, event)
	}

	b.mu.Lock()
	b.events = append(b.events, event)
	b.mu.Unlock()
}

func (b *reentrantBus) OnAll(handler eventbus.Handler) {
	if handler == nil {
		panic("events: OnAll called with nil handler")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.allH = append(b.allH, handler)
}

func (b *reentrantBus) OnType(eventType string, handler eventbus.Handler) {
	if eventType == "" {
		panic("events: OnType called with empty eventType")
	}

	if handler == nil {
		panic("events: OnType called with nil handler")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.typedH[eventType] = append(b.typedH[eventType], handler)
}

func (b *reentrantBus) allEvents() []eventbus.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]eventbus.Event, len(b.events))
	copy(result, b.events)

	return result
}

func (b *reentrantBus) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = nil
}

func eventsOfType[E eventbus.Event](bus *reentrantBus) []E {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	var result []E

	for _, event := range bus.events {
		if typed, ok := event.(E); ok {
			result = append(result, typed)
		}
	}

	return result
}

// gameCtx creates a GameContext for the given gameID with testAttacker as userID.
func gameCtx(gameID int64) ctx.GameContext {
	return ctx.WithGameID(
		kernelctx.WithUserID(kernelctx.WithSpan(context.Background(), noop.Span{}), testAttacker),
		gameID,
	)
}

// testContinents builds a board.Continents from a simple map of continent -> regions.
func testContinents(t *testing.T, continentMap map[string][]string) board.Continents {
	t.Helper()

	regions := make([]board.RegionDto, 0)
	continents := make([]board.ContinentDto, 0, len(continentMap))

	for continentID, regionIDs := range continentMap {
		continents = append(continents, board.ContinentDto{
			ExternalReference: continentID,
			BonusTroops:       2,
		})

		for _, regionID := range regionIDs {
			regions = append(regions, board.RegionDto{
				ExternalReference: regionID,
				Continent:         continentID,
			})
		}
	}

	result, err := board.NewContinents(&board.BoardDto{
		Regions:    regions,
		Continents: continents,
	})
	require.NoError(t, err)

	return result
}

// snapshotBoard builds a PublicSnapshot containing board regions with given ownership.
func snapshotBoard(regionOwnership map[string]string) *snapshot.PublicSnapshot {
	regions := make([]sqlc.GetRegionsByGameRow, 0, len(regionOwnership))
	for region, owner := range regionOwnership {
		regions = append(regions, sqlc.GetRegionsByGameRow{
			ExternalReference: region,
			UserID:            owner,
		})
	}

	return &snapshot.PublicSnapshot{
		Board: regions,
	}
}

// attackEvent creates a MoveExecuted event for an ATTACK with the given result.
func attackEvent(gameID int64, result *attack.MoveResult) *gameevt.MoveExecuted {
	return gameevt.NewMoveExecuted(
		gameID,
		testAttacker,
		fixedTime,
		sqlc.GamePhaseTypeATTACK,
		sqlc.GameMoveLog{ID: 1},
		sqlc.GamePhaseTypeATTACK,
		false,
		testTurn,
		result,
		nil,
	)
}

// setupDetector creates a reentrantBus with the detector registered and returns
// the bus and snapshot mock for configuring expectations.
func setupDetector(
	t *testing.T,
	continentDefs map[string][]string,
) (*reentrantBus, *mocksnapshot.Service) {
	t.Helper()

	bus := newReentrantBus()
	snapshotSvc := mocksnapshot.NewService(t)
	continents := testContinents(t, continentDefs)

	boardSvc := mockboard.NewService(t)
	boardSvc.EXPECT().GetContinents(mock.Anything).Return(continents, nil).Maybe()

	headlines.RegisterDetector(headlines.DetectorParams{
		Pub:      bus,
		Sub:      bus,
		Snapshot: snapshotSvc,
		Board:    boardSvc,
	})

	return bus, snapshotSvc
}

// defaultContinents returns a two-continent map used by most tests:
// "europe" has 3 regions, "asia" has 2 regions.
func defaultContinents() map[string][]string {
	return map[string][]string{
		"europe": {"france", "germany", "italy"},
		"asia":   {"china", "japan"},
	}
}

func TestDetector_IgnoresNonAttackMoves(t *testing.T) {
	t.Parallel()

	nonAttackPhases := []sqlc.GamePhaseType{
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GamePhaseTypeCONQUER,
		sqlc.GamePhaseTypeREINFORCE,
		sqlc.GamePhaseTypeCARDS,
	}

	for _, phase := range nonAttackPhases {
		t.Run(string(phase), func(t *testing.T) {
			t.Parallel()

			bus, _ := setupDetector(t, defaultContinents())

			event := gameevt.NewMoveExecuted(
				testGameID,
				testAttacker,
				fixedTime,
				phase,
				sqlc.GameMoveLog{ID: 1},
				phase,
				false,
				testTurn,
				nil,
				nil,
			)

			bus.Emit(gameCtx(testGameID), event)

			// Only the MoveExecuted event should be captured; no headlines.
			allEvents := bus.allEvents()
			require.Len(t, allEvents, 1)
			require.Equal(t, gameevt.TypeMoveExecuted, allEvents[0].EventType())
		})
	}
}

func TestDetector_IgnoresUnsuccessfulAttack(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// First: a successful attack to initialize the cache
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testDefender,
			"italy":   testDefender,
			"china":   testDefender,
			"japan":   testDefender,
		}), nil)

	initEvent := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "germany",
		ConqueringTroops:  1,
	})
	bus.Emit(gameCtx(testGameID), initEvent)
	bus.reset()

	// Now: attack with ConqueringTroops=0 (failed attack) -> no headlines
	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  0,
	})

	bus.Emit(gameCtx(testGameID), event)

	// Only the MoveExecuted event; no headlines
	allEvents := bus.allEvents()
	require.Len(t, allEvents, 1)
}

func TestDetector_InitsCacheOnFirstEvent(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// Expect exactly one snapshot call for cache init
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testAttacker,
			"china":   testDefender,
			"japan":   testDefender,
		}), nil).
		Once()

	// First event triggers cache init (must be a successful conquest)
	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "china",
		ConqueringTroops:  1,
	})
	bus.Emit(gameCtx(testGameID), event)

	// Second event should NOT call snapshot again (cached)
	event2 := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "japan",
		ConqueringTroops:  1,
	})
	bus.Emit(gameCtx(testGameID), event2)

	// Mock assertions verify GetPublicSnapshot called exactly once
}

func TestDetector_PlayerEliminated(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// Defender owns only one region
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testDefender, // defender's only region
			"china":   testAttacker,
			"japan":   testAttacker,
		}), nil)

	// Attack conquers defender's last region
	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  3,
	})

	bus.Emit(gameCtx(testGameID), event)

	eliminated := eventsOfType[*headlines.PlayerEliminated](bus)
	require.Len(t, eliminated, 1)
	require.Equal(t, testGameID, eliminated[0].GameID())
	require.Equal(t, testDefender, eliminated[0].EliminatedUserID())
	require.Equal(t, testAttacker, eliminated[0].EliminatorUserID())
}

func TestDetector_NoPlayerEliminated(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// Defender owns two regions
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testDefender,
			"italy":   testDefender, // defender still has germany after losing italy
			"china":   testAttacker,
			"japan":   testAttacker,
		}), nil)

	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  3,
	})

	bus.Emit(gameCtx(testGameID), event)

	eliminated := eventsOfType[*headlines.PlayerEliminated](bus)
	require.Empty(t, eliminated)
}

func TestDetector_ContinentCaptured(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// Attacker owns 2 of 3 europe regions, defender has the last one
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testDefender, // last europe region for attacker to capture
			"china":   testDefender,
			"japan":   testDefender,
		}), nil)

	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  3,
	})

	bus.Emit(gameCtx(testGameID), event)

	captured := eventsOfType[*headlines.ContinentCaptured](bus)
	require.Len(t, captured, 1)
	require.Equal(t, "europe", captured[0].ContinentID)
	require.Equal(t, testGameID, captured[0].GameID())
	require.Equal(t, testAttacker, captured[0].UserID())
}

func TestDetector_ContinentLost(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// Defender owns all of asia; attacker takes one
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testAttacker,
			"china":   testDefender,
			"japan":   testDefender, // defender owns all of asia
		}), nil)

	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "china",
		ConqueringTroops:  3,
	})

	bus.Emit(gameCtx(testGameID), event)

	lost := eventsOfType[*headlines.ContinentLost](bus)
	require.Len(t, lost, 1)
	require.Equal(t, "asia", lost[0].ContinentID)
	require.Equal(t, testGameID, lost[0].GameID())
	require.Equal(t, testDefender, lost[0].UserID())
}

func TestDetector_ContinentCapturedAndLost(t *testing.T) {
	t.Parallel()

	// Single-region continent so one conquest both captures and loses
	continentDefs := map[string][]string{
		"island": {"atoll"},
		"big":    {"north", "south"},
	}

	bus, snapshotSvc := setupDetector(t, continentDefs)

	// Defender owns the single-region continent "island"
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"atoll": testDefender, // defender's continent
			"north": testAttacker,
			"south": testAttacker,
		}), nil)

	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "north",
		DefendingRegionID: "atoll",
		ConqueringTroops:  2,
	})

	bus.Emit(gameCtx(testGameID), event)

	captured := eventsOfType[*headlines.ContinentCaptured](bus)
	require.Len(t, captured, 1)
	require.Equal(t, "island", captured[0].ContinentID)

	lost := eventsOfType[*headlines.ContinentLost](bus)
	require.Len(t, lost, 1)
	require.Equal(t, "island", lost[0].ContinentID)
}

func TestDetector_MultipleGamesIndependent(t *testing.T) {
	t.Parallel()

	const (
		game1 = int64(1)
		game2 = int64(2)
	)

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// Game 1: defender has one region
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.MatchedBy(func(gc ctx.GameContext) bool {
			return gc.GameID() == game1
		})).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testDefender,
			"china":   testAttacker,
			"japan":   testAttacker,
		}), nil)

	// Game 2: defender has many regions (no elimination)
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.MatchedBy(func(gc ctx.GameContext) bool {
			return gc.GameID() == game2
		})).
		Return(snapshotBoard(map[string]string{
			"france":  testDefender,
			"germany": testDefender,
			"italy":   testDefender,
			"china":   testAttacker,
			"japan":   testDefender,
		}), nil)

	// Game 1: conquer defender's last region -> elimination
	bus.Emit(gameCtx(game1), attackEvent(game1, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  3,
	}))

	// Game 2: conquer a region but defender still has others -> no elimination
	bus.Emit(gameCtx(game2), attackEvent(game2, &attack.MoveResult{
		AttackingRegionID: "china",
		DefendingRegionID: "japan",
		ConqueringTroops:  3,
	}))

	eliminated := eventsOfType[*headlines.PlayerEliminated](bus)
	require.Len(t, eliminated, 1)
	require.Equal(t, game1, eliminated[0].GameID())
}

func TestDetector_SnapshotErrorLogged(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(nil, errors.New("db connection failed"))

	// Should not panic; event is processed but no headlines emitted
	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "germany",
		ConqueringTroops:  3,
	})

	require.NotPanics(t, func() {
		bus.Emit(gameCtx(testGameID), event)
	})

	// Only the MoveExecuted event; no headlines
	allEvents := bus.allEvents()
	require.Len(t, allEvents, 1)
}

func TestDetector_GameCompletedCleansCache(t *testing.T) {
	t.Parallel()

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	// First call: init cache for the game
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testDefender,
			"china":   testDefender,
			"japan":   testDefender,
		}), nil)

	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  3,
	})
	bus.Emit(gameCtx(testGameID), event)
	bus.reset()

	// GameCompleted cleans up the cache
	completed := gameevt.NewGameCompleted(testGameID, testAttacker, fixedTime, testTurn)
	bus.Emit(gameCtx(testGameID), completed)
	bus.reset()

	// Second call: must re-fetch snapshot (cache was cleared)
	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testAttacker,
			"china":   testDefender,
			"japan":   testDefender,
		}), nil)

	event2 := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "china",
		ConqueringTroops:  1,
	})
	bus.Emit(gameCtx(testGameID), event2)

	// Snapshot was called twice total (once per cache init), proving cache was cleared
	snapshotSvc.AssertNumberOfCalls(t, "GetPublicSnapshot", 2)
}

func TestDetector_IgnoresNilAttackResult(t *testing.T) {
	t.Parallel()

	bus, _ := setupDetector(t, defaultContinents())

	// Attack event with nil AttackResult (distinct from ConqueringTroops=0)
	event := gameevt.NewMoveExecuted(
		testGameID,
		testAttacker,
		fixedTime,
		sqlc.GamePhaseTypeATTACK,
		sqlc.GameMoveLog{ID: 1},
		sqlc.GamePhaseTypeATTACK,
		false,
		testTurn,
		nil, // nil AttackResult
		nil,
	)

	bus.Emit(gameCtx(testGameID), event)

	// Only the MoveExecuted event; no headlines, no snapshot call
	allEvents := bus.allEvents()
	require.Len(t, allEvents, 1)
	require.Equal(t, gameevt.TypeMoveExecuted, allEvents[0].EventType())
}

// ---------------------------------------------------------------------------
// OTel setup helper
// ---------------------------------------------------------------------------

func setupOTelExporter(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)

	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tracerProvider.Shutdown(context.Background())
	})

	return exporter
}

// spanNames extracts span names from stubs for diagnostic messages.
func spanNames(stubs tracetest.SpanStubs) []string {
	names := make([]string, len(stubs))
	for i, s := range stubs {
		names[i] = s.Name
	}

	return names
}

// ---------------------------------------------------------------------------
// Test: MoveExecuted — headline detector creates detector.headlines span
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestDetector_CreatesHeadlineSpan(t *testing.T) {
	// Not t.Parallel() — swaps global TracerProvider.
	exporter := setupOTelExporter(t)

	bus, snapshotSvc := setupDetector(t, defaultContinents())

	snapshotSvc.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(snapshotBoard(map[string]string{
			"france":  testAttacker,
			"germany": testAttacker,
			"italy":   testDefender,
			"china":   testDefender,
			"japan":   testDefender,
		}), nil)

	event := attackEvent(testGameID, &attack.MoveResult{
		AttackingRegionID: "france",
		DefendingRegionID: "italy",
		ConqueringTroops:  3,
	})

	bus.Emit(gameCtx(testGameID), event)

	stubs := exporter.GetSpans()
	var found bool

	for _, stub := range stubs {
		if stub.Name == "detector.headlines" {
			found = true

			break
		}
	}

	require.True(t, found,
		"expected span named 'detector.headlines' in recorded spans, got: %v",
		spanNames(stubs))
}
