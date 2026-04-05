package metrics_test

import (
	"context"
	"testing"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/metrics"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	testGameID = int64(42)
	testUser   = "player1"
	testTurn   = int64(5)
)

var fixedTime = time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

// gameCtx creates a GameContext for the given gameID.
func gameCtx(gameID int64) gamectx.GameContext {
	return gamectx.WithGameID(
		kernelctx.WithUserID(kernelctx.WithSpan(context.Background(), noop.Span{}), testUser),
		gameID,
	)
}

// setupRecorder creates a TestBus, a real OTel meter with an in-memory reader,
// and registers the GameSummaryRecorder. Returns the bus and reader for assertions.
func setupRecorder(t *testing.T) (*eventbus.TestBus, *sdkmetric.ManualReader) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	meter := provider.Meter("test")
	gameMetrics, err := metrics.NewGameMetrics(meter)
	require.NoError(t, err)

	bus := eventbus.NewTestBus()
	metrics.RegisterGameSummaryRecorder(gameMetrics, bus)

	return bus, reader
}

// collectHistograms reads all metrics from the reader and returns histogram data
// keyed by metric name.
func collectHistograms(
	t *testing.T,
	reader *sdkmetric.ManualReader,
) map[string]metricdata.Histogram[float64] {
	t.Helper()

	var resourceMetrics metricdata.ResourceMetrics

	err := reader.Collect(context.Background(), &resourceMetrics)
	require.NoError(t, err)

	result := make(map[string]metricdata.Histogram[float64])

	for _, sm := range resourceMetrics.ScopeMetrics {
		for _, m := range sm.Metrics {
			if h, ok := m.Data.(metricdata.Histogram[float64]); ok {
				result[m.Name] = h
			}
		}
	}

	return result
}

// moveEvent creates a MoveCompleted for the given action type. No phase transition.
func moveEvent(actionType gameapi.GamePhaseType) *gameevt.MoveCompleted {
	return gameevt.NewMoveCompleted(
		testGameID, testUser, fixedTime,
		actionType, testTurn,
		actionType, actionType, // same phase = no transition
		false, nil, nil, nil,
	)
}

// moveEventWithTransition creates a MoveCompleted that represents a phase transition.
func moveEventWithTransition(
	from, to gameapi.GamePhaseType,
) *gameevt.MoveCompleted {
	return gameevt.NewMoveCompleted(
		testGameID, testUser, fixedTime,
		from, testTurn,
		from, to,
		false, nil, nil, nil,
	)
}

// moveEventGameOver creates a MoveCompleted with GameOver=true.
func moveEventGameOver(actionType gameapi.GamePhaseType) *gameevt.MoveCompleted {
	return gameevt.NewMoveCompleted(
		testGameID, testUser, fixedTime,
		actionType, testTurn,
		actionType, actionType,
		true, nil, nil, nil,
	)
}

// emitGameLifecycle emits a GameCreated, then the provided events (MoveCompleted),
// then a final MoveCompleted with GameOver=true.
func emitGameLifecycle(
	bus *eventbus.TestBus,
	events []eventbus.Event,
) {
	ctx := gameCtx(testGameID)

	bus.Emit(ctx, gameevt.NewGameCreated(testGameID, 0, fixedTime, 4, nil, nil))

	for _, e := range events {
		bus.Emit(ctx, e)
	}

	// Final move with GameOver=true triggers histogram recording
	bus.Emit(ctx, moveEventGameOver(gameapi.GamePhaseTypeATTACK))
}

func TestGameSummaryRecorder_AccumulatesMoveCounts(t *testing.T) {
	t.Parallel()

	bus, reader := setupRecorder(t)

	events := []eventbus.Event{
		moveEvent(gameapi.GamePhaseTypeDEPLOY),
		moveEvent(gameapi.GamePhaseTypeATTACK),
		moveEvent(gameapi.GamePhaseTypeCONQUER),
		moveEvent(gameapi.GamePhaseTypeREINFORCE),
		moveEvent(gameapi.GamePhaseTypeCARDS),
	}

	emitGameLifecycle(bus, events)

	histograms := collectHistograms(t, reader)
	moveHist, moveOk := histograms["game.summary.moves"]
	require.True(t, moveOk, "game.summary.moves histogram should exist")
	require.Len(t, moveHist.DataPoints, 1)
	require.Equal(t, uint64(1), moveHist.DataPoints[0].Count)
	// 5 normal moves + 1 final gameOver move = 6
	require.InDelta(t, 6.0, moveHist.DataPoints[0].Sum, 0.001)
}

func TestGameSummaryRecorder_CountsAttacksOnly(t *testing.T) {
	t.Parallel()

	bus, reader := setupRecorder(t)

	events := []eventbus.Event{
		moveEvent(gameapi.GamePhaseTypeDEPLOY),
		moveEvent(gameapi.GamePhaseTypeDEPLOY),
		moveEvent(gameapi.GamePhaseTypeATTACK),
		moveEvent(gameapi.GamePhaseTypeATTACK),
		moveEvent(gameapi.GamePhaseTypeATTACK),
		moveEvent(gameapi.GamePhaseTypeCONQUER),
	}

	emitGameLifecycle(bus, events)

	histograms := collectHistograms(t, reader)

	// attacks histogram: 3 ATTACK moves + 1 final (ATTACK gameOver) = 4
	attackHist, attackOk := histograms["game.summary.attacks"]
	require.True(t, attackOk, "game.summary.attacks histogram should exist")
	require.Len(t, attackHist.DataPoints, 1)
	require.InDelta(t, 4.0, attackHist.DataPoints[0].Sum, 0.001)

	// moves histogram: 6 + 1 final = 7
	moveHist, moveOk := histograms["game.summary.moves"]
	require.True(t, moveOk)
	require.InDelta(t, 7.0, moveHist.DataPoints[0].Sum, 0.001)
}

