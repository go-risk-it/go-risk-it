// Package otelsetup initializes the OpenTelemetry SDK for tracing and metrics.
//
// [SetupOTelSDK] configures trace and meter providers with OTLP/HTTP
// exporters, sets up context propagation (W3C TraceContext + Baggage),
// and starts Go runtime and host instrumentation. The SDK is wired into
// the fx lifecycle for graceful shutdown of all providers.
//
// # Layer
//
// Kernel — observability infrastructure bootstrap.
package otelsetup
