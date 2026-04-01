package runner

import (
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
)

// MetricsHandler records all performance metrics from events.
type MetricsHandler struct {
	collector       *metrics.Collector
	lastRESTEndTime time.Time
	lastPhase       metrics.Phase
	moveStartTime   time.Time
}

// Register subscribes to all events that produce metrics.
func (h *MetricsHandler) Register(bus *Bus) {
	bus.On(EventMoveDecided, h.handleMoveDecided)
	bus.On(EventMoveSucceeded, h.handleMoveSucceeded)
	bus.On(EventStateReceived, h.handleStateReceived)
	bus.On(EventMoveConflict, h.handleMoveConflict)
	bus.On(EventMoveFailed, h.handleMoveFailed)
	bus.On(EventGameComplete, h.handleGameComplete)
}

func (h *MetricsHandler) handleMoveDecided(_ *Bus, e Event) {
	evt := e.(MoveDecidedEvent)
	h.moveStartTime = time.Now()

	if evt.Phase != h.lastPhase {
		h.collector.RecordPhaseEntry(evt.Phase)
		h.lastPhase = evt.Phase
	}
}

func (h *MetricsHandler) handleMoveSucceeded(_ *Bus, e Event) {
	evt := e.(MoveSucceededEvent)

	actionName := actionTypeName(evt.Action.Type)
	h.collector.RecordREST(actionName, evt.RESTLatency)
	h.collector.RecordMove()
	h.collector.RecordTimedMove()
	h.collector.RecordPhaseMove(h.lastPhase)

	if !h.moveStartTime.IsZero() {
		e2e := time.Since(h.moveStartTime)
		h.collector.RecordE2E(e2e)
		h.collector.RecordPhaseLatency(string(h.lastPhase), e2e)
	}

	h.lastRESTEndTime = evt.RESTEndTime
}

func (h *MetricsHandler) handleStateReceived(_ *Bus, e Event) {
	evt := e.(StateReceivedEvent)

	if !h.lastRESTEndTime.IsZero() && evt.Timestamp.After(h.lastRESTEndTime) {
		h.collector.RecordWSDelivery(evt.Timestamp.Sub(h.lastRESTEndTime))
	}
}

func (h *MetricsHandler) handleMoveConflict(_ *Bus, _ Event) {
	h.collector.RecordConflict()
}

func (h *MetricsHandler) handleMoveFailed(_ *Bus, e Event) {
	evt := e.(MoveFailedEvent)
	h.collector.RecordError()
	h.collector.RecordErrorType(evt.ErrType)
}

func (h *MetricsHandler) handleGameComplete(_ *Bus, e Event) {
	evt := e.(GameCompleteEvent)
	result := evt.Result

	switch {
	case result.FatalError != nil:
		h.collector.RecordGameFatal()
	case result.TimedOut:
		h.collector.RecordGameTimedOut(result.Duration, result.Moves)
	default:
		h.collector.RecordGameComplete(result.Duration, result.Moves)
	}
}
