package runner

import (
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
)

// MetricsHandler records HDR histogram and atomic counter metrics from events.
// All timing-dependent measurements (E2E, REST, WS delivery, phase latency)
// are handled by TracingHandler via spans.
type MetricsHandler struct {
	collector *metrics.Collector
}

// Register subscribes to events that produce HDR/counter metrics.
func (h *MetricsHandler) Register(bus *Bus) {
	bus.On(EventMoveDecided, h.handleMoveDecided)
	bus.On(EventMoveSucceeded, h.handleMoveSucceeded)
	bus.On(EventMoveConflict, h.handleMoveConflict)
	bus.On(EventMoveFailed, h.handleMoveFailed)
	bus.On(EventGameComplete, h.handleGameComplete)
}

func (h *MetricsHandler) handleMoveDecided(_ *Bus, e Event) {
	evt := e.(MoveDecidedEvent)
	h.collector.RecordPhaseEntry(evt.Phase)
}

func (h *MetricsHandler) handleMoveSucceeded(_ *Bus, e Event) {
	evt := e.(MoveSucceededEvent)

	h.collector.RecordMove()
	h.collector.RecordTimedMove()
	h.collector.RecordPhaseMove(metrics.Phase(actionTypeName(evt.Action.Type)))
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
