// Perf Test Command Center dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/perf-test-command-center.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local thresholds = import 'thresholds.libsonnet';
local ooda = import 'ooda.libsonnet';

{
  uid: 'perf-test-command-center',
  title: 'Perf Test Command Center',
  description: 'Unified perf test dashboard: SLO status, latency attribution, saturation, leak detection, and test status',
  schemaVersion: 39,
  version: 1,
  timezone: 'browser',
  editable: true,
  graphTooltip: 1,
  tags: ['perf-test', 'command-center', 'risk-it'],
  time: { from: 'now-15m', to: 'now' },
  refresh: '5s',

  templating: {
    list: [
      {
        name: 'players_per_game',
        type: 'custom',
        label: 'Players per Game',
        current: { text: '4', value: '4' },
        options: [
          { text: '2', value: '2', selected: false },
          { text: '3', value: '3', selected: false },
          { text: '4', value: '4', selected: true },
          { text: '5', value: '5', selected: false },
          { text: '6', value: '6', selected: false },
        ],
        query: '2,3,4,5,6',
      },
    ],
  },

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
    // ── Row 100: Observe — Am I OK? ────────────────────────────────
    // SLO stat tiles (y=1) + Test Status panels (y=5)
    ooda.observeRow() + { gridPos: { h: 1, w: 24, x: 0, y: 0 } },

    // Panel 1: E2E p95
    common.statPanel(
      title='E2E p95',
      targets=[{
        refId: 'A',
        expr: 'histogram_quantile(0.95, sum(rate(perftest_e2e_duration_seconds_bucket{service_name="perftest"}[1m])) by (le))',
        legendFormat: 'E2E p95',
      }],
      thresholds=thresholds.e2eP95,
      unit='s',
      colorMode='background',
    ) + {
      id: 1,
      description: 'End-to-end move latency p95. SLO: < 500ms',
      gridPos: { h: 4, w: 6, x: 0, y: 1 },
    },

    // Panel 2: WS Delivery p95
    common.statPanel(
      title='WS Delivery p95',
      targets=[{
        refId: 'A',
        expr: 'histogram_quantile(0.95, sum(rate(perftest_ws_delivery_duration_seconds_bucket{service_name="perftest"}[1m])) by (le))',
        legendFormat: 'WS Delivery p95',
      }],
      thresholds=thresholds.wsDeliveryP95,
      unit='s',
      colorMode='background',
    ) + {
      id: 2,
      description: 'WebSocket delivery latency p95. SLO: < 200ms',
      gridPos: { h: 4, w: 6, x: 6, y: 1 },
    },

    // Panel 3: DB Txn p95
    common.statPanel(
      title='DB Txn p95',
      targets=[{
        refId: 'A',
        expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
        legendFormat: 'DB Txn p95',
      }],
      thresholds=thresholds.dbTxnP95,
      unit='s',
      colorMode='background',
    ) + {
      id: 3,
      description: 'Database transaction latency p95. SLO: < 50ms',
      gridPos: { h: 4, w: 6, x: 12, y: 1 },
    },

    // Panel 4: HTTP Error Rate
    common.statPanel(
      title='HTTP Error Rate',
      targets=[{
        refId: 'A',
        expr: '(sum(rate(http_server_requests_total{service_name="risk-it",http_status_code=~"5.."}[1m])) or vector(0)) / sum(rate(http_server_requests_total{service_name="risk-it"}[1m]))',
        legendFormat: 'Error Rate',
      }],
      thresholds=thresholds.httpErrorRate,
      unit='percentunit',
      colorMode='background',
    ) + {
      id: 4,
      description: '5xx error rate. SLO: < 1%',
      gridPos: { h: 4, w: 6, x: 18, y: 1 },
    },

    // Panel 5: Active Games (simple stat, no background, palette-classic)
    common.statPanel(
      title='Active Games',
      targets=[{
        refId: 'A',
        expr: 'perftest_games_active{service_name="perftest"}',
        legendFormat: 'Active',
      }],
      thresholds={ mode: 'absolute', steps: [{ color: 'green', value: null }] },
    ) + {
      id: 5,
      gridPos: { h: 8, w: 6, x: 0, y: 5 },
      // Override color to palette-classic (not thresholds-driven)
      fieldConfig+: {
        defaults+: {
          color: { mode: 'palette-classic' },
        },
      },
    },

    // Panel 6: Throughput
    common.timeseriesPanel(
      title='Throughput',
      targets=[{
        refId: 'A',
        expr: 'rate(perftest_moves_total{service_name="perftest"}[30s])',
        legendFormat: 'moves/s',
      }],
      unit='ops',
    ) + {
      id: 6,
      gridPos: { h: 8, w: 6, x: 6, y: 5 },
      options+: {
        tooltip: { mode: 'single' },
      },
    },

    // Panel 7: Completion Rate (inverted thresholds)
    common.statPanel(
      title='Completion Rate',
      targets=[{
        refId: 'A',
        expr: 'sum(perftest_games_completed_total{service_name="perftest"}) / (sum(perftest_games_completed_total{service_name="perftest"}) + (sum(perftest_games_timed_out_total{service_name="perftest"}) or vector(0)) + (sum(perftest_games_fatal_total{service_name="perftest"}) or vector(0)))',
        legendFormat: 'Completion',
      }],
      thresholds=thresholds.completionRate,
      unit='percentunit',
      colorMode='background',
    ) + {
      id: 7,
      gridPos: { h: 8, w: 6, x: 12, y: 5 },
    },

    // Panel 8: Error Breakdown (stacked timeseries)
    common.timeseriesPanel(
      title='Error Breakdown',
      targets=[{
        refId: 'A',
        expr: 'sum(rate(perftest_errors_total{service_name="perftest"}[1m])) by (type)',
        legendFormat: '{{type}}',
      }],
      unit='ops',
    ) + {
      id: 8,
      gridPos: { h: 8, w: 6, x: 18, y: 5 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 30,
            lineWidth: 1,
            stacking: { group: 'A', mode: 'normal' },
          },
        },
      },
      options+: {
        legend+: { calcs: ['sum'] },
      },
    },

    // ── Row 200: Orient — What's the shape? ────────────────────────
    // Hero row: E2E Percentile Bands + Latency Attribution p95
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 13 } },

    // Panel 9: E2E Latency Percentile Bands
    common.percentileBandsPanel(
      title='E2E Latency Percentile Bands',
      metric='perftest_e2e_duration_seconds_bucket',
      unit='s',
      serviceName='perftest',
    ) + {
      id: 9,
      description: 'E2E move latency distribution with p50/p95/p99 filled bands.',
      gridPos: { h: 8, w: 24, x: 0, y: 14 },
    },

    // Panel 10: Latency Attribution (p95)
    common.timeseriesPanel(
      title='Latency Attribution (p95)',
      targets=[
        {
          refId: 'A',
          expr: 'histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'HTTP Total',
        },
        {
          refId: 'B',
          expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'DB Transaction',
        },
        {
          refId: 'C',
          expr: 'histogram_quantile(0.95, sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'Game Logic',
        },
        {
          refId: 'D',
          expr: 'histogram_quantile(0.95, sum(rate(ws_broadcast_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'WS Broadcast',
        },
      ],
      unit='s',
      overrides=[
        {
          matcher: { id: 'byName', options: 'HTTP Total' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.http } },
            { id: 'custom.lineWidth', value: 3 },
          ],
        },
        {
          matcher: { id: 'byName', options: 'DB Transaction' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.db } },
          ],
        },
        {
          matcher: { id: 'byName', options: 'Game Logic' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.gameLogic } },
          ],
        },
        {
          matcher: { id: 'byName', options: 'WS Broadcast' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.ws } },
          ],
        },
      ],
    ) + {
      id: 10,
      description: 'Per-boundary p95 latency overlaid on HTTP total. Shows where time is spent.',
      gridPos: { h: 8, w: 24, x: 0, y: 22 },
    },

    // ── Row 300: Decide — Where's the bottleneck? ──────────────────
    // Boundary Saturation tiles
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 30 } },

    // Panel 11: DB Pool Utilization %
    common.statPanel(
      title='DB Pool Utilization %',
      targets=[{
        refId: 'A',
        expr: 'db_pool_active{service_name="risk-it"} / db_pool_total{service_name="risk-it"} * 100',
        legendFormat: 'Pool %',
      }],
      thresholds=thresholds.ccPoolUtil,
      unit='percent',
      colorMode='background',
    ) + {
      id: 11,
      gridPos: { h: 6, w: 4, x: 0, y: 31 },
    },

    // Panel 12: DB Pool Wait Rate
    common.statPanel(
      title='DB Pool Wait Rate',
      targets=[{
        refId: 'A',
        expr: 'rate(db_pool_empty_acquires_total{service_name="risk-it"}[1m])',
        legendFormat: 'waits/s',
      }],
      thresholds=thresholds.ccPoolWaitRate,
      unit='ops',
      colorMode='background',
    ) + {
      id: 12,
      gridPos: { h: 6, w: 4, x: 4, y: 31 },
    },

    // Panel 13: WS Active Connections
    common.statPanel(
      title='WS Active Connections',
      targets=[{
        refId: 'A',
        expr: 'ws_connections_active{service_name="risk-it"}',
        legendFormat: 'connections',
      }],
      thresholds=thresholds.ccWsConnections,
      unit='short',
      colorMode='background',
    ) + {
      id: 13,
      gridPos: { h: 6, w: 4, x: 8, y: 31 },
    },

    // Panel 14: WS Broadcast Rate
    common.statPanel(
      title='WS Broadcast Rate',
      targets=[{
        refId: 'A',
        expr: 'rate(ws_messages_sent_total{service_name="risk-it"}[1m])',
        legendFormat: 'msg/s',
      }],
      thresholds=thresholds.ccWsBroadcastRate,
      unit='ops',
      colorMode='background',
    ) + {
      id: 14,
      gridPos: { h: 6, w: 4, x: 12, y: 31 },
    },

    // Panel 15: Game Moves/s
    common.statPanel(
      title='Game Moves/s',
      targets=[{
        refId: 'A',
        expr: 'sum(rate(game_moves_total{service_name="risk-it"}[1m]))',
        legendFormat: 'moves/s',
      }],
      thresholds=thresholds.ccGameMoves,
      unit='ops',
      colorMode='background',
    ) + {
      id: 15,
      gridPos: { h: 6, w: 4, x: 16, y: 31 },
    },

    // Panel 16: HTTP Req/s
    common.statPanel(
      title='HTTP Req/s',
      targets=[{
        refId: 'A',
        expr: 'sum(rate(http_server_requests_total{service_name="risk-it"}[1m]))',
        legendFormat: 'req/s',
      }],
      thresholds=thresholds.ccHttpReqs,
      unit='reqps',
      colorMode='background',
    ) + {
      id: 16,
      gridPos: { h: 6, w: 4, x: 20, y: 31 },
    },

    // ── Row 400: Act — What's the evidence? ────────────────────────
    // Resource Leak Detection panels
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 37 } },

    // Panel 17: Goroutine Count
    common.timeseriesPanel(
      title='Goroutine Count',
      targets=[{
        refId: 'A',
        expr: 'go_goroutine_count{service_name="risk-it"}',
        legendFormat: 'goroutines',
      }],
      unit='short',
    ) + {
      id: 17,
      description: 'Monotonic increase suggests a goroutine leak',
      gridPos: { h: 8, w: 8, x: 0, y: 38 },
      options+: {
        legend+: { calcs: ['min', 'max', 'last'] },
        tooltip: { mode: 'single' },
      },
    },

    // Panel 18: WS Connection Drift
    common.timeseriesPanel(
      title='WS Connection Drift',
      targets=[
        {
          refId: 'A',
          expr: 'ws_connections_active{service_name="risk-it"}',
          legendFormat: 'Actual WS',
        },
        {
          refId: 'B',
          expr: 'perftest_games_active{service_name="perftest"} * $players_per_game',
          legendFormat: 'Expected WS',
        },
      ],
      unit='short',
    ) + {
      id: 18,
      description: 'Active WS connections vs expected (active games \u00d7 players). Divergence suggests a connection leak.',
      gridPos: { h: 8, w: 8, x: 8, y: 38 },
      options+: {
        legend+: { calcs: ['last'] },
      },
    },

    // Panel 19: DB Pool Health
    common.timeseriesPanel(
      title='DB Pool Health',
      targets=[
        {
          refId: 'A',
          expr: 'db_pool_active{service_name="risk-it"}',
          legendFormat: 'Active',
        },
        {
          refId: 'B',
          expr: 'db_pool_idle{service_name="risk-it"}',
          legendFormat: 'Idle',
        },
        {
          refId: 'C',
          expr: 'db_pool_total{service_name="risk-it"}',
          legendFormat: 'Total',
        },
      ],
      unit='short',
    ) + {
      id: 19,
      description: 'Active connections trending up without idle recovery suggests a connection leak.',
      gridPos: { h: 8, w: 8, x: 16, y: 38 },
      options+: {
        legend+: { calcs: ['min', 'max', 'last'] },
      },
    },
  ],
}
