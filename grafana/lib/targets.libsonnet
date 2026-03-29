// Reusable Prometheus target generators for go-risk-it Grafana dashboards.
// Contains target constructors, histogram quantile helpers, and shared lifecycle targets/overrides.
local colors = import 'colors.libsonnet';
{
  // ── Service name constants ──
  // Use these instead of hardcoding service names in PromQL expressions.
  serviceName:: 'risk-it',
  perfTestServiceName:: 'perftest',

  // ── Target constructors ──

  // Build a single Prometheus target.
  // expr: PromQL string, legend: legendFormat string.
  // refId defaults to 'A'. Set exemplar=true for trace exemplar support.
  target(expr, legend, refId='A', exemplar=false):: {
    expr: expr,
    legendFormat: legend,
    refId: refId,
    [if exemplar then 'exemplar']: true,
  },

  // Build a heatmap target (format:'heatmap', legendFormat:'{{le}}').
  // Caller must pass a PromQL expression that groups by (le).
  heatmapTarget(expr, refId='A'):: {
    expr: expr,
    format: 'heatmap',
    legendFormat: '{{le}}',
    refId: refId,
  },

  // ── Histogram quantile helpers ──

  // Generate histogram_quantile targets for a metric.
  // metric: string (bucket metric name), quantiles: array of [quantile_str, label] pairs
  // quantile_str should be a string like "0.5", "0.95", "0.99" to avoid float precision issues.
  // serviceName: string (OTel service_name label, default 'risk-it'; use 'perftest' for client metrics)
  // Returns array of target objects.
  histogramQuantileTargets(metric, quantiles, serviceName=$.serviceName)::
    [
      {
        expr: 'histogram_quantile(%s, sum(rate(%s{service_name="%s"}[1m])) by (le))' % [q[0], metric, serviceName],
        legendFormat: q[1],
        refId: std.char(65 + i),  // A, B, C, ...
      }
      for i in std.range(0, std.length(quantiles) - 1)
      for q in [quantiles[i]]
    ],

  // Same as histogramQuantileTargets but with exemplar support enabled.
  // Drop-in replacement where trace exemplars are desired on histogram panels.
  histogramQuantileTargetsWithExemplars(metric, quantiles, serviceName=$.serviceName)::
    [
      {
        expr: 'histogram_quantile(%s, sum(rate(%s{service_name="%s"}[1m])) by (le))' % [q[0], metric, serviceName],
        legendFormat: q[1],
        refId: std.char(65 + i),  // A, B, C, ...
        exemplar: true,
      }
      for i in std.range(0, std.length(quantiles) - 1)
      for q in [quantiles[i]]
    ],

  // Per-phase histogram quantile targets for game_phase_duration_seconds_bucket.
  // phase: string (e.g. 'DEPLOY', 'ATTACK').
  // Returns p50/p95/p99 targets for the given phase.
  phaseLatencyTargets(phase):: [
    {
      expr: 'histogram_quantile(%s, sum(rate(game_phase_duration_seconds_bucket{service_name="%s",phase="%s"}[1m])) by (le))' % [q[0], $.serviceName, phase],
      legendFormat: q[1],
      refId: std.char(65 + i),
    }
    for i in std.range(0, 2)
    for q in [[['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']][i]]
  ],

  // ── Lifecycle boundary targets ──

  // Standard lifecycle latency targets (p95) for the 5 server boundaries.
  // Used by: perf-test (Latency Attribution) and system-health (Lifecycle Timing).
  lifecycleTargets:: [
    {
      refId: 'A',
      expr: 'histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="%s"}[1m])) by (le))' % $.serviceName,
      legendFormat: 'HTTP Total',
      exemplar: true,
    },
    {
      refId: 'B',
      expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="%s"}[1m])) by (le))' % $.serviceName,
      legendFormat: 'DB Transaction',
      exemplar: true,
    },
    {
      refId: 'C',
      expr: 'histogram_quantile(0.95, sum(rate(game_phase_duration_seconds_bucket{service_name="%s"}[1m])) by (le))' % $.serviceName,
      legendFormat: 'Game Logic',
      exemplar: true,
    },
    {
      refId: 'D',
      expr: 'histogram_quantile(0.95, sum(rate(ws_broadcast_duration_seconds_bucket{service_name="%s"}[1m])) by (le))' % $.serviceName,
      legendFormat: 'WS Broadcast',
      exemplar: true,
    },
    {
      refId: 'E',
      expr: 'histogram_quantile(0.95, sum(rate(event_handler_duration_seconds_bucket{service_name="%s"}[1m])) by (le))' % $.serviceName,
      legendFormat: 'Event Handler (post-response)',
      exemplar: true,
    },
  ],

  // Standard lifecycle color/style overrides for the 5 server boundaries.
  // HTTP Total: amber, bold line. DB: blue. Game Logic: green.
  // WS Broadcast: purple. Event Handler: teal, dashed, no fill (async).
  lifecycleOverrides:: [
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
}
