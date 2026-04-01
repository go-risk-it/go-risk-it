// Package loadtest implements a performance testing harness for the go-risk-it
// game server. It lives in the same Go module as the server but shares only
// kernel packages (observe, ctx) — enforced by arch_test rules L1 and L2.
//
// # Architecture
//
// The loadtest binary (cmd/loadtest/) orchestrates concurrent game simulations
// through three layers:
//
//   - Orchestrator: staircase/adaptive step execution, game pool management,
//     health monitoring, SLO evaluation.
//   - Runner: event-driven game lifecycle (synchronous Bus with typed events,
//     8 handler types: Tracing, Metrics, Health, Strategy, Protocol, Error,
//     StateWatcher, Chaos).
//   - Client: REST + WebSocket communication with the game server, including
//     W3C trace propagation for cross-service span linking.
//
// Metrics are split into two components with zero shared mutable state:
//
//   - StepAccumulator: per-step HDR histograms, atomic counters, warm-up
//     gating, throughput buckets. Created fresh for each staircase step.
//   - LiveMetrics: cross-step OTel state gauges (games.active, health
//     classification). Wraps OTelExporter for the full run lifetime.
//
// # Observability
//
// The loadtest uses the server's observe package (observe.Info/Warn/Error,
// observe.RawSpan) for all logging and tracing. A full otelslog bridge
// provides trace-log correlation. Two span types are produced:
//
//   - perftest.game.run: root span covering an entire game lifecycle.
//   - perftest.move.execute: child span per move (strategy + REST + WS wait).
//
// A separate spanmetrics connector in the OTel Collector derives RED metrics
// from these spans, avoiding manual OTel instrument duplication.
//
// # Module Evolution (5-year path)
//
// The single-module design (one go.mod, multiple cmd/ binaries) scales to
// 5+ binaries without structural changes:
//
//   - Year 0 (now): cmd/risk-it (server) + cmd/loadtest (perf harness).
//     Shared surface: kernel/observe, kernel/ctx.
//   - Year 1-2: cmd/risk-cli (standalone admin CLI) with own otelcfg (no-fx).
//     May trigger shared/gametypes/ extraction if CLI needs board state types.
//   - Year 3-5: cmd/risk-bot (AI opponent), cmd/risk-replay (game replay).
//     May trigger shared/protocol/ and shared/gameclient/ (promoted from
//     loadtest/client if a third consumer appears).
//
// Key invariants that enable this evolution:
//
//   - Per-binary compilation isolation: go build ./cmd/X/ compiles only
//     packages reachable from X's main.go. No binary pays for another's deps.
//   - Arch test enforcement: rules L1/L2 prevent loadtest↔game/lobby imports.
//     New binaries get equivalent rules.
//   - Shared packages live in kernel/ (infra) or shared/ (domain). Never
//     import upward into cmd/ or feature packages.
//   - Each binary owns its own OTel provider lifecycle. No shared global state
//     beyond otel.SetTracerProvider (which is per-process anyway).
package loadtest
