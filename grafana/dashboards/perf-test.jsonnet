// Perf Test (Client-Side) dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/perf-test.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local ooda = import 'ooda.libsonnet';
local thresholds = import 'thresholds.libsonnet';

// Local helper: histogram_quantile targets for perftest client metrics.
// The shared common.histogramQuantileTargets() hardcodes service_name="risk-it",
// but perf-test client metrics use service_name="perftest".
local perfHistTargets(metric, quantiles) =
  [
    {
      expr: 'histogram_quantile(%s, sum(rate(%s{service_name="perftest"}[1m])) by (le))' % [q[0], metric],
      legendFormat: q[1],
      refId: std.char(65 + i),
    }
    for i in std.range(0, std.length(quantiles) - 1)
    for q in [quantiles[i]]
  ];

// Local helper: right-axis override for "active games" series on correlation panels.
local activeGamesRightAxis = {
  matcher: { id: 'byName', options: 'active games' },
  properties: [
    { id: 'custom.axisPlacement', value: 'right' },
    { id: 'unit', value: 'short' },
    { id: 'custom.fillOpacity', value: 5 },
    { id: 'color', value: colors.fixedColor(colors.http) },
  ],
};

{
  uid: 'perf-test',
  title: 'Perf Test (Client-Side)',
  description: 'Load test client metrics from the perf-test harness',
  schemaVersion: 39,
  version: 1,
  timezone: 'browser',
  editable: true,
  graphTooltip: 1,
  time: { from: 'now-15m', to: 'now' },
  refresh: '5s',
  templating: { list: [] },
  tags: ['perf-test', 'risk-it', 'load-test'],

  annotations: {
    list: [
      {
        builtIn: 1,
        datasource: { type: 'grafana', uid: '-- Grafana --' },
        enable: true,
        hide: false,
        iconColor: 'rgba(0, 211, 255, 1)',
        name: 'Perf Test Phases',
        type: 'dashboard',
        target: { matchAny: true, tags: ['perf-test'], type: 'tags' },
      },
    ],
  },

  panels: [
    // ── Observe — Am I OK? ───────────────────────────────────────────
    ooda.observeRow() + { gridPos: { h: 1, w: 24, x: 0, y: 0 } },

    // Panel 1: Active Games (stat)
    common.statPanel(
      title='Active Games',
      targets=[
        {
          expr: 'perftest_games_active{service_name="perftest"}',
          legendFormat: 'active',
          refId: 'A',
        },
      ],
      thresholds=thresholds.perfTestActiveGames,
    ) + {
      id: 1,
      gridPos: { h: 6, w: 6, x: 0, y: 1 },
    },

    // Panel 2: Moves/sec (timeseries, fillOpacity=15)
    common.timeseriesPanel(
      title='Moves/sec',
      targets=[
        {
          expr: 'rate(perftest_moves_total{service_name="perftest"}[30s])',
          legendFormat: 'moves/s',
          refId: 'A',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.client),
    ) + {
      id: 2,
      gridPos: { h: 6, w: 9, x: 6, y: 1 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
        },
      },
    },

    // Panel 3: Game Completion (timeseries, per-series color overrides)
    common.timeseriesPanel(
      title='Game Completion',
      targets=[
        {
          expr: 'perftest_games_completed_total{service_name="perftest"}',
          legendFormat: 'completed',
          refId: 'A',
        },
        {
          expr: 'perftest_games_timed_out_total{service_name="perftest"}',
          legendFormat: 'timed out',
          refId: 'B',
        },
        {
          expr: 'perftest_games_fatal_total{service_name="perftest"}',
          legendFormat: 'fatal',
          refId: 'C',
        },
      ],
      unit='short',
      overrides=[
        {
          matcher: { id: 'byName', options: 'fatal' },
          properties: [{ id: 'color', value: colors.fixedColor(colors.errors) }],
        },
        {
          matcher: { id: 'byName', options: 'timed out' },
          properties: [{ id: 'color', value: colors.fixedColor(colors.http) }],
        },
      ],
    ) + {
      id: 3,
      gridPos: { h: 6, w: 9, x: 15, y: 1 },
      options+: {
        legend+: {
          calcs: ['lastNotNull'],
        },
      },
    },

    // ── Orient — What's the shape? ───────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 7 } },

    // Panel 13: Client E2E Latency Distribution (heatmap, YlOrRd scheme)
    common.heatmapPanel(
      title='Client E2E Latency Distribution',
      targets=[
        {
          expr: 'sum(rate(perftest_e2e_duration_seconds_bucket{service_name="perftest"}[$__rate_interval])) by (le)',
          format: 'heatmap',
          legendFormat: '{{le}}',
          refId: 'A',
        },
      ],
      unit='s',
      colorScheme='YlOrRd',
      colorFill='dark-red',
    ) + {
      id: 13,
      gridPos: { h: 8, w: 24, x: 0, y: 8 },
    },

    // Panel 5: E2E Move Latency P50/P95/P99 (SLO threshold overlay)
    common.timeseriesPanel(
      title='E2E Move Latency (P50/P95/P99)',
      targets=perfHistTargets(
        'perftest_e2e_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.client),
    ) + {
      id: 5,
      gridPos: { h: 8, w: 12, x: 12, y: 16 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.e2eP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
        },
      },
    },

    // Panel 6: WS Delivery Latency P50/P95/P99 (SLO threshold overlay)
    common.timeseriesPanel(
      title='WS Delivery Latency (P50/P95/P99)',
      targets=perfHistTargets(
        'perftest_ws_delivery_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.ws),
    ) + {
      id: 6,
      gridPos: { h: 8, w: 12, x: 0, y: 16 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.wsDeliveryP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
        },
      },
    },

    // ── Decide — Where's the bottleneck? ─────────────────────────────
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 24 } },

    // Panel 4: REST Latency by Action P50/P95/P99 (dynamic labels, palette-classic)
    common.timeseriesPanel(
      title='REST Latency by Action (P50/P95/P99)',
      targets=[
        {
          expr: 'histogram_quantile(0.5, sum(rate(perftest_rest_duration_seconds_bucket{service_name="perftest"}[1m])) by (le, action))',
          legendFormat: '{{action}} p50',
          refId: 'A',
        },
        {
          expr: 'histogram_quantile(0.95, sum(rate(perftest_rest_duration_seconds_bucket{service_name="perftest"}[1m])) by (le, action))',
          legendFormat: '{{action}} p95',
          refId: 'B',
        },
        {
          expr: 'histogram_quantile(0.99, sum(rate(perftest_rest_duration_seconds_bucket{service_name="perftest"}[1m])) by (le, action))',
          legendFormat: '{{action}} p99',
          refId: 'C',
        },
      ],
      unit='s',
    ) + {
      id: 4,
      gridPos: { h: 8, w: 12, x: 0, y: 25 },
    },

    // Panel 8: Conflicts/sec (fixed orange, fillOpacity=15)
    common.timeseriesPanel(
      title='Conflicts/sec',
      targets=[
        {
          expr: 'rate(perftest_conflicts_total{service_name="perftest"}[30s])',
          legendFormat: 'conflicts/s',
          refId: 'A',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.http),
    ) + {
      id: 8,
      gridPos: { h: 8, w: 12, x: 12, y: 25 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
        },
      },
    },

    // Panel 9: E2E Latency vs Concurrency (right-axis override for active games)
    common.timeseriesPanel(
      title='E2E Latency vs Concurrency',
      targets=[
        {
          expr: 'histogram_quantile(0.95, sum(rate(perftest_e2e_duration_seconds_bucket{service_name="perftest"}[1m])) by (le))',
          legendFormat: 'E2E p95',
          refId: 'A',
        },
        {
          expr: 'perftest_games_active{service_name="perftest"}',
          legendFormat: 'active games',
          refId: 'B',
        },
      ],
      unit='s',
      overrides=[activeGamesRightAxis],
    ) + {
      id: 9,
      description: 'Shows the performance curve: at what concurrency does latency degrade?',
      gridPos: { h: 8, w: 12, x: 0, y: 33 },
    },

    // Panel 10: Server Latency vs Concurrency (right-axis override for active games)
    common.timeseriesPanel(
      title='Server Latency vs Concurrency',
      targets=[
        {
          expr: 'histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'HTTP p95',
          refId: 'A',
        },
        {
          expr: 'perftest_games_active{service_name="perftest"}',
          legendFormat: 'active games',
          refId: 'B',
        },
      ],
      unit='s',
      overrides=[activeGamesRightAxis],
    ) + {
      id: 10,
      description: 'Server-side HTTP P95 overlaid with active game count',
      gridPos: { h: 8, w: 12, x: 12, y: 33 },
    },

    // ── Act — What's the evidence? ───────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 41 } },

    // Panel 7: Error Rate by Type (stacking normal, palette-classic)
    common.timeseriesPanel(
      title='Error Rate by Type',
      targets=[
        {
          expr: 'sum(rate(perftest_errors_total{service_name="perftest"}[1m])) by (type)',
          legendFormat: '{{type}}',
          refId: 'A',
        },
      ],
      unit='ops',
    ) + {
      id: 7,
      gridPos: { h: 8, w: 8, x: 0, y: 42 },
      fieldConfig+: {
        defaults+: {
          color: { mode: 'palette-classic' },
          custom+: {
            stacking: { group: 'A', mode: 'normal' },
          },
        },
      },
    },

    // Panel 11: Game Duration P50/P95
    common.timeseriesPanel(
      title='Game Duration (P50/P95)',
      targets=perfHistTargets(
        'perftest_game_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95']],
      ),
      unit='s',
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 11,
      gridPos: { h: 8, w: 8, x: 8, y: 42 },
    },

    // Panel 12: Moves per Game P50/P95
    common.timeseriesPanel(
      title='Moves per Game (P50/P95)',
      targets=perfHistTargets(
        'perftest_game_moves_bucket',
        [['0.5', 'p50'], ['0.95', 'p95']],
      ),
      unit='short',
      color=colors.fixedColor(colors.client),
    ) + {
      id: 12,
      gridPos: { h: 8, w: 8, x: 16, y: 42 },
    },
  ],
}
