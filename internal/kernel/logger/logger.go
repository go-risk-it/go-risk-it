package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"go.uber.org/fx"
)

// Params holds the dependencies for registering the event logger.
type Params struct {
	fx.In

	Bus    bus.Subscriber
	Logger *slog.Logger `optional:"true"`
}

// Register subscribes to all events and logs each as a single structured
// slog line at LevelInfo. Top-level attributes (event_type, event_timestamp)
// are always present. Scope IDs (game_id, lobby_id) are added via the
// ScopedEvent interface. The ToRecord() payload appears as a nested
// "payload" group for full event data.
func Register(params Params) {
	log := params.Logger
	if log == nil {
		log = slog.Default()
	}

	params.Bus.OnAll(func(ctx context.Context, event bus.Event) {
		record := event.ToRecord()

		payloadAttrs := make([]any, 0, len(record)*2)
		for k, v := range record {
			payloadAttrs = append(payloadAttrs, slog.Any(k, v))
		}

		attrs := []slog.Attr{
			slog.String("eventType", event.EventType()),
			slog.String("eventTimestamp", event.EventTimestamp().Format(time.RFC3339)),
		}

		// Scope ID: use structured ScopeAttrs if the event implements ScopedEvent.
		if scoped, ok := event.(bus.ScopedEvent); ok {
			attrs = append(attrs, scoped.ScopeAttrs()...)
		}

		attrs = append(attrs, slog.Group("payload", payloadAttrs...))

		// Use event.EventTimestamp() as the log record timestamp so OTLP/Loki
		// sorts entries by domain event time, not goroutine-scheduling time.
		// Critical for TurnEnded: its 1ms offset must reach the actual log
		// timestamp to guarantee it sorts AFTER the preceding REINFORCE move.
		logRecord := slog.NewRecord(event.EventTimestamp(), slog.LevelInfo, "game_event", 0)
		logRecord.AddAttrs(attrs...)
		_ = log.Handler().Handle(ctx, logRecord)
	})
}
