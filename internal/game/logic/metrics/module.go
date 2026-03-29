package metrics

import (
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() (*GameMetrics, error) {
		meter := otel.GetMeterProvider().Meter("go-risk-it")

		return NewGameMetrics(meter)
	}),
)
