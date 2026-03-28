package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/safego"
	"go.uber.org/fx"
)

const defaultHandlerTimeout = 10 * time.Second

// Bus dispatches events to registered handlers. Each handler runs in its own
// goroutine with a detached context. Close gracefully drains in-flight handlers.
type Bus interface {
	// Emit dispatches event to all matching handlers. Each handler runs in a separate
	// goroutine via safego.Go. Panics if event is nil. No-op after Close.
	Emit(ctx context.Context, event Event)

	// OnAll registers a handler that receives every emitted event.
	// Panics if handler is nil.
	OnAll(handler Handler)

	// OnType registers a handler that receives only events matching the given type.
	// Panics if handler is nil or eventType is empty.
	OnType(eventType string, handler Handler)

	// Close gracefully shuts down the bus. It sets the bus to closed state and waits
	// for all in-flight handlers to complete. Returns an error wrapping the context's
	// error if the context expires before all handlers finish. Idempotent: subsequent
	// calls return nil immediately.
	Close(ctx context.Context) error
}

type busImpl struct {
	mu      sync.RWMutex
	allH    []Handler
	typedH  map[string][]Handler
	wg      sync.WaitGroup
	closed  bool
	timeout time.Duration
}

var _ Bus = (*busImpl)(nil)

// NewBus creates a new Bus and registers an fx.OnStop hook for graceful shutdown.
func NewBus(lifecycle fx.Lifecycle) Bus {
	bus := &busImpl{
		typedH:  make(map[string][]Handler),
		timeout: defaultHandlerTimeout,
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(stopCtx context.Context) error {
			return bus.Close(stopCtx)
		},
	})

	return bus
}

func (b *busImpl) Emit(parent context.Context, event Event) {
	if event == nil {
		panic("events: Emit called with nil event")
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		slog.DebugContext(parent, "emit after close, dropping event",
			"eventType", event.EventType(),
		)

		return
	}

	for _, h := range b.allH {
		b.dispatch(parent, h, event)
	}

	if typed, ok := b.typedH[event.EventType()]; ok {
		for _, h := range typed {
			b.dispatch(parent, h, event)
		}
	}
}

func (b *busImpl) dispatch(parent context.Context, handler Handler, event Event) {
	b.wg.Add(1)

	detached, cancel := detachContext(parent, b.timeout)

	safego.Go(detached, func() {
		defer b.wg.Done()
		defer cancel()

		handler(detached, event)
	})
}

func (b *busImpl) OnAll(handler Handler) {
	if handler == nil {
		panic("events: OnAll called with nil handler")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.allH = append(b.allH, handler)
}

func (b *busImpl) OnType(eventType string, handler Handler) {
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

func (b *busImpl) Close(closeCtx context.Context) error {
	b.mu.Lock()

	if b.closed {
		b.mu.Unlock()

		return nil
	}

	b.closed = true
	b.mu.Unlock()

	done := make(chan struct{})

	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-closeCtx.Done():
		return fmt.Errorf("event bus close: %w", closeCtx.Err())
	}
}

// detachContext creates a context detached from the parent's cancellation chain but
// preserving domain metadata. If the parent implements ctx.Detachable, its Detach
// method is used to preserve domain-specific scope (GameID, LobbyID, etc.). All other
// context types get a plain context.WithTimeout(Background, timeout).
func detachContext(
	parent context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	if d, ok := parent.(ctx.Detachable); ok {
		return d.Detach(timeout)
	}

	return context.WithTimeout(context.Background(), timeout)
}
