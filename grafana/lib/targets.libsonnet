// Reusable Prometheus target generators for go-risk-it Grafana dashboards.
// Contains target constructors, histogram quantile helpers, spanmetrics helpers,
// and shared lifecycle targets/overrides.
//
// Two metric families:
//   1. Manual metrics (service_name label) — emitted by app code or perf-test client.
//      Helpers: histogramQuantileTargets, histogramQuantileTargetsWithExemplars, phaseLatencyTargets.
//   2. Spanmetrics (service label) — derived by the OTel Collector spanmetrics connector from traces.
//      Helpers: spanDuration, spanRate, spanErrorRate.
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

  // ══════════════════════════════════════════════════════════════════
  // Manual metric helpers (service_name label)
  // Used by perf-test client metrics and any manually-instrumented counters/histograms.
  // ══════════════════════════════════════════════════════════════════

  // Generate histogram_quantile targets for a manual metric.
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

  // ══════════════════════════════════════════════════════════════════
  // Spanmetrics helpers (service label)
  // Derived by the OTel Collector spanmetrics connector from traces.
  // Metric names: traces_spanmetrics_duration_seconds_bucket,
  //               traces_spanmetrics_calls_total
  // ══════════════════════════════════════════════════════════════════

  // Base metric names produced by the spanmetrics connector.
  // The LGTM stack's built-in spanmetrics uses 'latency' naming (seconds).
  spanmetricsMetric:: {
    duration: 'traces_spanmetrics_latency_bucket',
    calls: 'traces_spanmetrics_calls_total',
  },

  // Span name regex patterns for each instrumented domain.
  // Used as the spanNameFilter argument to spanDuration/spanRate/spanErrorRate.
  spans:: {
    http: '(GET|POST|PUT|DELETE) .*',
    db: 'db[.]transaction',
    gameLogic: 'game[.]orchestrate_move',
    wsBroadcast: 'ws[.]broadcast',
    eventHandler: 'consumer[.].*',
    busDispatch: 'bus:.*',
    snapshot: 'snapshot[.].*',
    gameCreate: 'game[.]create',
    lobbyLobbies: 'lobby[.]get_user_lobbies',
  },

  // Generate histogram_quantile targets from spanmetrics duration buckets.
  // spanNameFilter: regex string matching span_name (use spans.* constants).
  // quantiles: array of [quantile_str, label] pairs.
  // exemplars: bool (default false) — enable trace exemplar support.
  // extraLabels: string (optional) — additional label matchers, e.g. ',phase="ATTACK"'.
  spanDuration(spanNameFilter, quantiles, exemplars=false, extraLabels='')::
    [
      {
        expr: 'histogram_quantile(%s, sum(rate(%s{service="%s", span_name=~"%s"%s}[1m])) by (le))' % [q[0], $.spanmetricsMetric.duration, $.serviceName, spanNameFilter, extraLabels],
        legendFormat: q[1],
        refId: std.char(65 + i),
        [if exemplars then 'exemplar']: true,
      }
      for i in std.range(0, std.length(quantiles) - 1)
      for q in [quantiles[i]]
    ],

  // Generate a rate target from spanmetrics calls_total.
  // spanNameFilter: regex string matching span_name.
  // legend: legendFormat string.
  // extraLabels: string (optional) — additional label matchers.
  spanRate(spanNameFilter, legend, extraLabels='')::
    $.target(
      'sum(rate(%s{service="%s", span_name=~"%s", status_code!="STATUS_CODE_ERROR"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceName, spanNameFilter, extraLabels],
      legend,
    ),

  // Generate an error rate target from spanmetrics calls_total (status_code=STATUS_CODE_ERROR).
  // spanNameFilter: regex string matching span_name.
  // legend: legendFormat string.
  // extraLabels: string (optional) — additional label matchers.
  spanErrorRate(spanNameFilter, legend, extraLabels='')::
    $.target(
      'sum(rate(%s{service="%s", span_name=~"%s", status_code="STATUS_CODE_ERROR"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceName, spanNameFilter, extraLabels],
      legend,
    ),

  // Generate a rate target from spanmetrics calls_total grouped by a label.
  // spanNameFilter: regex string matching span_name.
  // groupBy: label to group by (e.g. 'span_name', 'phase', 'event_type').
  // legend: legendFormat template string (e.g. '{{span_name}}').
  // extraLabels: string (optional) — additional label matchers.
  spanRateBy(spanNameFilter, groupBy, legend, extraLabels='')::
    $.target(
      'sum(rate(%s{service="%s", span_name=~"%s"%s}[1m])) by (%s)' % [$.spanmetricsMetric.calls, $.serviceName, spanNameFilter, extraLabels, groupBy],
      legend,
    ),

  // Generate histogram_quantile targets from spanmetrics, grouped by a label.
  // spanNameFilter: regex string matching span_name.
  // quantile: string (e.g. '0.95').
  // groupBy: label to include in by clause alongside le.
  // legend: legendFormat template string.
  // extraLabels: string (optional) — additional label matchers.
  spanDurationBy(spanNameFilter, quantile, groupBy, legend, extraLabels='')::
    $.target(
      'histogram_quantile(%s, sum(rate(%s{service="%s", span_name=~"%s"%s}[1m])) by (le, %s))' % [quantile, $.spanmetricsMetric.duration, $.serviceName, spanNameFilter, extraLabels, groupBy],
      legend,
    ),

  // ── Lifecycle boundary targets (spanmetrics) ──

  // Standard lifecycle latency targets (p95) for the 5 server boundaries.
  // Uses spanmetrics (service label) for server-side instrumentation.
  // Used by: system-health (Lifecycle Timing) and perf-test (Latency Attribution).
  spanLifecycleTargets:: [
    {
      refId: 'A',
      expr: 'histogram_quantile(0.95, sum(rate(%s{service="%s", span_name=~"%s"}[1m])) by (le))' % [$.spanmetricsMetric.duration, $.serviceName, $.spans.http],
      legendFormat: 'HTTP Total',
      exemplar: true,
    },
    {
      refId: 'B',
      expr: 'histogram_quantile(0.95, sum(rate(%s{service="%s", span_name=~"%s"}[1m])) by (le))' % [$.spanmetricsMetric.duration, $.serviceName, $.spans.db],
      legendFormat: 'DB Transaction',
      exemplar: true,
    },
    {
      refId: 'C',
      expr: 'histogram_quantile(0.95, sum(rate(%s{service="%s", span_name=~"%s"}[1m])) by (le))' % [$.spanmetricsMetric.duration, $.serviceName, $.spans.gameLogic],
      legendFormat: 'Game Logic',
      exemplar: true,
    },
    {
      refId: 'D',
      expr: 'histogram_quantile(0.95, sum(rate(%s{service="%s", span_name=~"%s"}[1m])) by (le))' % [$.spanmetricsMetric.duration, $.serviceName, $.spans.wsBroadcast],
      legendFormat: 'WS Broadcast',
      exemplar: true,
    },
    {
      refId: 'E',
      expr: 'histogram_quantile(0.95, sum(rate(%s{service="%s", span_name=~"%s"}[1m])) by (le))' % [$.spanmetricsMetric.duration, $.serviceName, $.spans.eventHandler],
      legendFormat: 'Event Handler (post-response)',
      exemplar: true,
    },
  ],

  // ── Legacy lifecycle boundary targets (manual metrics) ──
  // Kept for backward compatibility — used by perf-test client metrics
  // that report via manual histograms with service_name label.
  lifecycleTargets:: $.spanLifecycleTargets,

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
