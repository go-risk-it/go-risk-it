// Perf Test Command Center dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/perf-test-command-center.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local links = import 'links.libsonnet';
local thresholds = import 'thresholds.libsonnet';
local ooda = import 'ooda.libsonnet';

dashboard.new(
  uid='perf-test-command-center',
  title='Perf Test Command Center',
  description='Unified perf test dashboard: SLO status, latency attribution, saturation, leak detection, and test status',
  tags=['perf-test', 'command-center', 'risk-it'],
  refresh='5s',
  graphTooltip=1,
  templating={
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
  annotations={
    list: [
      dashboard.perfTestAnnotation,
    ],
  },
  panels=[
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
      description: 'End-to-end move latency P95 measured by the client. SLO: < 500ms. Check next: Perf Test dashboard E2E Move Latency panel for percentile trend over time.',
      gridPos: { h: 4, w: 6, x: 0, y: 1 },
      options+: {
        links: [links.toDashboard('Perf Test Detail', links.dashboardUids.perfTest)],
      },
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
      description: 'WebSocket state delivery latency P95. SLO: < 200ms. Check next: WebSocket dashboard for connection lifecycle and broadcast latency breakdown.',
      gridPos: { h: 4, w: 6, x: 6, y: 1 },
      options+: {
        links: [links.toDashboard('WebSocket Detail', links.dashboardUids.websocket)],
      },
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
      description: 'Database transaction latency P95. SLO: < 50ms. Check next: Database dashboard for pool utilization, query latency, and cache hit rate.',
      gridPos: { h: 4, w: 6, x: 12, y: 1 },
      options+: {
        links: [links.toDashboard('Database Detail', links.dashboardUids.database)],
      },
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
      description: 'Server-side 5xx error rate as percentage of total requests. SLO: < 1%. Check next: Error Breakdown panel for error type categorization.',
      gridPos: { h: 4, w: 6, x: 18, y: 1 },
      options+: {
        links: [links.toDashboard('Server Golden Signals', links.dashboardUids.serverGoldenSignals)],
      },
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
      description: 'Number of games currently in progress. Normal: matches configured concurrency target. Watch for: stuck at 0 or exceeding target (games not draining). Check next: Completion Rate to see if games are finishing successfully.',
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
      description: 'Client-side move throughput (30s rate). Normal: proportional to active games. Watch for: sudden drops (server overload) or flat at zero (test harness stuck). Check next: Perf Test dashboard Moves/sec for detailed throughput trend.',
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
      description: 'Ratio of completed games to total outcomes. SLO: > 95% green, > 80% yellow. Normal: > 95%. Watch for: dropping below 80% (too many timeouts or fatals). Check next: Error Breakdown for error types causing failures.',
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
      description: 'Client-side errors stacked by type (connection, timeout, HTTP 5xx, etc.). Normal: zero or near-zero. Watch for: any sustained error rate or new error types appearing. Check next: Server Golden Signals dashboard for server-side error details.',
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
      description: 'E2E move latency with filled bands between P50, P95, and P99. Normal: tight bands with P95 < 500ms. Watch for: bands widening (latency variance increasing) or P99 diverging from P95 (tail latency problem). Check next: Latency Attribution to identify which server boundary causes the spread.',
      gridPos: { h: 8, w: 24, x: 0, y: 14 },
    },

    // Panel 10: Latency Attribution (p95)
    common.timeseriesPanel(
      title='Latency Attribution (p95)',
      targets=common.lifecycleTargets,
      unit='s',
      overrides=common.lifecycleOverrides,
    ) + {
      id: 10,
      description: 'P95 latency for each server boundary overlaid: HTTP total, DB transaction, game logic, WS broadcast, event handler (dashed, async post-response). Normal: DB + game logic + WS sum to roughly HTTP total; event handler runs independently after response. Watch for: one boundary dominating (e.g. DB > 70% of HTTP) or event handler exceeding HTTP total. Check next: Database dashboard if DB dominates, WebSocket dashboard if WS dominates, Request Lifecycle for event handler breakdown.',
      gridPos: { h: 8, w: 24, x: 0, y: 22 },
      fieldConfig+: {
        defaults+: {
          links: [
            links.toDashboard('Database Detail', links.dashboardUids.database),
            links.toDashboard('WebSocket Detail', links.dashboardUids.websocket),
            links.toDashboard('Request Lifecycle', links.dashboardUids.requestLifecycle),
          ],
        },
      },
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
      description: 'Percentage of DB connection pool in use. Normal: < 70%. Watch for: > 90% (pool exhaustion imminent, queries will queue). Check next: DB Pool Wait Rate tile — non-zero waits confirm saturation.',
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
      description: 'Rate of connection acquires that had to wait for a free connection. Normal: 0. Watch for: > 10/s (pool too small for concurrency level). Check next: DB Pool Utilization tile to confirm pool is near capacity.',
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
      description: 'Total active WebSocket connections on the server. Normal: active_games x players_per_game. Watch for: > 5000 (approaching connection limits). Check next: WS Connection Drift panel to detect connection leaks.',
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
      description: 'Rate of WebSocket messages broadcast by the server. Normal: proportional to moves/s x players_per_game. Watch for: > 50K/s (fan-out amplification or broadcast storms). Check next: Fan-out Amplification panel for the WS-to-move ratio.',
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
      description: 'Server-side game move processing rate. Normal: proportional to active games. Watch for: > 5K/s (approaching high-throughput territory). Check next: Game Engine dashboard for per-phase move breakdown.',
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
      description: 'Total HTTP request rate hitting the server. Normal: higher than moves/s due to game creation, WS upgrades, and state queries. Watch for: > 5K/s (load test pushing high throughput). Check next: Server Golden Signals dashboard for latency and error rate at this request volume.',
      gridPos: { h: 6, w: 4, x: 20, y: 31 },
    },

    // Panel 20: Fan-out Amplification
    common.timeseriesPanel(
      title='Fan-out Amplification',
      targets=[{
        refId: 'A',
        expr: 'rate(ws_messages_sent_total{service_name="risk-it"}[1m]) / rate(game_moves_total{service_name="risk-it"}[1m])',
        legendFormat: 'WS msgs / move',
      }],
      unit='short',
      color=colors.fixedColor(colors.ws),
    ) + {
      id: 20,
      description: |||
        Ratio of WS broadcasts to game moves. Each move triggers a state broadcast to all players in the game, so the baseline ratio equals players_per_game (typically 4).
        Normal: ~4 (one broadcast per player per move). Watch for: > 10 suggests duplicate broadcasts or fan-out bugs. Check next: WebSocket dashboard for broadcast latency and connection counts.
        Denominator: game_moves_total counts server-side move completions (one per successful move request), not client retries.
      |||,
      gridPos: { h: 8, w: 8, x: 0, y: 37 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('WebSocket Detail', links.dashboardUids.websocket)],
        },
      },
    },

    // Panel 21: DB Latency Share
    common.timeseriesPanel(
      title='DB Latency Share',
      targets=[{
        refId: 'A',
        expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)) / histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)) * 100',
        legendFormat: 'DB share %',
      }],
      unit='percent',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 21,
      description: |||
        Percentage of HTTP p95 latency spent in DB transactions. Derived from: DB txn p95 / HTTP request p95 x 100.
        Normal: 30-50%. Watch for: > 70% means DB dominates request latency — look at pool saturation and query plans. Check next: Database dashboard for pool utilization and per-query latency.
      |||,
      gridPos: { h: 8, w: 8, x: 8, y: 37 },
    },

    // Panel 22: Effective Concurrency
    common.timeseriesPanel(
      title='Effective Concurrency',
      targets=[
        {
          refId: 'A',
          expr: 'perftest_health_healthy{service_name="perftest"} + perftest_health_slow{service_name="perftest"}',
          legendFormat: 'Effective (healthy+slow)',
        },
        {
          refId: 'B',
          expr: 'perftest_games_active{service_name="perftest"}',
          legendFormat: 'Active (total)',
        },
      ],
      unit='short',
      overrides=[
        {
          matcher: { id: 'byName', options: 'Effective (healthy+slow)' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.gameLogic) },
          ],
        },
        {
          matcher: { id: 'byName', options: 'Active (total)' },
          properties: [
            { id: 'color', value: colors.fixedColor(colors.client) },
            { id: 'custom.lineStyle', value: { fill: 'dash', dash: [10, 10] } },
            { id: 'custom.fillOpacity', value: 0 },
          ],
        },
      ],
    ) + {
      id: 22,
      description: |||
        Effective concurrency (healthy + slow games) vs total active games. The gap represents stalled/zombie games not making progress.
        Normal: effective ≈ active. Watch for: growing gap means games are getting stuck — check DB pool and goroutine panels. Check next: DB Pool Health and Goroutine Count panels for resource leak evidence.
      |||,
      gridPos: { h: 8, w: 8, x: 16, y: 37 },
    },

    // ── Row 400: Act — What's the evidence? ────────────────────────
    // Resource Leak Detection panels
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 45 } },

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
      description: 'Server goroutine count over time. Normal: stable, proportional to active connections. Watch for: monotonic increase (goroutine leak — each game/connection spawns goroutines that never exit). Check next: WS Connection Drift to see if leaked goroutines correlate with leaked connections.',
      gridPos: { h: 8, w: 8, x: 0, y: 46 },
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
      description: 'Actual WS connections vs expected (active_games x players_per_game). Normal: lines track closely. Watch for: actual exceeding expected (connection leak — connections not closing when games end). Check next: DB Pool Health to check for correlated connection pool leaks.',
      gridPos: { h: 8, w: 8, x: 8, y: 46 },
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
      description: 'DB connection pool breakdown: active, idle, and total connections. Normal: active rises under load then returns to idle. Watch for: active trending up without idle recovery (connection leak) or total hitting pool max. Check next: Database dashboard for pool acquire latency and canceled acquires.',
      gridPos: { h: 8, w: 8, x: 16, y: 46 },
      options+: {
        legend+: { calcs: ['min', 'max', 'last'] },
      },
    },
  ],
)
