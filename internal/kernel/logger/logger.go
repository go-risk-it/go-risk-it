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

	Bus    bus.Bus
	Logger *slog.Logger `optional:"true"`
}

// Register subscribes to all events and logs each as a single structured
// slog line at LevelInfo. Top-level attributes (event_type, event_timestamp)
// are always present. Scope IDs (game_id, lobby_id) are added via type-switch
// on the concrete event interface. The ToRecord() payload appears as a nested
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

		// Scope ID: type-switch on concrete event interface for domain-specific identifiers.
		switch e := event.(type) {
		case interface{ GameID() int64 }:
			attrs = append(attrs, slog.Int64("gameId", e.GameID()))
		case interface{ LobbyID() int64 }:
			attrs = append(attrs, slog.Int64("lobbyId", e.LobbyID()))
		}

		attrs = append(attrs, slog.Group("payload", payloadAttrs...))

		log.LogAttrs(
			ctx,
			slog.LevelInfo,
			"game_event",
			attrs...,
		)
	})
}