func TestGameSummaryRecorder_RecordsOnCompletion(t *testing.T) {
	t.Parallel()

	bus, reader := setupRecorder(t)

	gCtx := gameCtx(testGameID)

	// Create game and emit some events without GameOver
	bus.Emit(gCtx, gameevt.NewGameCreated(testGameID, 0, fixedTime, 4, nil, nil))
	bus.Emit(gCtx, moveEvent(gameapi.GamePhaseTypeDEPLOY))
	bus.Emit(
		gCtx,
		moveEventWithTransition(gameapi.GamePhaseTypeDEPLOY, gameapi.GamePhaseTypeATTACK),
	)

	histogramsBefore := collectHistograms(t, reader)
	if h, ok := histogramsBefore["game.summary.moves"]; ok {
		for _, dp := range h.DataPoints {
			require.Equal(t, uint64(0), dp.Count,
				"no histogram data points should be recorded before GameOver")
		}
	}

	// Now emit MoveCompleted with GameOver=true — histograms should be recorded.
	bus.Emit(gCtx, moveEventGameOver(gameapi.GamePhaseTypeATTACK))

	histogramsAfter := collectHistograms(t, reader)
	moveHist, moveOk := histogramsAfter["game.summary.moves"]
	require.True(t, moveOk)
	require.Len(t, moveHist.DataPoints, 1)
	require.Equal(t, uint64(1), moveHist.DataPoints[0].Count)
	require.InDelta(t, 3.0, moveHist.DataPoints[0].Sum, 0.001) // 3 moves total

	turnHist, turnOk := histogramsAfter["game.summary.turns"]
	require.True(t, turnOk)
	require.Len(t, turnHist.DataPoints, 1)
	require.InDelta(t, 1.0, turnHist.DataPoints[0].Sum, 0.001) // 1 phase transition
}

func TestGameSummaryRecorder_CleansUpOnCompletion(t *testing.T) {
	t.Parallel()

	bus, reader := setupRecorder(t)

	// Complete game 1
	emitGameLifecycle(bus, []eventbus.Event{
		moveEvent(gameapi.GamePhaseTypeDEPLOY),
		moveEvent(gameapi.GamePhaseTypeDEPLOY),
	})

	// Complete game 2 (same gameID reused — should start fresh, not accumulate)
	emitGameLifecycle(bus, []eventbus.Event{
		moveEvent(gameapi.GamePhaseTypeDEPLOY),
	})

	histograms := collectHistograms(t, reader)
	moveHist, moveOk := histograms["game.summary.moves"]
	require.True(t, moveOk)
	require.Equal(t, uint64(2), moveHist.DataPoints[0].Count)
	// Game 1: 2 + 1 final = 3. Game 2: 1 + 1 final = 2. Total sum = 5
	require.InDelta(t, 5.0, moveHist.DataPoints[0].Sum, 0.001)
}

func TestGameSummaryRecorder_IgnoresUnknownGame(t *testing.T) {
	t.Parallel()

	bus, reader := setupRecorder(t)

	gCtx := gameCtx(testGameID)

	// Emit events for a game that was never created — should be silently ignored.
	bus.Emit(gCtx, moveEvent(gameapi.GamePhaseTypeDEPLOY))
	bus.Emit(
		gCtx,
		gameevt.NewPlayerEliminated(testGameID, "victim", testUser, fixedTime, testTurn),
	)

	// Emit GameOver for unknown game — should not record anything.
	bus.Emit(gCtx, moveEventGameOver(gameapi.GamePhaseTypeATTACK))

	histograms := collectHistograms(t, reader)
	if h, ok := histograms["game.summary.moves"]; ok {
		for _, dp := range h.DataPoints {
			require.Equal(t, uint64(0), dp.Count,
				"no histogram data points should be recorded for unknown game")
		}
	}
}

func TestNewGameMetrics_CreatesAllHistograms(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	meter := provider.Meter("test")
	gameMetrics, err := metrics.NewGameMetrics(meter)
	require.NoError(t, err)
	require.NotNil(t, gameMetrics)

	ctx := context.Background()
	gameMetrics.SummaryMoves.Record(ctx, 10)
	gameMetrics.SummaryAttacks.Record(ctx, 5)
	gameMetrics.SummaryTurns.Record(ctx, 20)
	gameMetrics.SummaryHeadlines.Record(ctx, 3)

	histograms := collectHistograms(t, reader)

	for _, name := range []string{
		"game.summary.moves",
		"game.summary.attacks",
		"game.summary.turns",
		"game.summary.headlines",
	} {
		hist, found := histograms[name]
		require.True(t, found, "histogram %s should exist", name)
		require.Len(t, hist.DataPoints, 1, "histogram %s should have 1 data point", name)
		require.Equal(
			t,
			uint64(1),
			hist.DataPoints[0].Count,
			"histogram %s should have count=1",
			name,
		)
	}
}
