// Request Lifecycle dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/request-lifecycle.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local links = import 'links.libsonnet';
local thresholds = import 'thresholds.libsonnet';
local ooda = import 'ooda.libsonnet';

// Template variables for trace and game correlation.
local traceIdVar = {
  name: 'traceId',
  type: 'textbox',
  label: 'Trace ID',
  current: { value: '' },
};
local gameIdVar = {
  name: 'gameId',
  type: 'textbox',
  label: 'Game ID',
  current: { value: '' },
};

{
  uid: 'request-lifecycle',
  title: 'Request Lifecycle',
  description: 'Full request lifecycle: HTTP ingress through game logic, DB transaction, WS broadcast, and async event handling',
  schemaVersion: 39,
  version: 1,
  timezone: 'browser',
  editable: true,
  graphTooltip: 1,
  tags: ['risk-it', 'lifecycle'],
  time: { from: 'now-15m', to: 'now' },
  refresh: '10s',

  templating: {
    list: [traceIdVar, gameIdVar],
  },

  annotations: { list: [] },

  panels: [
    // ── Observe — Am I OK? ──────────────────────────────────────────
    ooda.observeRow() + { gridPos: { h: 1, w: 24, x: 0, y: 0 } },

    // Panel 1: Full Lifecycle Timing (hero panel, w=24)
    // Shows 5 latency series: HTTP Total, DB Transaction, Game Logic, WS Broadcast (solid),
    // Event Handler (dashed, post-response async work).
    common.timeseriesPanel(
      title='Full Lifecycle Timing (p95)',
      targets=[
        {
          refId: 'A',
          expr: 'histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'HTTP Total',
          exemplar: true,
        },
        {
          refId: 'B',
          expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'DB Transaction',
          exemplar: true,
        },
        {
          refId: 'C',
          expr: 'histogram_quantile(0.95, sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'Game Logic',
          exemplar: true,
        },
        {
          refId: 'D',
          expr: 'histogram_quantile(0.95, sum(rate(ws_broadcast_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'WS Broadcast',
          exemplar: true,
        },
        {
          refId: 'E',
          expr: 'histogram_quantile(0.95, sum(rate(event_handler_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'Event Handler (post-response)',
          exemplar: true,
        },
      ],
      unit='s',
      overrides=[
        {
          matcher: { id: 'byName', options: 'HTTP Total' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.http) },
            { id: 'custom.lineWidth', value: 3 },
          ],
        },
        {
          matcher: { id: 'byName', options: 'DB Transaction' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.db) },
          ],
        },
        {
          matcher: { id: 'byName', options: 'Game Logic' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.gameLogic) },
          ],
        },
        {
          matcher: { id: 'byName', options: 'WS Broadcast' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.ws) },
          ],
        },
        {
          matcher: { id: 'byName', options: 'Event Handler (post-response)' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.eventBus) },
            { id: 'custom.lineStyle', value: { fill: 'dash', dash: [10, 10] } },
            { id: 'custom.fillOpacity', value: 0 },
          ],
        },
      ],
    ) + {
      id: 1,
      description: 'P95 latency for each boundary in the request lifecycle overlaid. HTTP Total is the synchronous request; Event Handler is async post-response work (dashed). Normal: DB + Game Logic + WS sum to roughly HTTP Total; Event Handler runs independently after response. Watch for: Event Handler exceeding HTTP Total (async work slower than the request itself). Check next: Event Handler Latency panel for per-consumer breakdown.',
      gridPos: { h: 10, w: 24, x: 0, y: 1 },
      fieldConfig+: {
        defaults+: {
          links: [
            links.toDashboard('Database', links.dashboardUids.database),
            links.toDashboard('WebSocket', links.dashboardUids.websocket),
            links.toDashboard('Game Engine', links.dashboardUids.gameEngine),
            links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter),
          ],
        },
      },
    },

    // Panel 2: E2E Percentile Bands (client-side end-to-end latency)
    common.percentileBandsPanel(
      title='E2E Latency Percentile Bands',
      metric='http_server_request_duration_seconds_bucket',
      unit='s',
    ) + {
      id: 2,
      description: 'Server-side HTTP request latency with filled bands between P50, P95, and P99. Normal: tight bands with P95 < 200ms. Watch for: bands widening (latency variance increasing) or P99 diverging sharply from P95 (tail latency problem). Check next: Full Lifecycle Timing to identify which boundary causes the spread.',
      gridPos: { h: 8, w: 24, x: 0, y: 11 },
      fieldConfig+: {
        defaults+: {
          links: [
            links.toDashboard('Perf Test Detail', links.dashboardUids.perfTest),
          ],
        },
      },
    },

    // ── Orient — What's the shape? ──────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 19 } },

    // Panel 3: Event Handler Latency (p50/p95/p99 per consumer)
    common.timeseriesPanel(
      title='Event Handler Latency by Consumer',
      targets=[
        {
          refId: 'A',
          expr: 'histogram_quantile(0.5, sum(rate(event_handler_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, handler))',
          legendFormat: '{{handler}} p50',
        },
        {
          refId: 'B',
          expr: 'histogram_quantile(0.95, sum(rate(event_handler_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, handler))',
          legendFormat: '{{handler}} p95',
        },
        {
          refId: 'C',
          expr: 'histogram_quantile(0.99, sum(rate(event_handler_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, handler))',
          legendFormat: '{{handler}} p99',
        },
      ],
      unit='s',
    ) + {
      id: 3,
      description: 'Event handler latency broken down by consumer (e.g. WS broadcast, headline detection, logging) at P50/P95/P99. Normal: all handlers < 100ms at P95. Watch for: single handler diverging (e.g. WS broadcast spiking while others stay flat). Check next: WebSocket dashboard if broadcast handler dominates.',
      gridPos: { h: 8, w: 12, x: 0, y: 20 },
      fieldConfig+: {
        defaults+: {
          links: [
            links.toDashboard('WebSocket Detail', links.dashboardUids.websocket),
          ],
        },
      },
    },

    // Panel 4: Bus Dispatch Duration (time from event emission to handler start)
    common.timeseriesPanel(
      title='Bus Dispatch Duration',
      targets=common.histogramQuantileTargetsWithExemplars(
        'event_handler_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.eventBus),
    ) + {
      id: 4,
      description: 'Aggregate event handler duration P50/P95/P99 across all consumers. Normal: P95 < 50ms. Watch for: P95 growing over time (handler goroutines competing for resources). Check next: Handler Throughput to correlate duration increase with event volume.',
      gridPos: { h: 8, w: 12, x: 12, y: 20 },
    },

    // ── Decide — Where's the bottleneck? ────────────────────────────
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 28 } },

    // Panel 5: Handler Throughput (events/sec by type)
    common.timeseriesPanel(
      title='Handler Throughput by Event Type',
      targets=[
        {
          refId: 'A',
          expr: 'sum(rate(event_bus_events_total{service_name="risk-it"}[1m])) by (event_type)',
          legendFormat: '{{event_type}}',
        },
      ],
      unit='ops',
    ) + {
      id: 5,
      description: 'Event bus throughput broken down by event type (MoveExecuted, PhaseTransitioned, GameCompleted, etc.). Normal: MoveExecuted is highest, proportional to game moves/s. Watch for: event types dropping to zero (handler registration issue) or unexpected spikes (cascade events). Check next: Game Engine dashboard for per-phase move rates.',
      gridPos: { h: 8, w: 24, x: 0, y: 29 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
          links: [
            links.toDashboard('Game Engine', links.dashboardUids.gameEngine),
          ],
        },
      },
    },

    // ── Act — What's the evidence? ──────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 37 } },

    // Panel 6: Placeholder — reserved for future Phase 5 panels
    common.statPanel(
      title='Event Bus Events (Total)',
      targets=[{
        refId: 'A',
        expr: 'sum(event_bus_events_total{service_name="risk-it"})',
        legendFormat: 'Total Events',
      }],
      thresholds={ mode: 'absolute', steps: [{ color: 'green', value: null }] },
    ) + {
      id: 6,
      description: 'Total events processed by the event bus since startup. Normal: grows proportionally to game activity. Watch for: stagnation (bus stopped processing) or counter resets (service restart). Check next: Handler Throughput for rate view.',
      gridPos: { h: 8, w: 12, x: 0, y: 38 },
      fieldConfig+: {
        defaults+: {
          color: { mode: 'palette-classic' },
        },
      },
    },

    // Panel 7: Event Handler Error Rate (if errors are tracked)
    common.timeseriesPanel(
      title='Event Handler Errors',
      targets=[
        {
          refId: 'A',
          expr: 'sum(rate(event_handler_errors_total{service_name="risk-it"}[1m])) by (handler)',
          legendFormat: '{{handler}}',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.errors),
    ) + {
      id: 7,
      description: 'Error rate in event handlers by consumer. Normal: zero. Watch for: any sustained error rate indicates handler failures (panics recovered by safego.Go, DB errors, etc.). Check next: Server Golden Signals for correlated HTTP errors.',
      gridPos: { h: 8, w: 12, x: 12, y: 38 },
      fieldConfig+: {
        defaults+: {
          links: [
            links.toDashboard('Server Golden Signals', links.dashboardUids.serverGoldenSignals),
          ],
        },
      },
    },

    // Panel 8: Recent Traces table (Tempo TraceQL search)
    {
      id: 8,
      title: 'Recent Traces',
      description: 'Normal: Recent request traces. Watch for: High error rates or missing services. Check next: Copy a Trace ID and paste into the Trace ID textbox above to load the waterfall.',
      type: 'table',
      datasource: { type: 'tempo', uid: 'tempo' },
      targets: [{
        refId: 'A',
        queryType: 'traceqlSearch',
        query: '{resource.service.name="risk-it"}',
        limit: 20,
        tableType: 'traces',
      }],
      gridPos: { h: 8, w: 24, x: 0, y: 46 },
      fieldConfig: { defaults: {}, overrides: [] },
      options: {},
    },

    // Panel 9: Trace Waterfall (driven by $traceId variable)
    {
      id: 9,
      title: 'Trace Waterfall',
      description: 'Normal: Shows full request lifecycle. Watch for: Missing spans, long gaps. Check next: Correlated Logs panel below.',
      type: 'traces',
      datasource: { type: 'tempo', uid: 'tempo' },
      targets: [{
        refId: 'A',
        queryType: 'traceId',
        query: '${traceId}',
      }],
      gridPos: { h: 12, w: 24, x: 0, y: 54 },
    },

    // Panel 10: Correlated Logs (Loki entries matching selected trace)
    {
      id: 10,
      title: 'Correlated Logs',
      description: 'Normal: Log entries matching the selected trace. Watch for: Error-level entries. Check next: Click trace_id to cross-reference.',
      type: 'logs',
      datasource: { type: 'loki', uid: 'loki' },
      targets: [{
        refId: 'A',
        expr: '{service_name="risk-it"} | trace_id=`${traceId}`',
      }],
      gridPos: { h: 10, w: 24, x: 0, y: 66 },
      options: {
        showTime: true,
        showLabels: false,
        showCommonLabels: false,
        wrapLogMessage: true,
        prettifyLogMessage: false,
        enableLogDetails: true,
        sortOrder: 'Descending',
        dedupStrategy: 'none',
      },
    },
  ],
}
