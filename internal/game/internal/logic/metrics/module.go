package metrics

import (
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() (*GameMetrics, error) {
		meter := otel.GetMeterProvider().Meter("go-risk-it")

		return NewGameMetrics(meter)
	}),
	fx.Provide(NewGameTiming),
	fx.Invoke(func(gameMetrics *GameMetrics, sub eventbus.Subscriber, pub eventbus.Publisher) {
		RegisterGameSummaryRecorder(gameMetrics, sub)
		// Package placement note: TurnEnded derivation is game logic,
		// not metrics. Consider a dedicated game/consumers or game/derived package
		// for bus consumers that emit derived events (TurnEnded, headlines).
		gameevt.RegisterTurnEndedEmitter(sub, pub)
	}),
)
