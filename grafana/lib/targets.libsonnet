// Reusable Prometheus target generators for go-risk-it Grafana dashboards.
// Contains target constructors, histogram quantile helpers, spanmetrics helpers,
// and shared lifecycle targets/overrides.
//
// Two metric families:
//   1. Manual metrics (service_name label) — emitted by app code or perf-test client.
//      Helpers: histogramQuantileTargets, histogramQuantileTargetsWithExemplars.
//   2. Spanmetrics (service_name label) — derived by the OTel Collector spanmetrics connector from traces.
//      Helpers: spanDuration, spanRate, spanErrorRate.
local colors = import 'colors.libsonnet';
{
  // ── Service name constants ──
  // Use these instead of hardcoding service names in PromQL expressions.
  serviceName:: 'risk-it',
  perfTestServiceName:: 'perftest',

  // ── Datasource UIDs ──
  // Single source of truth for all datasource references.
  // Matches provisioned UIDs in grafana/provisioning/datasources/.
  datasources:: {
    prometheus: { type: 'prometheus', uid: 'prometheus' },
    loki: { type: 'loki', uid: 'loki' },
    tempo: { type: 'tempo', uid: 'tempo' },
  },

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

  // Build a Loki log query target.
  // expr: LogQL string, legend: legendFormat string.
  lokiTarget(expr, legend, refId='A'):: {
    datasource: $.datasources.loki,
    expr: expr,
    legendFormat: legend,
    refId: refId,
  },

  // Build a Tempo trace query target.
  // query: TraceQL string.
  tempoTarget(query, refId='A'):: {
    datasource: $.datasources.tempo,
    queryType: 'traceql',
    query: query,
    refId: refId,
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

  // ══════════════════════════════════════════════════════════════════
  // Spanmetrics helpers (service_name label)
  // Derived by the OTel Collector spanmetrics connector from traces.
  // Metric names: traces_span_metrics_duration_milliseconds_bucket,
  //               traces_span_metrics_calls_total
  // ══════════════════════════════════════════════════════════════════

  // Base metric names produced by the spanmetrics connector.
  // Duration is in milliseconds — histogram_quantile results are divided by 1000
  // to produce seconds for dashboard display.
  spanmetricsMetric:: {
    duration: 'traces_span_metrics_duration_milliseconds_bucket',
    calls: 'traces_span_metrics_calls_total',
  },

  serviceLabel:: 'service_name',

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
        expr: 'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le)) / 1000' % [q[0], $.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, spanNameFilter, extraLabels],
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
      'sum(rate(%s{%s="%s", span_name=~"%s", status_code!="STATUS_CODE_ERROR"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceLabel, $.serviceName, spanNameFilter, extraLabels],
      legend,
    ),

  // Generate an error rate target from spanmetrics calls_total (status_code=STATUS_CODE_ERROR).
  // spanNameFilter: regex string matching span_name.
  // legend: legendFormat string.
  // extraLabels: string (optional) — additional label matchers.
  spanErrorRate(spanNameFilter, legend, extraLabels='')::
    $.target(
      'sum(rate(%s{%s="%s", span_name=~"%s", status_code="STATUS_CODE_ERROR"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceLabel, $.serviceName, spanNameFilter, extraLabels],
      legend,
    ),

  // Generate a rate target from spanmetrics calls_total grouped by a label.
  // spanNameFilter: regex string matching span_name.
  // groupBy: label to group by (e.g. 'span_name', 'phase', 'event_type').
  // legend: legendFormat template string (e.g. '{{span_name}}').
  // extraLabels: string (optional) — additional label matchers.
  spanRateBy(spanNameFilter, groupBy, legend, extraLabels='')::
    $.target(
      'sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (%s)' % [$.spanmetricsMetric.calls, $.serviceLabel, $.serviceName, spanNameFilter, extraLabels, groupBy],
      legend,
    ),

  // Generate histogram_quantile targets from spanmetrics, grouped by a label.
  // spanNameFilter: regex string matching span_name.
  // quantile: string (e.g. '0.95').
  // groupBy: label to include in by clause alongside le.
  // legend: legendFormat template string.
  // extraLabels: string (optional) — additional label matchers.
  // exemplars: bool (default false) — enable trace exemplar support.
  spanDurationBy(spanNameFilter, quantile, groupBy, legend, extraLabels='', exemplars=false)::
    $.target(
      'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le, %s)) / 1000' % [quantile, $.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, spanNameFilter, extraLabels, groupBy],
      legend,
      exemplar=exemplars,
    ),

  // ══════════════════════════════════════════════════════════════════
  // Perf-test spanmetrics helpers (perftest service)
  // Derived from perftest.game.run and perftest.move.execute spans.
  // Same query patterns as server spanmetrics, different service name.
  // ══════════════════════════════════════════════════════════════════

  // Span name constants for perf-test spans.
  perfTestSpans:: {
    move: 'perftest[.]move[.]execute',
    game: 'perftest[.]game[.]run',
  },

  // Rate of perftest.move.execute spans (successful + error).
  // filters: string (optional) — additional label matchers, e.g. ',action="attack"'.
  perfTestMoveRate(filters='')::
    $.target(
      'sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.move, filters],
      'moves/s',
    ),

  // Duration histogram of perftest.move.execute spans.
  // quantile: string (e.g. '0.95').
  // filters: string (optional) — additional label matchers.
  perfTestMoveDuration(quantile, filters='')::
    $.target(
      'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le))' % [quantile, $.spanmetricsMetric.duration, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.move, filters],
      'p' + std.strReplace(std.strReplace(quantile, '0.', ''), '.', ''),
    ),

  // Multiple duration percentiles for perftest.move.execute (convenience wrapper).
  // quantiles: array of [quantile_str, label] pairs.
  // filters: string (optional) — additional label matchers.
  // exemplars: bool (default false).
  perfTestMoveDurations(quantiles, filters='', exemplars=false)::
    [
      {
        expr: 'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le))' % [q[0], $.spanmetricsMetric.duration, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.move, filters],
        legendFormat: q[1],
        refId: std.char(65 + i),
        [if exemplars then 'exemplar']: true,
      }
      for i in std.range(0, std.length(quantiles) - 1)
      for q in [quantiles[i]]
    ],

  // Rate of perftest.game.run spans.
  // filters: string (optional) — additional label matchers, e.g. ',outcome="completed"'.
  perfTestGameRate(filters='')::
    $.target(
      'sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.game, filters],
      'games/s',
    ),

  // Duration histogram of perftest.game.run spans.
  // quantile: string (e.g. '0.95').
  // filters: string (optional) — additional label matchers.
  perfTestGameDuration(quantile, filters='')::
    $.target(
      'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le))' % [quantile, $.spanmetricsMetric.duration, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.game, filters],
      'p' + std.strReplace(std.strReplace(quantile, '0.', ''), '.', ''),
    ),

  // Multiple duration percentiles for perftest.game.run (convenience wrapper).
  // quantiles: array of [quantile_str, label] pairs.
  // filters: string (optional) — additional label matchers.
  // exemplars: bool (default false).
  perfTestGameDurations(quantiles, filters='', exemplars=false)::
    [
      {
        expr: 'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le))' % [q[0], $.spanmetricsMetric.duration, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.game, filters],
        legendFormat: q[1],
        refId: std.char(65 + i),
        [if exemplars then 'exemplar']: true,
      }
      for i in std.range(0, std.length(quantiles) - 1)
      for q in [quantiles[i]]
    ],

  // Error rate of perftest.move.execute spans (status_code=STATUS_CODE_ERROR).
  // filters: string (optional) — additional label matchers.
  perfTestMoveErrorRate(filters='')::
    $.target(
      'sum(rate(%s{%s="%s", span_name=~"%s", status_code="STATUS_CODE_ERROR"%s}[1m]))' % [$.spanmetricsMetric.calls, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.move, filters],
      'move errors/s',
    ),

  // Rate of perftest.move.execute spans grouped by a label.
  // groupBy: label to group by (e.g. 'action', 'phase').
  // legend: legendFormat template string (e.g. '{{action}}').
  // filters: string (optional) — additional label matchers.
  perfTestMoveRateBy(groupBy, legend, filters='')::
    $.target(
      'sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (%s)' % [$.spanmetricsMetric.calls, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.move, filters, groupBy],
      legend,
    ),

  // Duration histogram of perftest.move.execute grouped by a label.
  // quantile: string (e.g. '0.95').
  // groupBy: label to include in by clause alongside le.
  // legend: legendFormat template string.
  // filters: string (optional) — additional label matchers.
  // exemplars: bool (default false) — enable trace exemplar support.
  perfTestMoveDurationBy(quantile, groupBy, legend, filters='', exemplars=false)::
    $.target(
      'histogram_quantile(%s, sum(rate(%s{%s="%s", span_name=~"%s"%s}[1m])) by (le, %s))' % [quantile, $.spanmetricsMetric.duration, $.serviceLabel, $.perfTestServiceName, $.perfTestSpans.move, filters, groupBy],
      legend,
      exemplar=exemplars,
    ),

  // ── Lifecycle boundary targets (spanmetrics) ──

  // Standard lifecycle latency targets (p95) for the 5 server boundaries.
  // Uses spanmetrics (service_name label) for server-side instrumentation.
  // Used by: system-health (Lifecycle Timing) and perf-test (Latency Attribution).
  spanLifecycleTargets:: [
    {
      refId: 'A',
      expr: 'histogram_quantile(0.95, sum(rate(%s{%s="%s", span_name=~"%s"}[1m])) by (le)) / 1000' % [$.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, $.spans.http],
      legendFormat: 'HTTP Total',
      exemplar: true,
    },
    {
      refId: 'B',
      expr: 'histogram_quantile(0.95, sum(rate(%s{%s="%s", span_name=~"%s"}[1m])) by (le)) / 1000' % [$.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, $.spans.db],
      legendFormat: 'DB Transaction',
      exemplar: true,
    },
    {
      refId: 'C',
      expr: 'histogram_quantile(0.95, sum(rate(%s{%s="%s", span_name=~"%s"}[1m])) by (le)) / 1000' % [$.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, $.spans.gameLogic],
      legendFormat: 'Game Logic',
      exemplar: true,
    },
    {
      refId: 'D',
      expr: 'histogram_quantile(0.95, sum(rate(%s{%s="%s", span_name=~"%s"}[1m])) by (le)) / 1000' % [$.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, $.spans.wsBroadcast],
      legendFormat: 'WS Broadcast',
      exemplar: true,
    },
    {
      refId: 'E',
      expr: 'histogram_quantile(0.95, sum(rate(%s{%s="%s", span_name=~"%s"}[1m])) by (le)) / 1000' % [$.spanmetricsMetric.duration, $.serviceLabel, $.serviceName, $.spans.eventHandler],
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
