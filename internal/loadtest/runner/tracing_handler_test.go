package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// setupTestTracer installs an in-memory span exporter and returns it along
// with a cleanup function that restores the previous TracerProvider.
func setupTestTracer(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)

	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	})

	return exporter
}

func makeTracingHandler() (*TracingHandler, *GameSession) {
	gameCtx := &GameSession{
		Ctx:       context.Background(),
		GameIndex: 7,
	}

	h := &TracingHandler{session: gameCtx}

	return h, gameCtx
}

func TestTracingHandler_GameLifecycle(t *testing.T) {
	exporter := setupTestTracer(t)
	h, _ := makeTracingHandler()
	bus := NewTestBus()
	h.Register(bus)

	// Start game.
	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	// Complete game.
	bus.Emit(GameCompleteEvent{Result: GameResult{
		GameIndex: 7,
		Moves:     15,
		Winner:    "u0",
	}})

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "should have one game span")

	gameSpan := spans[0]
	assert.Equal(t, "perftest.game.run", gameSpan.Name)

	// Check game_index attribute.
	assertHasAttr(t, gameSpan.Attributes, attribute.Int("game_index", 7))

	// Check span events for game.complete.
	foundComplete := false

	for _, evt := range gameSpan.Events {
		if evt.Name == "game.complete" {
			foundComplete = true
			assertHasAttrInSlice(t, evt.Attributes, attribute.String("outcome", "completed"))
			assertHasAttrInSlice(t, evt.Attributes, attribute.Int("moves", 15))
		}
	}

	assert.True(t, foundComplete, "should have game.complete span event")
}

func TestTracingHandler_MoveSpan_Success(t *testing.T) {
	exporter := setupTestTracer(t)
	h, _ := makeTracingHandler()
	bus := NewTestBus()
	h.Register(bus)

	// Start game.
	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	// Decide and succeed a move.
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: 42 * time.Millisecond,
		RESTEndTime: time.Now(),
	})

	// End game to flush the game span.
	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 7, Moves: 1}})

	spans := exporter.GetSpans()
	require.Len(t, spans, 2, "should have move span + game span")

	// Move span ends first (index 0).
	moveSpan := spans[0]
	assert.Equal(t, "perftest.move.execute", moveSpan.Name)
	assertHasAttr(t, moveSpan.Attributes, attribute.String("phase", "deploy"))
	assertHasAttr(t, moveSpan.Attributes, attribute.String("action", "deploy"))
	assertHasAttr(t, moveSpan.Attributes, attribute.String("user_id", "u0"))

	// Check rest.complete span event.
	foundREST := false

	for _, evt := range moveSpan.Events {
		if evt.Name == "rest.complete" {
			foundREST = true
		}
	}

	assert.True(t, foundREST, "should have rest.complete span event on move span")
}

func TestTracingHandler_MoveSpan_Conflict(t *testing.T) {
	exporter := setupTestTracer(t)
	h, _ := makeTracingHandler()
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionAttack},
		UserID: "u1",
		Phase:  metrics.PhaseAttack,
	})

	bus.Emit(MoveConflictEvent{Action: &player.Action{Type: player.ActionAttack}})

	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 7}})

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)

	moveSpan := spans[0]
	assert.Equal(t, "perftest.move.execute", moveSpan.Name)

	// Conflict should mark the span with error status.
	assert.Equal(t, "Error", moveSpan.Status.Code.String(),
		"conflict move span should have error status")
}

func TestTracingHandler_MoveSpan_Failed(t *testing.T) {
	exporter := setupTestTracer(t)
	h, _ := makeTracingHandler()
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("stale state"),
		ErrType: metrics.ErrorTypeStaleState,
	})

	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 7}})

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)

	moveSpan := spans[0]
	assert.Equal(t, "perftest.move.execute", moveSpan.Name)
	assert.Equal(t, "Error", moveSpan.Status.Code.String(),
		"failed move span should have error status")

	// Check move.failed event with error_type.
	foundFailed := false

	for _, evt := range moveSpan.Events {
		if evt.Name == "move.failed" {
			foundFailed = true
			assertHasAttrInSlice(t, evt.Attributes, attribute.String("error_type", "stale_state"))
		}
	}

	assert.True(t, foundFailed, "should have move.failed span event")
}

func TestTracingHandler_PhaseTracking(t *testing.T) {
	exporter := setupTestTracer(t)
	h, _ := makeTracingHandler()
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	// First move: deploy phase.
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})
	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: time.Millisecond,
		RESTEndTime: time.Now(),
	})

	// Second move: attack phase.
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionAttack},
		UserID: "u0",
		Phase:  metrics.PhaseAttack,
	})
	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionAttack},
		RESTLatency: time.Millisecond,
		RESTEndTime: time.Now(),
	})

	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 7, Moves: 2}})

	spans := exporter.GetSpans()
	require.Len(t, spans, 3, "2 move spans + 1 game span")

	// First move span has deploy phase.
	assertHasAttr(t, spans[0].Attributes, attribute.String("phase", "deploy"))

	// Second move span has attack phase.
	assertHasAttr(t, spans[1].Attributes, attribute.String("phase", "attack"))
}

func TestTracingHandler_RetryEvent(t *testing.T) {
	exporter := setupTestTracer(t)
	h, _ := makeTracingHandler()
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	// First failure — no retry event (first attempt).
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})
	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("transient"),
		ErrType: metrics.ErrorTypeTransient,
	})

	// Second failure — should have a retry event.
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})
	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("transient again"),
		ErrType: metrics.ErrorTypeTransient,
	})

	bus.Emit(GameCompleteEvent{Result: GameResult{GameIndex: 7}})

	spans := exporter.GetSpans()
	require.GreaterOrEqual(t, len(spans), 3, "at least 2 move spans + game span")

	// Check second move span for retry event.
	secondMoveSpan := spans[1]
	foundRetry := false

	for _, evt := range secondMoveSpan.Events {
		if evt.Name == "retry" {
			foundRetry = true
			assertHasAttrInSlice(t, evt.Attributes, attribute.Int("retry_count", 1))
		}
	}

	assert.True(t, foundRetry, "second move span should have retry span event")
}

// --- helpers ---

func assertHasAttr(
	t *testing.T,
	attrs []attribute.KeyValue,
	want attribute.KeyValue,
) {
	t.Helper()

	for _, a := range attrs {
		if a.Key == want.Key && a.Value == want.Value {
			return
		}
	}

	t.Errorf("expected attribute %v not found in %v", want, attrs)
}

func assertHasAttrInSlice(
	t *testing.T,
	attrs []attribute.KeyValue,
	want attribute.KeyValue,
) {
	t.Helper()

	for _, a := range attrs {
		if a.Key == want.Key && a.Value == want.Value {
			return
		}
	}

	t.Errorf("expected attribute %v not found in %v", want, attrs)
}
