// Perf Test (Client-Side) dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/perf-test.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local links = import 'links.libsonnet';
local ooda = import 'ooda.libsonnet';
local thresholds = import 'thresholds.libsonnet';

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
      description: 'Number of games currently in progress. Normal: matches configured concurrency. Watch for: stuck at 0 (no games starting) or exceeding target (games not completing). Check next: Game Completion panel for timeout/fatal counts.',
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
      description: 'Client-side move throughput (30s rate). Normal: proportional to active games. Watch for: sudden drops (server overload) or flat line at zero (test harness stuck). Check next: E2E Move Latency to see if slowdown explains throughput drop.',
      gridPos: { h: 6, w: 9, x: 6, y: 1 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
          links: [links.toDashboard('Game Engine', links.dashboardUids.gameEngine)],
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
      description: 'Cumulative game outcomes: completed, timed out, and fatal. Normal: completed grows steadily, others stay flat. Watch for: timed-out or fatal counts climbing (server cannot finish games in time). Check next: Error Rate by Type for error categorization.',
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
      description: 'Heatmap of end-to-end move latency bucket distribution over time. Normal: dense band below 500ms. Watch for: color spreading into higher buckets (latency distribution widening under load). Check next: E2E Move Latency percentile lines for exact P50/P95/P99 values.',
      gridPos: { h: 8, w: 24, x: 0, y: 8 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter)],
        },
      },
    },

    // Panel 5: E2E Move Latency P50/P95/P99 (SLO threshold overlay)
    common.timeseriesPanel(
      title='E2E Move Latency (P50/P95/P99)',
      targets=common.histogramQuantileTargetsWithExemplars(
        'perftest_e2e_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
        'perftest',
      ),
      unit='s',
      color=colors.fixedColor(colors.client),
    ) + {
      id: 5,
      description: 'End-to-end move latency percentiles measured by the client. SLO threshold overlay at 500ms. Normal: P95 < 500ms, P99 < 1s. Watch for: P95 crossing the threshold line (SLO breach). Check next: REST Latency by Action to identify which move type is slowest.',
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
      targets=common.histogramQuantileTargetsWithExemplars(
        'perftest_ws_delivery_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
        'perftest',
      ),
      unit='s',
      color=colors.fixedColor(colors.ws),
    ) + {
      id: 6,
      description: 'WebSocket state delivery latency percentiles. SLO threshold overlay at 200ms. Normal: P95 < 200ms. Watch for: P95 crossing the threshold line (WS delivery SLO breach). Check next: WebSocket dashboard for connection and broadcast details.',
      gridPos: { h: 8, w: 12, x: 0, y: 16 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.wsDeliveryP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
          links: [links.toDashboard('WebSocket Detail', links.dashboardUids.websocket)],
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
      description: 'REST API latency broken down by move action (deploy, attack, conquer, reinforce, cards) at P50/P95/P99. Normal: all actions < 200ms at P95. Watch for: single action diverging (e.g. attack P95 spiking while others stay flat). Check next: Database dashboard for query-level latency.',
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
      description: 'Rate of client-side move conflicts (HTTP 409 — optimistic lock retry). Normal: low, proportional to concurrency. Watch for: conflicts/s exceeding moves/s (more retries than successes, contention too high). Check next: E2E Latency vs Concurrency to correlate conflict rate with load.',
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
      description: 'E2E P95 latency overlaid with active game count (right axis). Normal: latency stays flat as concurrency increases. Watch for: inflection point where latency climbs sharply with concurrency (saturation). Check next: Server Latency vs Concurrency to see if server-side shows the same knee.',
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
          exemplar: true,
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
      description: 'Server-side HTTP P95 overlaid with active game count (right axis). Normal: HTTP P95 < 100ms regardless of concurrency. Watch for: server latency rising before client E2E does (server is the bottleneck). Check next: Command Center Latency Attribution to identify which server boundary (DB, game logic, WS) dominates.',
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
      description: 'Client-side errors stacked by type (connection, timeout, HTTP 5xx, etc.). Normal: zero or near-zero. Watch for: any sustained error rate, especially new error types appearing. Check next: Game Completion panel to see if errors cause game failures.',
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
      targets=common.histogramQuantileTargetsWithExemplars(
        'perftest_game_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95']],
        'perftest',
      ),
      unit='s',
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 11,
      description: 'Time to complete a full game at P50 and P95. Normal: consistent across the test run. Watch for: game duration growing over time (server slowing down under sustained load). Check next: Moves per Game to check if duration increase is from more moves or slower moves.',
      gridPos: { h: 8, w: 8, x: 8, y: 42 },
    },

    // Panel 12: Moves per Game P50/P95
    common.timeseriesPanel(
      title='Moves per Game (P50/P95)',
      targets=common.histogramQuantileTargetsWithExemplars(
        'perftest_game_moves_bucket',
        [['0.5', 'p50'], ['0.95', 'p95']],
        'perftest',
      ),
      unit='short',
      color=colors.fixedColor(colors.client),
    ) + {
      id: 12,
      description: 'Number of moves per game at P50 and P95. Normal: stable distribution determined by game logic, not server performance. Watch for: sudden changes in move count (game logic bug or strategy change). Check next: Game Duration to correlate move count with total time.',
      gridPos: { h: 8, w: 8, x: 16, y: 42 },
    },
  ],
}
