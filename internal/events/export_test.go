package events

import (
	"context"
	"time"
)

// NewBusForTest creates a Bus without fx lifecycle for unit testing.
func NewBusForTest() Bus {
	return &busImpl{
		typedH:  make(map[string][]Handler),
		timeout: defaultHandlerTimeout,
	}
}

// DetachContextForTest exposes detachContext for testing.
func DetachContextForTest(
	parent context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	return detachContext(parent, timeout)
}
