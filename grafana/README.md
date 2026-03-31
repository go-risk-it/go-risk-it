# Observability Architecture

This document describes the go-risk-it observability stack: how telemetry is
produced, collected, and consumed in dashboards. It is the single reference for
understanding the spanmetrics-first architecture.

## Table of Contents

- [Observability Setup](#observability-setup)
- [Three-Signal Model](#three-signal-model)
- [Spanmetrics-First Principle](#spanmetrics-first-principle)
- [Collector Pipeline](#collector-pipeline)
- [Spanmetrics Dimensions](#spanmetrics-dimensions)
- [Span Catalog](#span-catalog)
- [RED Metric Helpers](#red-metric-helpers)
- [Manual Metrics (State Gauges Only)](#manual-metrics-state-gauges-only)
- [How to Add a New Metric](#how-to-add-a-new-metric)
- [Dashboard Toolchain](#dashboard-toolchain)

---

## Observability Setup

OpenTelemetry is initialized in `internal/kernel/otelsetup/`:

1. **`SetupOTelSDK`** configures three providers and registers them globally:
   - **TracerProvider** — OTLP/HTTP batch exporter for traces.
   - **MeterProvider** — OTLP/HTTP periodic reader (10s interval) with a
     `runtime.NewProducer()` for precomputed Go runtime histograms.
   - **LoggerProvider** — OTLP/HTTP batch processor for Loki + stdout simple
     processor for local dev visibility.
2. **Propagation** — W3C TraceContext + Baggage.
3. **Runtime instrumentation** — `runtime.Start()` emits goroutine/GC/heap
   gauges; `host.Start()` emits host-level metrics (CPU, memory, network).
4. All providers are wired into the `fx.Lifecycle` for graceful shutdown.

### The `observe` Package

`internal/kernel/observe/` is the **single API** for all business-logic
observability. It exposes eight functions:

| Function         | What it does                                                  |
|------------------|---------------------------------------------------------------|
| `Span`           | Starts a child span, returns `(ctx, done)` closure. Auto-rebases typed contexts via `Rebaseable` |
| `TypedSpan`      | Generic `Span` that preserves the caller's context type (e.g., `GameContext` in → `GameContext` out) |
| `TypedSpanFunc`  | Wraps a `(T, error)` function on a typed context with automatic span lifecycle — eliminates `done(nil)` bugs and type assertions |
| `TypedSpanErr`   | Like `TypedSpanFunc` but for functions returning only `error`  |
| `SpanEvent`      | Adds a named event to the current span                        |
| `Info`           | Dual-signal: slog INFO log + span event                       |
| `Warn`           | Dual-signal: slog WARN log + span event                       |
| `Error`          | Dual-signal: slog ERROR log + span event + RecordError + status. **Partial failures only** — if returning the error, use `done(err)` or `TypedSpanFunc`/`TypedSpanErr` instead |

The package is stateless: it uses `otel.GetTracerProvider()` and
`slog.Default()`, requiring no injected dependencies. Context attributes
(user_id, game_id, lobby_id) are automatically extracted via the `ctx.LogEnricher`
interface.

**Auto-rebase:** When `Span` or `TypedSpan` receives a typed context
(`GameContext`, `LobbyContext`, `UserContext`), the domain metadata is
automatically preserved across the OTel span boundary via the
`kernel/ctx.Rebaseable` interface. This means `observe.TypedSpan(gameCtx, "name")`
returns a `GameContext` with the new child span threaded through — no manual
context rebuilding needed.

**Choosing the right span function:**

| Pattern | When to use |
|---------|-------------|
| `TypedSpan(ctx, name)` | Typed contexts (GameContext, LobbyContext) — type inferred from argument |
| `TypedSpanFunc(ctx, name, fn)` | Typed-context functions returning `(T, error)` — zero-discipline error capture |
| `TypedSpanErr(ctx, name, fn)` | Typed-context functions returning `error` only |
| `Span(ctx, name)` + `done(nil)` | Void functions (no error return) — only correct use of `done(nil)` |

**Import rule:** Logic and domain packages must use `kernel/observe` for all
observability. Direct imports of `log/slog` or `go.opentelemetry.io/otel/trace`
in logic packages are banned by architecture tests.

---

## Three-Signal Model

Every observation in go-risk-it falls into one of three categories:

### 1. Spans — Operations with Duration

```go
// Typed context — preserves GameContext type:
ctx, done := observe.TypedSpan(ctx, "game.orchestrate_move", attrs...)
defer func() { done(err) }()

// Or zero-discipline wrapper (preferred for error-returning functions):
return observe.TypedSpanErr(ctx, "game.move.validate", func(ctx gamectx.GameContext) error {
    // ... validation logic ...
})
```

Spans represent timed operations. The OTel Collector's **spanmetrics connector**
automatically derives RED metrics (Rate, Errors, Duration) from every span.
Spans are the primary signal — they feed both traces (Tempo) and metrics
(Mimir) simultaneously.

### 2. Span Events — In-Span Decision Points

```go
observe.SpanEvent(ctx, "player_eliminated", attrs...)
```

Span events are lightweight annotations within a span. They appear in traces
but do not generate metrics. Use them for decision points, state transitions,
or checkpoints that help debug a specific trace.

### 3. Info / Warn / Error — Dual-Signal Emission

```go
observe.Info(ctx, "move validated", attribute.String("phase", "ATTACK"))
observe.Error(ctx, err, "transaction failed")
```

These emit **two signals** from a single call:
- An **slog log record** (shipped to Loki via the OTLP log exporter).
- A **span event** on the current trace (visible in Tempo).

`Error` additionally calls `span.RecordError()` and `span.SetStatus(Error)`,
marking the span as failed in both the trace and spanmetrics.

---

## Spanmetrics-First Principle

> **Spans are the single source of truth for RED metrics. No manual RED
> metrics exist in the codebase.**

The architecture follows one rule: if you need Rate, Error rate, or Duration
for an operation, you instrument it with a span. The OTel Collector's
spanmetrics connector automatically produces:

- `traces_spanmetrics_calls_total` — counter by span name, status code, and
  configured dimensions.
- `traces_spanmetrics_duration_seconds_bucket` — histogram with explicit
  buckets from 1ms to 10s.

This means:
- **No `metric.Float64Histogram` for latency.** Span duration IS the histogram.
- **No `metric.Int64Counter` for request rates.** Span call count IS the rate.
- **No `metric.Int64Counter` for error rates.** Span error status IS the error
  rate.

### What IS manually instrumented

Only **state gauges** that have no span equivalent are manually registered in
`internal/kernel/metrics/` and `internal/game/logic/metrics/`:

| Metric                       | Type              | Package            |
|------------------------------|-------------------|--------------------|
| `ws.connections.active`      | Int64UpDownCounter| kernel/metrics     |
| `db.transaction.retries.total` | Int64Counter    | kernel/metrics     |
| `game.active`                | Int64UpDownCounter| game/logic/metrics |
| `game.duration`              | Float64Histogram  | game/logic/metrics |

These represent ongoing state or lifecycle measurements that cannot be derived
from request-scoped spans.

---

## Collector Pipeline

Telemetry flows through a three-layer pipeline from application to dashboards:

```
 Layer 1: Application (Go backend)
 ──────────────────────────────────
  observe.Span()  ─┐
  observe.Info()   ├─► OTLP/gRPC ─────────────────────────┐
  observe.Error()  ┘                                       │
  Manual gauges ──────► OTLP/HTTP ──────────────────────┐  │
                                                        │  │
 Layer 2: OTel Collector                                ▼  ▼
 ──────────────────────────────────   ┌─────────────────────────┐
  Receivers:                          │    OTel Collector        │
    otlp (gRPC :4317, HTTP :4318)     │                         │
    prometheus/collector (:8888)       │  traces pipeline:       │
    prometheus/postgres  (:9187)       │    otlp → batch →       │
                                      │      ├─► Tempo (traces) │
  Connectors:                         │      └─► spanmetrics ───┤
    spanmetrics ──────────────────────│─┐                       │
      histogram: 12 explicit buckets  │ │  metrics pipeline:    │
      dimensions: 10 labels           │ │    otlp + prometheus  │
      exemplars: enabled              │ │    + spanmetrics      │
      flush: 5s                       │ │    → batch → Mimir    │
      expiration: 5m                  │ │                       │
                                      │ │  logs pipeline:       │
                                      │ │    otlp → batch → Loki│
                                      └─┴───────────────────────┘
                                                   │
 Layer 3: Storage & Visualization                  ▼
 ──────────────────────────────────   ┌─────────────────────────┐
                                      │  Mimir    (metrics)     │
                                      │  Tempo    (traces)      │
                                      │  Loki     (logs)        │
                                      │          │              │
                                      │     Grafana             │
                                      │   3 dashboards          │
                                      └─────────────────────────┘
```

The critical path is: **spans → spanmetrics connector → Mimir → Grafana**.
Every dashboard panel showing rate, latency, or error rate queries
`traces_spanmetrics_*` metrics, not application-emitted counters.

### Exemplar Flow

The spanmetrics connector has `exemplars.enabled: true`. This means duration
histogram buckets carry trace IDs as exemplars. In Grafana, clicking an exemplar
dot on a latency panel jumps directly to the corresponding trace in Tempo.

---

## Spanmetrics Dimensions

The spanmetrics connector extracts 10 span attributes as metric labels
(configured in `grafana/otelcol-extra.yaml`):

| Dimension          | Description                                        | Example values                     |
|--------------------|----------------------------------------------------|------------------------------------|
| `http_route`       | HTTP route pattern from the request                | `POST /api/v1/games/{id}/move`     |
| `http_method`      | HTTP method                                        | `GET`, `POST`, `PUT`, `DELETE`     |
| `http_status_code` | HTTP response status code                          | `200`, `400`, `500`                |
| `error_category`   | Classified error type                              | `validation`, `not_found`          |
| `event_type`       | Event bus event type                               | `move_executed`, `game_completed`  |
| `handler`          | Event handler name                                 | `broadcaster`, `move_log`          |
| `isolation`        | Transaction isolation level                        | `serializable`, `read_committed`   |
| `phase`            | Game phase                                         | `DEPLOY`, `ATTACK`, `REINFORCE`    |
| `db_outcome`       | Database transaction outcome                       | `committed`, `retried`, `failed`   |
| `ws_fanout`        | WebSocket fanout target type                       | `public`, `private`                |

These appear as labels on both `traces_spanmetrics_calls_total` and
`traces_spanmetrics_duration_seconds_bucket`. Dashboard queries filter and
group by these dimensions.

### Histogram Buckets

```
1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
```

### Cardinality Controls

- `metrics_expiration: 5m` — stale series are cleaned up after 5 minutes of
  inactivity.
- `aggregation_cardinality_limit: 10000` — hard cap on unique label
  combinations.

---

## Span Catalog

Span name patterns used across the codebase, defined as constants in
`grafana/lib/targets.libsonnet` under `spans::`:

| Key             | Pattern                         | What it covers                          |
|-----------------|---------------------------------|-----------------------------------------|
| `http`          | `(GET\|POST\|PUT\|DELETE) .*`   | All HTTP request spans                  |
| `db`            | `db\.transaction`               | Database transaction spans              |
| `gameLogic`     | `game\.orchestrate_move`        | Move orchestration (core game loop)     |
| `wsBroadcast`   | `ws\.broadcast`                 | WebSocket state broadcast               |
| `eventHandler`  | `consumer\..*`                  | All event bus consumer handlers         |
| `busDispatch`   | `bus:.*`                        | Event bus dispatch spans                |
| `snapshot`      | `snapshot\..*`                  | Snapshot service operations             |
| `gameCreate`    | `game\.create`                  | Game creation                           |
| `lobbyLobbies`  | `lobby\.get_user_lobbies`       | Lobby listing                           |

These patterns are used as the `spanNameFilter` argument to the RED metric
helpers below.

---

## RED Metric Helpers

`grafana/lib/targets.libsonnet` provides six helpers for building PromQL
queries against spanmetrics. All helpers use `service="risk-it"` by default.

### `spanDuration(spanNameFilter, quantiles, exemplars=false, extraLabels='')`

Generates `histogram_quantile` targets from `traces_spanmetrics_duration_seconds_bucket`.

```jsonnet
// p50/p95/p99 latency for HTTP requests
targets.spanDuration(targets.spans.http, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']])
```

### `spanRate(spanNameFilter, legend, extraLabels='')`

Generates a `sum(rate(...))` target from `traces_spanmetrics_calls_total`,
excluding error spans (`status_code!="STATUS_CODE_ERROR"`).

```jsonnet
// Success rate for game logic
targets.spanRate(targets.spans.gameLogic, 'Moves/s')
```

### `spanErrorRate(spanNameFilter, legend, extraLabels='')`

Same as `spanRate` but filters FOR errors (`status_code="STATUS_CODE_ERROR"`).

```jsonnet
// Error rate for HTTP requests
targets.spanErrorRate(targets.spans.http, 'Errors/s')
```

### `spanRateBy(spanNameFilter, groupBy, legend, extraLabels='')`

Rate grouped by a dimension label (e.g., `span_name`, `phase`, `event_type`).

```jsonnet
// Event throughput by type
targets.spanRateBy(targets.spans.eventHandler, 'event_type', '{{event_type}}')
```

### `spanDurationBy(spanNameFilter, quantile, groupBy, legend, extraLabels='')`

Duration percentile grouped by a dimension label.

```jsonnet
// p95 latency by HTTP route
targets.spanDurationBy(targets.spans.http, '0.95', 'http_route', '{{http_route}}')
```

### `spanLifecycleTargets`

Pre-built array of p95 latency targets for the five server boundaries (HTTP,
DB Transaction, Game Logic, WS Broadcast, Event Handler), all with exemplar
support enabled. Used by the system-health and perf-test dashboards for
lifecycle timing panels.

---

## Manual Metrics (State Gauges Only)

Two metric families coexist, distinguished by their service name label:

| Family       | Label         | Source                     | Helpers                      |
|--------------|---------------|----------------------------|------------------------------|
| Spanmetrics  | `service`     | OTel Collector connector   | `spanDuration`, `spanRate`, etc. |
| Manual       | `service_name`| App code / perf-test client| `histogramQuantileTargets`, `phaseLatencyTargets` |

Manual metric helpers in `targets.libsonnet`:

- **`histogramQuantileTargets(metric, quantiles, serviceName)`** — histogram
  percentiles for manually instrumented histograms.
- **`histogramQuantileTargetsWithExemplars(...)`** — same with exemplar support.
- **`phaseLatencyTargets(phase)`** — per-phase p50/p95/p99 for
  `game_phase_duration_seconds_bucket`.

---

## How to Add a New Metric

### Scenario A: You need rate, latency, or error rate for an operation

**Do NOT create a manual metric.** Add a span instead.

1. **Add a span** in your service method:
   ```go
   ctx, done := observe.Span(ctx, "mypackage.operation_name",
       attribute.String("my_dimension", value),
   )
   defer done(nil)
   ```

2. **Add a dimension** (if the attribute should be a metric label) by appending
   to the `dimensions` list in `grafana/otelcol-extra.yaml`:
   ```yaml
   dimensions:
     - name: my_dimension
   ```
   Skip this step if you are using an existing dimension (e.g., `phase`,
   `http_route`).

3. **Add a span catalog entry** in `grafana/lib/targets.libsonnet`:
   ```jsonnet
   spans:: {
     // ... existing entries ...
     myOperation: 'mypackage\\.operation_name',
   },
   ```

4. **Build dashboard panels** using the RED helpers:
   ```jsonnet
   // Latency
   panels.spanPercentileBandsPanel('My Operation Latency', targets.spans.myOperation, 's')
   // Throughput
   panels.timeseriesPanel('My Operation Rate', [targets.spanRate(targets.spans.myOperation, 'ops/s')], 'ops')
   // Error rate
   panels.timeseriesPanel('My Operation Errors', [targets.spanErrorRate(targets.spans.myOperation, 'errors/s')], 'ops')
   ```

5. **Regenerate dashboards**:
   ```bash
   make dashboards
   ```

### Scenario B: You need a state gauge (no span equivalent)

Only for ongoing state measurements (active connections, active games).

1. **Register the instrument** in `internal/kernel/metrics/` or
   `internal/game/logic/metrics/`:
   ```go
   MyGauge, err = meter.Int64UpDownCounter("my.state.gauge",
       metric.WithDescription("Number of active X"),
   )
   ```

2. **Increment/decrement** at state transitions in your service code.

3. **Query in dashboards** using `service_name="risk-it"` (manual metrics use
   `service_name`, not `service`).

---

## Dashboard Toolchain

Dashboards are defined as Jsonnet files using custom shared libraries in
`grafana/lib/`.

### File Layout

```
grafana/
  otelcol-extra.yaml           # OTel Collector config (spanmetrics connector)
  dashboards/
    game-engine.jsonnet         # Game engine operational dashboard
    perf-test.jsonnet           # Performance test analysis dashboard
    system-health.jsonnet       # System health overview dashboard
    *.json                      # Generated — do not edit by hand
  lib/
    targets.libsonnet           # PromQL target builders + span catalog
    panels.libsonnet            # Panel builder functions (timeseries, stat, gauge, heatmap, etc.)
    colors.libsonnet            # Semantic boundary colors (7 architectural + shades)
    thresholds.libsonnet        # SLO threshold definitions
    links.libsonnet             # Cross-dashboard data links
    modifiers.libsonnet         # Panel modifier helpers (percentile colors)
    layout.libsonnet            # Grid layout helpers
    dashboard.libsonnet         # Dashboard-level settings
  provisioning/                 # Grafana provisioning config
```

### Build Commands

```bash
# Generate JSON from all .jsonnet sources
make dashboards

# Verify committed JSON is up to date (used in CI)
make dashboards-check
```

Both commands use `jsonnet -J grafana/lib` for import resolution and
`python3 -m json.tool` for deterministic JSON formatting.

### Prerequisites

```bash
brew install go-jsonnet
```

### Conventions

- **OODA row layout** — dashboards follow an Observe/Orient/Decide/Act
  narrative structure with row IDs 100/200/300/400.
- **Panel widths** — `w=8` (3-across), `w=12` (2-across), `w=24` (full-width
  hero).
- **Panel descriptions** — follow the format "Normal: X. Watch for: Y. Check
  next: Z."
- **Colors** — single-series panels use semantic boundary colors from
  `colors.libsonnet`; multi-series panels use `palette-classic`.
- **Exemplar panels** — use `spanDuration(..., exemplars=true)` only on
  timeseries panels (stat panels cannot render exemplar dots).
