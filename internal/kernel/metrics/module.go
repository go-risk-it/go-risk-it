package metrics

import (
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() (*Metrics, error) {
		meter := otel.GetMeterProvider().Meter("go-risk-it")

		return NewMetrics(meter)
	}),
)
