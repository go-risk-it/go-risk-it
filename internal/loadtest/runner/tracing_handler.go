package runner

import (
	"errors"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"go.opentelemetry.io/otel/attribute"
)

// TracingHandler creates two span levels: a root game span and per-move child spans.
// It must be registered BEFORE MetricsHandler so that session.Ctx carries the
// trace-enriched context for all downstream handlers.

// maxRecordableLatency caps E2E/WS histogram recordings.
// Values above this threshold indicate a WS wait timeout, not real latency.
const maxRecordableLatency = 2 * time.Second

type TracingHandler struct {
	session         *GameSession
	gameDone        func(error)
	moveDone        func(error)
	lastPhase       metrics.Phase
	moveStartTime   time.Time
	lastRESTEndTime time.Time
	retryCount      int
}

// Register subscribes to all events that affect span lifecycle.
func (h *TracingHandler) Register(bus *Bus) {
	bus.On(EventGameStarted, h.handleGameStarted)
	bus.On(EventMoveDecided, h.handleMoveDecided)
	bus.On(EventMoveSucceeded, h.handleMoveSucceeded)
	bus.On(EventMoveConflict, h.handleMoveConflict)
	bus.On(EventMoveFailed, h.handleMoveFailed)
	bus.On(EventStateReceived, h.handleStateReceived)
	bus.On(EventGameComplete, h.handleGameComplete)
}

func (h *TracingHandler) handleGameStarted(_ *Bus, e Event) {
	evt, ok := e.(GameStartedEvent)
	if !ok {
		return
	}

	ctx, done := observe.RawSpan(
		h.session.Ctx,
		"perftest.game.run",
		attribute.Int("gameIndex", evt.GameIndex),
	)
	h.session.Ctx = ctx
	h.gameDone = done
}

func (h *TracingHandler) handleMoveDecided(_ *Bus, e Event) {
	evt, ok := e.(MoveDecidedEvent)
	if !ok {
		return
	}

	// End any previous move span that was not closed (defensive).
	h.endMoveSpan(nil)

	ctx, done := observe.RawSpan(
		h.session.Ctx,
		"perftest.move.execute",
		attribute.String("phase", string(evt.Phase)),
		attribute.String("action", actionTypeName(evt.Action.Type)),
		attribute.String("user_id", evt.UserID),
	)
	h.session.Ctx = ctx
	h.moveDone = done
	h.moveStartTime = time.Now()
	h.lastPhase = evt.Phase
}

func (h *TracingHandler) handleMoveSucceeded(_ *Bus, e Event) {
	evt, ok := e.(MoveSucceededEvent)
	if !ok {
		return
	}

	observe.SpanEvent(
		h.session.Ctx,
		"rest.complete",
		attribute.Float64("rest.duration_ms", float64(evt.RESTLatency.Milliseconds())),
		attribute.String("action", actionTypeName(evt.Action.Type)),
	)

	h.endMoveSpan(nil)
	h.lastRESTEndTime = evt.RESTEndTime
	h.retryCount = 0
}

func (h *TracingHandler) handleMoveConflict(_ *Bus, _ Event) {
	h.endMoveSpan(errors.New("conflict"))
	h.retryCount = 0
}

func (h *TracingHandler) handleMoveFailed(_ *Bus, e Event) {
	evt, ok := e.(MoveFailedEvent)
	if !ok {
		return
	}

	if h.retryCount > 0 {
		observe.SpanEvent(
			h.session.Ctx,
			"retry",
			attribute.Int("retry_count", h.retryCount),
		)
	}

	h.retryCount++

	if h.moveDone != nil {
		observe.SpanEvent(
			h.session.Ctx,
			"move.failed",
			attribute.String("error_type", string(evt.ErrType)),
		)
	}

	h.endMoveSpan(evt.Err)
}

func (h *TracingHandler) handleStateReceived(_ *Bus, e Event) {
	evt, ok := e.(StateReceivedEvent)
	if !ok {
		return
	}

	if !h.lastRESTEndTime.IsZero() && evt.Timestamp.After(h.lastRESTEndTime) {
		wsDelivery := evt.Timestamp.Sub(h.lastRESTEndTime)
		observe.SpanEvent(
			h.session.Ctx,
			"ws.delivery",
			attribute.Float64("ws.delivery.duration_ms", float64(wsDelivery.Milliseconds())),
		)

		// Only record to histogram if within timeout bounds (excludes WS wait timeouts).
		if wsDelivery < maxRecordableLatency {
			h.session.Accumulator.RecordWSDelivery(wsDelivery)
		}
	}

	// Record E2E latency: move decision → WS state update.
	// Exclude timeout cases where the WS wait timed out.
	if !h.moveStartTime.IsZero() && evt.Timestamp.After(h.moveStartTime) {
		e2e := evt.Timestamp.Sub(h.moveStartTime)
		if e2e < maxRecordableLatency {
			h.session.Accumulator.RecordE2E(e2e)
		}
	}
}

func (h *TracingHandler) handleGameComplete(_ *Bus, e Event) {
	evt, ok := e.(GameCompleteEvent)
	if !ok {
		return
	}
	result := evt.Result

	// End any outstanding move span.
	h.endMoveSpan(nil)

	var outcome string

	var gameErr error

	switch {
	case result.FatalError != nil:
		outcome = "fatal"
		gameErr = result.FatalError
	case result.TimedOut:
		outcome = "timedOut"
		gameErr = errors.New("game timed out")
	default:
		outcome = "completed"
	}

	if h.gameDone != nil {
		observe.SpanEvent(
			h.session.Ctx,
			"game.complete",
			attribute.String("outcome", outcome),
			attribute.Int("moves", result.Moves),
		)
		h.gameDone(gameErr)
		h.gameDone = nil
	}
}

// endMoveSpan ends the current move span if one is active.
func (h *TracingHandler) endMoveSpan(err error) {
	if h.moveDone != nil {
		h.moveDone(err)
		h.moveDone = nil
	}
}
