// Reusable Prometheus target generators for go-risk-it Grafana dashboards.
// Contains histogram quantile helpers and shared lifecycle targets/overrides.
local colors = import 'colors.libsonnet';
{
  // Generate histogram_quantile targets for a metric.
  // metric: string (bucket metric name), quantiles: array of [quantile_str, label] pairs
  // quantile_str should be a string like "0.5", "0.95", "0.99" to avoid float precision issues.
  // serviceName: string (OTel service_name label, default 'risk-it'; use 'perftest' for client metrics)
  // Returns array of target objects.
  histogramQuantileTargets(metric, quantiles, serviceName='risk-it')::
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
  histogramQuantileTargetsWithExemplars(metric, quantiles, serviceName='risk-it')::
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

  // Standard lifecycle latency targets (p95) for the 5 server boundaries.
  // Used by: perf-test-command-center (Latency Attribution) and
  // request-lifecycle (Full Lifecycle Timing).
  lifecycleTargets:: [
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
