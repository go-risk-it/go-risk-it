package events

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// NewBusForTest creates a Bus without fx lifecycle for unit testing.
func NewBusForTest() Bus {
	return &busImpl{
		typedH:  make(map[string][]Handler),
		timeout: defaultHandlerTimeout,
		metrics: nil,
	}
}

// DetachContextForTest exposes detachContext for testing.
func DetachContextForTest(
	parent context.Context,
	event Event,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	return detachContext(parent, event, timeout)
}

// StartLinkedSpanForTest exposes startLinkedSpan for testing.
func StartLinkedSpanForTest(
	parent context.Context,
	spanName string,
) (context.Context, trace.Span) {
	return startLinkedSpan(parent, spanName)
}

// CollectHandlersForTest exposes collectHandlers for testing handler ordering.
func CollectHandlersForTest(bus Bus, eventType string) []Handler {
	impl, ok := bus.(*busImpl)
	if !ok {
		panic("CollectHandlersForTest: bus is not *busImpl")
	}

	return impl.collectHandlers(eventType)
}
