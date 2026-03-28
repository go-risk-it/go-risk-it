package events

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

const defaultHandlerTimeout = 10 * time.Second

// Bus dispatches events to registered handlers. Each event gets one goroutine;
// handlers run sequentially within that goroutine (OnAll before OnType) with
// per-handler panic recovery. Close gracefully drains in-flight handlers.
type Bus interface {
	// Emit dispatches event to all matching handlers. Handlers run sequentially in a
	// single goroutine per event (OnAll before OnType). Panics if event is nil. No-op
	// after Close.
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
	metrics *metrics.Metrics
}

var _ Bus = (*busImpl)(nil)

// NewBus creates a new Bus and registers an fx.OnStop hook for graceful shutdown.
func NewBus(lifecycle fx.Lifecycle, m *metrics.Metrics) Bus {
	bus := &busImpl{
		typedH:  make(map[string][]Handler),
		timeout: defaultHandlerTimeout,
		metrics: m,
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

	eventType := event.EventType()

	if b.metrics != nil {
		b.metrics.EventBusEventsTotal.Add(parent, 1,
			metric.WithAttributes(attribute.String("event_type", eventType)))
	}

	handlers := b.collectHandlers(eventType)
	if len(handlers) == 0 {
		return
	}

	b.dispatchEvent(handlers, parent, event)
}

// collectHandlers returns a copy of all matching handlers (OnAll first, then OnType)
// under a read lock. Returns nil if the bus is closed.
func (b *busImpl) collectHandlers(eventType string) []Handler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil
	}

	handlers := make([]Handler, 0, len(b.allH)+len(b.typedH[eventType]))
	handlers = append(handlers, b.allH...)
	handlers = append(handlers, b.typedH[eventType]...)

	return handlers
}

// dispatchEvent launches a single goroutine for the event that runs all handlers
// sequentially. Each handler is wrapped in runHandler for per-handler panic recovery.
func (b *busImpl) dispatchEvent(
	handlers []Handler,
	parent context.Context,
	event Event,
) {
	b.wg.Add(1)

	go func() {
		detached, cancel := detachContext(parent, event, b.timeout)
		defer b.wg.Done()
		defer cancel()

		start := time.Now()

		for _, handler := range handlers {
			runHandler(detached, handler, event)
		}

		if b.metrics != nil {
			duration := time.Since(start).Seconds()
			b.metrics.EventBusDispatchDuration.Record(detached, duration,
				metric.WithAttributes(attribute.String("event_type", event.EventType())))
		}
	}()
}

// runHandler invokes a single handler with per-handler panic recovery. A panicking
// handler is logged and does not prevent subsequent handlers from executing.
func runHandler(handlerCtx context.Context, handler Handler, event Event) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.ErrorContext(
				handlerCtx,
				"panic recovered in event handler",
				"panic", recovered,
				"stack", string(debug.Stack()),
				"eventType", event.EventType(),
			)
		}
	}()

	handler(handlerCtx, event)
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
// preserving domain metadata. It starts a linked span rooted in its own trace and
// applies a timeout. If the parent implements ctx.Detachable, DetachOnto enriches the
// timeout context with domain-specific scope (GameID, LobbyID, etc.). The returned
// cancel function ends the linked span and cancels the timeout context.
func detachContext(
	parent context.Context,
	event Event,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	spanName := "bus:" + event.EventType()
	spanCtx, span := startLinkedSpan(parent, spanName)

	timeoutCtx, timeoutCancel := context.WithTimeout(spanCtx, timeout)

	result := timeoutCtx
	if d, ok := parent.(ctx.Detachable); ok {
		result = d.DetachOnto(timeoutCtx)
	}

	cancel := func() {
		span.End()
		timeoutCancel()
	}

	return result, cancel
}

const busTracerName = "go-risk-it-eventbus"

// startLinkedSpan creates a new root span linked to the trigger span from the parent
// context. This provides trace correlation (via the link) while ensuring handler spans
// live in their own trace (via WithNewRoot on context.Background). Degrades gracefully
// to a noop span when the global TracerProvider is a noop.
func startLinkedSpan(
	parent context.Context,
	spanName string,
) (context.Context, trace.Span) {
	triggerSpan := trace.SpanFromContext(parent)
	opts := []trace.SpanStartOption{trace.WithNewRoot()}

	if triggerSpan.SpanContext().IsValid() {
		opts = append(opts, trace.WithLinks(trace.Link{
			SpanContext: triggerSpan.SpanContext(),
		}))
	}

	//nolint:spancheck // span is returned to the caller who manages its lifecycle
	return otel.GetTracerProvider().Tracer(busTracerName).Start(
		context.Background(), spanName, opts...,
	)
}
