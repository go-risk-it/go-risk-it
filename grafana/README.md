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
- [Dashboard Library Architecture](#dashboard-library-architecture)
- [OODA Layout Pattern](#ooda-layout-pattern)
- [Color System](#color-system)
- [Panel Conventions](#panel-conventions)
- [Cross-Dashboard Navigation](#cross-dashboard-navigation)
- [How to Add a New Panel](#how-to-add-a-new-panel)
- [How to Add a New Dashboard](#how-to-add-a-new-dashboard)

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
| `Span`           | Generic span that preserves the caller's context type (e.g., `GameContext` in → `GameContext` out). Primary API for business logic |
| `SpanFunc`       | Wraps a `(T, error)` function on a typed context with automatic span lifecycle — eliminates `done(nil)` bugs and type assertions |
| `SpanErr`        | Like `SpanFunc` but for functions returning only `error`      |
| `RawSpan`        | Starts a child span on plain `context.Context`, returns `(ctx, done)` closure. Auto-rebases typed contexts via `Rebaseable`. For infrastructure callers only |
| `SpanEvent`      | Adds a named event to the current span                        |
| `Info`           | Dual-signal: slog INFO log + span event                       |
| `Warn`           | Dual-signal: slog WARN log + span event                       |
| `Error`          | Dual-signal: slog ERROR log + span event + RecordError + status. **Partial failures only** — if returning the error, use `done(err)` or `SpanFunc`/`SpanErr` instead |

The package is stateless: it uses `otel.GetTracerProvider()` and
`slog.Default()`, requiring no injected dependencies. Context attributes
(user_id, game_id, lobby_id) are automatically extracted via the `ctx.LogEnricher`
interface.

**Auto-rebase:** When `Span` or `RawSpan` receives a typed context
(`GameContext`, `LobbyContext`, `UserContext`), the domain metadata is
automatically preserved across the OTel span boundary via the
`kernel/ctx.Rebaseable` interface. This means `observe.Span(gameCtx, "name")`
returns a `GameContext` with the new child span threaded through — no manual
context rebuilding needed.

**Choosing the right span function:**

| Pattern | When to use |
|---------|-------------|
| `Span(ctx, name)` | Typed contexts (GameContext, LobbyContext) — type inferred from argument |
| `SpanFunc(ctx, name, fn)` | Typed-context functions returning `(T, error)` — zero-discipline error capture |
| `SpanErr(ctx, name, fn)` | Typed-context functions returning `error` only |
| `RawSpan(ctx, name)` + `done(nil)` | Void functions on plain context.Context (no error return) — only correct use of `done(nil)` |

**Import rule:** Logic and domain packages must use `kernel/observe` for all
observability. Direct imports of `log/slog` or `go.opentelemetry.io/otel/trace`
in logic packages are banned by architecture tests.

---

## Three-Signal Model

Every observation in go-risk-it falls into one of three categories:

### 1. Spans — Operations with Duration

```go
// Typed context — preserves GameContext type:
ctx, done := observe.Span(ctx, "game.orchestrate_move", attrs...)
defer func() { done(err) }()

// Or zero-discipline wrapper (preferred for error-returning functions):
return observe.SpanErr(ctx, "game.move.validate", func(ctx gamectx.GameContext) error {
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

- `traces_span_metrics_calls_total` — counter by span name, status code, and
  configured dimensions.
- `traces_span_metrics_duration_milliseconds_bucket` — histogram with explicit
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
                                      │   4 dashboards          │
                                      └─────────────────────────┘
```

The critical path is: **spans → spanmetrics connector → Mimir → Grafana**.
Every dashboard panel showing rate, latency, or error rate queries
`traces_span_metrics_*` metrics, not application-emitted counters.

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

These appear as labels on both `traces_span_metrics_calls_total` and
`traces_span_metrics_duration_milliseconds_bucket`. Dashboard queries filter and
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

Generates `histogram_quantile` targets from `traces_span_metrics_duration_milliseconds_bucket`.

```jsonnet
// p50/p95/p99 latency for HTTP requests
targets.spanDuration(targets.spans.http, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']])
```

### `spanRate(spanNameFilter, legend, extraLabels='')`

Generates a `sum(rate(...))` target from `traces_span_metrics_calls_total`,
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
`grafana/lib/`. The Jsonnet source is compiled to JSON and provisioned into
Grafana via the `grafana/provisioning/` directory.

### File Layout

```
grafana/
  otelcol-extra.yaml           # OTel Collector config (spanmetrics connector)
  dashboards/
    system-health.jsonnet       # Infrastructure & RED overview
    game-engine.jsonnet         # Game lifecycle, phase timing, event bus
    game-theater.jsonnet        # Single-game deep dive (tapestry, chronicle, forensics)
    perf-test.jsonnet           # Load test client metrics + SLO compliance
    *.json                      # Generated — do not edit by hand
  lib/
    targets.libsonnet           # PromQL target builders + span catalog
    panels.libsonnet            # Panel builder functions (timeseries, stat, gauge, heatmap, etc.)
    colors.libsonnet            # Semantic boundary colors (7 architectural + shades)
    thresholds.libsonnet        # SLO threshold definitions
    links.libsonnet             # Cross-dashboard data links + UID registry
    modifiers.libsonnet         # Composable panel modifiers (SLO, stacked, colors, links)
    layout.libsonnet            # Auto-layout engine (OODA structure)
    queries.libsonnet           # Composable LogQL query helpers
    dashboard.libsonnet         # Dashboard envelope factory + annotations
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

---

## Dashboard Library Architecture

The 9 `.libsonnet` files form a layered library. Lower layers have no
upward dependencies; higher layers compose from lower ones.

### Layer 0: Data & Constants (no imports)

| File | Role |
|------|------|
| `colors.libsonnet` | 7 architectural boundary colors, shade system, phase/signal/headline/eventType palettes, `fixedColor()` helper |
| `thresholds.libsonnet` | SLO threshold objects (`e2eP95`, `dbTxnP95`, `poolUtil`, etc.) |
| `links.libsonnet` | Dashboard UID registry, `toDashboard()` and `toDashboardWithVar()` constructors, variable contract documentation |
| `queries.libsonnet` | LogQL stream selectors, event type constants, `lokiEventStream()`, `lokiCountOverTime()` composable helpers |

### Layer 1: Target Builders (imports: colors)

| File | Role |
|------|------|
| `targets.libsonnet` | Datasource UIDs, service name constants, span catalog (`spans::`), PromQL target constructors (`target()`, `lokiTarget()`, `tempoTarget()`, `heatmapTarget()`), spanmetrics helpers (`spanDuration`, `spanRate`, `spanDurationBy`, etc.), manual metric helpers (`histogramQuantileTargets`, `phaseLatencyTargets`), perf-test helpers, lifecycle boundary targets + overrides |

### Layer 2: Panel Builders (imports: targets)

| File | Role |
|------|------|
| `panels.libsonnet` | Panel constructors: `timeseriesPanel()`, `statPanel()`, `gaugePanel()`, `heatmapPanel()`, `barGaugePanel()`, `logPanel()`, `tracesPanel()`, `stateTimelinePanel()`, `percentileBandsPanel()`, `spanPercentileBandsPanel()`, `rowPanel()`. Encodes visual defaults (line style, legend, tooltip) |
| `modifiers.libsonnet` | Composable merge objects: `withSloThreshold()`, `withStackedArea()`, `withSeriesColors()`, `withDashedSeries()`, `withLinks()`, `withPercentileColors()`. Applied with `panel + modifier` syntax |

### Layer 3: Layout Engine (no imports)

| File | Role |
|------|------|
| `layout.libsonnet` | `panel(p, w, h, description)` wrapper, `ooda()` main entry point. Auto-assigns panel IDs and grid positions. Handles collapsed depth rows |

### Layer 4: Dashboard Envelope (no imports)

| File | Role |
|------|------|
| `dashboard.libsonnet` | `new(uid, title, panels, ...)` factory, `perfTestAnnotation` constant |

### Import Graph

```
dashboard.libsonnet ─────────────────────────────────────────────> dashboards/*.jsonnet
layout.libsonnet ────────────────────────────────────────────────> dashboards/*.jsonnet
panels.libsonnet ──imports──> targets.libsonnet ──imports──> colors.libsonnet
modifiers.libsonnet ─imports─> colors.libsonnet
links.libsonnet (standalone)
queries.libsonnet (standalone)
thresholds.libsonnet (standalone)
```

---

## OODA Layout Pattern

Every dashboard follows the OODA loop narrative structure. The `layout.ooda()`
function accepts 8 parameters — 4 visible sections and 4 depth sections:

```jsonnet
layout.ooda(
  observe=[...],        // Am I OK? — stat tiles, summary gauges
  observeDepth={...},   // Collapsed sub-rows below Observe
  orient=[...],         // What's the shape? — hero panels, heatmaps, trends
  orientDepth={...},    // Collapsed sub-rows below Orient
  decide=[...],         // Where's the bottleneck? — comparisons, correlations
  decideDepth={...},    // Collapsed sub-rows below Decide
  act=[...],            // What's the evidence? — traces, logs, tables
  actDepth={...},       // Collapsed sub-rows below Act
)
```

### Section Conventions

| Section | Purpose | Typical panels |
|---------|---------|----------------|
| **Observe** | Glanceable health status | `statPanel` (w=4-8), background color mode |
| **Orient** | Shape and distribution | `timeseriesPanel` (w=12-24), `heatmapPanel`, hero panels |
| **Decide** | Bottleneck identification | Correlation panels, grouped-by latency, dual-axis |
| **Act** | Evidence and investigation | `logPanel`, `tracesPanel`, table panels |

### Depth Rows (Collapsed)

Depth sections are objects mapping row title to panel array:

```jsonnet
orientDepth={
  'Database': [layout.panel(p1, 12, 8), layout.panel(p2, 12, 8)],
  'Server & HTTP': [layout.panel(p3, 24, 8)],
},
```

Collapsed rows are sorted alphabetically by title. They render as clickable
row headers that expand to reveal nested panels.

### Row Titles

The engine generates 4 em-dash question titles:
- `Observe — Am I OK?`
- `Orient — What's the shape?`
- `Decide — Where's the bottleneck?`
- `Act — What's the evidence?`

### Panel Wrapping

`layout.panel(p, w, h, description)` wraps any panel object with grid
metadata. Panels flow left-to-right, wrapping at x=24 (Grafana grid width).
IDs are auto-assigned sequentially — never set `id` or `gridPos` manually.

---

## Color System

### 7 Architectural Boundary Colors

| Name | Hex | Semantic |
|------|-----|----------|
| `db` | `#3274D9` | Database — blue |
| `ws` | `#8F3BB8` | WebSocket — purple |
| `gameLogic` | `#56A64B` | Game Logic — green |
| `http` | `#FF9830` | HTTP — amber |
| `errors` | `#E02F44` | Errors — red |
| `client` | `#73BF69` | Client/perf-test — cyan-green |
| `eventBus` | `#00BCD4` | Event Bus — teal |

### Shade System (Percentile Panels)

Each boundary has light/medium/dark variants in `colors.shades`:

```jsonnet
shades: {
  db: { light: '#73A9F2', medium: '#3274D9', dark: '#1F4FA0' },
  // ... one entry per boundary
}
```

Applied via `modifiers.withPercentileColors('db')` which maps p50=light,
p95=medium, p99=dark.

### Palette Hierarchy

| Context | Color Strategy |
|---------|---------------|
| Single-series panel | `color=colors.fixedColor(colors.XX)` — boundary identification |
| Multi-series data-driven (by route, phase, label) | No color param — `palette-classic` |
| Multi-series percentile (p50/p95/p99) | `+ modifiers.withPercentileColors('XX')` |

### Domain Palettes

- **`phase`** — 5 game phase colors (deploy=blue, attack=red, cards=gold, conquer=green, reinforce=purple)
- **`signal`** — 5 operational severity colors (ok=green, warning=amber, error=red, info=blue, muted=gray)
- **`headline`** — 3 notable game moment colors (continent=gold, elimination=red, victory=green)
- **`eventTypes`** — 8 event type colors for stacked area charts
- **`lockModes`** — 8 Postgres lock mode colors for contention visualization

---

## Panel Conventions

### Widths

| Width | Layout | Use case |
|-------|--------|----------|
| `w=4` | 6-across | Compact stat tiles (Game Theater Observe) |
| `w=6` | 4-across | Standard stat tiles (System Health Observe, Perf Test SLO) |
| `w=8` | 3-across | Context panels, sparkline stats |
| `w=12` | 2-across | Primary analysis panels (default for timeseries) |
| `w=24` | Full-width hero | Lifecycle Timing, Phase Tapestry, heatmaps |

### Description Format

Every panel has a description following:

```
Normal: <what healthy looks like>. Watch for: <degradation signals>.
Check next: <where to go for deeper investigation>.
```

### Stat Panel `colorMode`

- `'background'` — for SLO command-center tiles (threshold colors fill the entire panel background)
- `'value'` (default) — for regular stats (only the number changes color)

### Exemplar Support

Exemplars link histogram data points to specific traces in Tempo. Rules:

- **Timeseries panels with `histogram_quantile`**: Always enable exemplars
  - `spanDuration(..., exemplars=true)`
  - `spanDurationBy(..., exemplars=true)`
  - `histogramQuantileTargetsWithExemplars(...)` (drop-in for `histogramQuantileTargets`)
  - `perfTestMoveDurations(..., exemplars=true)`
  - `target(..., exemplar=true)` for custom histogram queries
- **Stat panels, bar gauges, heatmaps**: Never enable exemplars (cannot render dots)

---

## Cross-Dashboard Navigation

### Dashboard UIDs

Registered in `links.libsonnet`:

| Constant | UID | Dashboard |
|----------|-----|-----------|
| `systemHealth` | `system-health` | Infrastructure & RED overview |
| `gameEngine` | `game-engine` | Game lifecycle, phase timing, event bus |
| `gameTheater` | `game-theater` | Single-game deep dive |
| `perfTest` | `perf-test` | Load test analysis |

### Link Constructors

```jsonnet
// Basic cross-dashboard link (preserves time range)
links.toDashboard('System Health', links.dashboardUids.systemHealth)

// Link with variable pre-population (preserves time range + sets variable)
links.toDashboardWithVar('Game Theater', links.dashboardUids.gameTheater, 'gameId', '${gameId}')
```

### Applying Links

Two patterns for adding data links to panels:

```jsonnet
// Pattern 1: modifiers.withLinks() — preferred for clean composition
panel + modifiers.withLinks([link1, link2])

// Pattern 2: manual merge — when other fieldConfig merges are needed
panel + { fieldConfig+: { defaults+: { links: [link1] } } }
```

### Variable Contract

Dashboards that define template variables participate in the variable contract.
Links that pre-populate variables must use the canonical names:

| Variable | Type | Defined on | Purpose |
|----------|------|------------|---------|
| `$gameId` | text | game-engine, game-theater | Filter to single game instance |
| `$traceId` | text | system-health, game-theater | Deep-link into Tempo trace view |

### Navigation Map

```
                    ┌──────────────┐
         ┌────────→│ System Health │←────────┐
         │         └──────┬───────┘         │
         │                │                  │
         │         ┌──────▼───────┐         │
         │    ┌───→│ Game Engine  │←──┐     │
         │    │    └──────┬───────┘   │     │
         │    │           │ ($gameId) │     │
         │    │    ┌──────▼───────┐   │     │
         │    │    │ Game Theater │───┘     │
         │    │    └──────────────┘         │
         │    │                              │
    ┌────┴────┴───┐                         │
    │  Perf Test  │─────────────────────────┘
    └─────────────┘
```

---

## How to Add a New Panel

1. **Choose panel type** from `panels.libsonnet`:
   - Latency over time: `timeseriesPanel`
   - Current value with thresholds: `statPanel`
   - Distribution shape: `heatmapPanel`
   - Ranked comparison: `barGaugePanel`
   - Logs: `logPanel`
   - Traces: `tracesPanel`
   - Phase progression: `stateTimelinePanel`
   - Percentile bands: `spanPercentileBandsPanel`

2. **Build targets** using `targets.libsonnet` helpers:
   ```jsonnet
   // Spanmetrics (preferred for RED)
   targets.spanDuration(targets.spans.gameLogic, [['0.95', 'p95']], exemplars=true)
   targets.spanRate(targets.spans.http, 'req/s')
   targets.spanDurationBy(targets.spans.eventHandler, '0.95', 'span_name', '{{span_name}}', exemplars=true)

   // Manual metrics
   targets.target('my_gauge{service_name="risk-it"}', 'value', 'A')
   targets.histogramQuantileTargetsWithExemplars('my_bucket', [['0.95', 'p95']])
   ```

3. **Compose the panel** with modifiers:
   ```jsonnet
   layout.panel(
     panels.timeseriesPanel(
       title='My Panel',
       targets=targets.spanDuration(targets.spans.db, [['0.95', 'p95']], exemplars=true),
       unit='s',
       color=colors.fixedColor(colors.db),
     ) + modifiers.withSloThreshold(thresholds.dbTxnP95)
       + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
     w=12, h=8,
     description='Normal: < 50ms. Watch for: > 50ms sustained. Check next: Pool Usage.',
   ),
   ```

4. **Place in OODA section** (observe/orient/decide/act or their depth variants).

5. **Regenerate**: `make dashboards`

---

## How to Add a New Dashboard

1. **Register the UID** in `links.libsonnet`:
   ```jsonnet
   dashboardUids: {
     // ... existing ...
     myDashboard: 'my-dashboard',
   },
   ```

2. **Create the Jsonnet file** at `grafana/dashboards/my-dashboard.jsonnet`:
   ```jsonnet
   local dashboard = import 'dashboard.libsonnet';
   local layout = import 'layout.libsonnet';
   local panels = import 'panels.libsonnet';
   local targets = import 'targets.libsonnet';

   dashboard.new(
     uid='my-dashboard',
     title='My Dashboard',
     description='What this dashboard shows',
     tags=['risk-it', 'my-dashboard'],
     panels=layout.ooda(
       observe=[
         layout.panel(
           panels.statPanel(
             title='My Stat',
             targets=[targets.target('up', 'up')],
             thresholds={ mode: 'absolute', steps: [{ color: 'green', value: null }] },
           ),
           w=8, h=6,
           description='Normal: 1. Watch for: 0. Check next: logs.',
         ),
       ],
     ),
   )
   ```

3. **Document variables** in the `links.libsonnet` variable contract table
   if the dashboard defines template variables.

4. **Add cross-dashboard links** on relevant panels in existing dashboards
   using `links.toDashboard()`.

5. **Compile and verify**:
   ```bash
   make dashboards
   make dashboards-check
   ```
