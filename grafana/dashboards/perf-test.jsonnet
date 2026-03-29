// Perf Test dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/perf-test.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local layout = import 'layout.libsonnet';
local links = import 'links.libsonnet';
local modifiers = import 'modifiers.libsonnet';
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

// Shared data links for cross-dashboard navigation.
local crossLinksSystemHealth = links.toDashboard('System Health', links.dashboardUids.systemHealth);
local crossLinksGameEngine = links.toDashboard('Game Engine', links.dashboardUids.gameEngine);

dashboard.new(
  uid='perf-test',
  title='Perf Test',
  description='Load test client metrics + SLO compliance',
  tags=['perf-test', 'risk-it', 'load-test'],
  refresh='5s',
  graphTooltip=1,
  annotations={ list: [dashboard.perfTestAnnotation] },
  panels=layout.ooda(

    // ================================================================
    // OBSERVE — Am I OK? (8 panels)
    // ================================================================
    observe=[
      // Row 1: 4 SLO tiles (w=6 each)

      // E2E p95 (stat bg)
      layout.panel(
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
          fieldConfig+: {
            defaults+: {
              links: [crossLinksSystemHealth, crossLinksGameEngine],
            },
          },
        },
        w=6, h=4,
        description='Normal: < 500ms (green). Watch for: crossing 500ms (yellow) or 1s (red) sustained. Check next: E2E Move Latency panel for percentile trend over time.',
      ),

      // WS Delivery p95 (stat bg)
      layout.panel(
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
          fieldConfig+: {
            defaults+: {
              links: [crossLinksSystemHealth],
            },
          },
        },
        w=6, h=4,
        description='Normal: < 200ms (green). Watch for: crossing 200ms (yellow) or 500ms (red) sustained. Check next: WS Delivery Latency panel for percentile breakdown.',
      ),

      // DB Txn p95 (stat bg)
      layout.panel(
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
          fieldConfig+: {
            defaults+: {
              links: [crossLinksSystemHealth],
            },
          },
        },
        w=6, h=4,
        description='Normal: < 50ms (green). Watch for: crossing 50ms (yellow) or 100ms (red) sustained. Check next: System Health DB Pool panels for saturation evidence.',
      ),

      // HTTP Error Rate (stat bg)
      layout.panel(
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
          fieldConfig+: {
            defaults+: {
              links: [crossLinksSystemHealth],
            },
          },
        },
        w=6, h=4,
        description='Normal: < 1% (green). Watch for: crossing 1% (yellow) or 5% (red) sustained. Check next: Error Breakdown panel for error type categorization.',
      ),

      // Row 2: Active Games + Throughput + Completion Rate + Error Breakdown

      // Active Games (stat, no background)
      layout.panel(
        common.statPanel(
          title='Active Games',
          targets=[{
            refId: 'A',
            expr: 'perftest_games_active{service_name="perftest"}',
            legendFormat: 'active',
          }],
          thresholds=thresholds.perfTestActiveGames,
        ),
        w=6, h=6,
        description='Normal: matches configured concurrency. Watch for: stuck at 0 (no games starting) or exceeding target (games not completing). Check next: Completion Rate and Game Completion panels.',
      ),

      // Throughput (timeseries)
      layout.panel(
        common.timeseriesPanel(
          title='Throughput',
          targets=[{
            refId: 'A',
            expr: 'rate(perftest_moves_total{service_name="perftest"}[30s])',
            legendFormat: 'moves/s',
          }],
          unit='ops',
          color=colors.fixedColor(colors.client),
        ) + {
          fieldConfig+: {
            defaults+: {
              custom+: { fillOpacity: 15 },
              links: [crossLinksGameEngine],
            },
          },
        },
        w=6, h=6,
        description='Normal: proportional to active games. Watch for: sudden drops (server overload) or flat at zero (test harness stuck). Check next: E2E Move Latency to see if slowdown explains throughput drop.',
      ),

      // Completion Rate (stat bg, inverted thresholds)
      layout.panel(
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
        ),
        w=6, h=6,
        description='Normal: > 95% (green). Watch for: dropping below 80% (yellow) means too many timeouts or fatals. Check next: Error Breakdown for error types causing failures.',
      ),

      // Error Breakdown (timeseries stacked)
      layout.panel(
        common.timeseriesPanel(
          title='Error Breakdown',
          targets=[{
            refId: 'A',
            expr: 'sum(rate(perftest_errors_total{service_name="perftest"}[1m])) by (type)',
            legendFormat: '{{type}}',
          }],
          unit='ops',
        ) + {
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
        w=6, h=6,
        description='Normal: zero or near-zero. Watch for: any sustained error rate or new error types appearing. Check next: System Health dashboard for server-side error details.',
      ),
    ],

    // ================================================================
    // ORIENT — What's the shape? (3 hero panels)
    // ================================================================
    orient=[
      // E2E Latency Heatmap (w=24, h=8, YlOrRd)
      layout.panel(
        common.heatmapPanel(
          title='E2E Latency Heatmap',
          targets=[{
            expr: 'sum(rate(perftest_e2e_duration_seconds_bucket{service_name="perftest"}[$__rate_interval])) by (le)',
            format: 'heatmap',
            legendFormat: '{{le}}',
            refId: 'A',
          }],
          unit='s',
          colorScheme='YlOrRd',
          colorFill='dark-red',
        ),
        w=24, h=8,
        description='Normal: dense band below 500ms. Watch for: color spreading into higher buckets (latency distribution widening under load). Check next: Percentile Bands below for exact P50/P95/P99 values.',
      ),

      // E2E Percentile Bands (w=24, h=8)
      layout.panel(
        common.percentileBandsPanel(
          title='E2E Latency Percentile Bands',
          metric='perftest_e2e_duration_seconds_bucket',
          unit='s',
          serviceName='perftest',
        ),
        w=24, h=8,
        description='Normal: tight bands with P95 < 500ms. Watch for: bands widening (latency variance increasing) or P99 diverging from P95 (tail latency problem). Check next: Latency Attribution to identify which server boundary causes the spread.',
      ),

      // Latency Attribution (w=24, h=8)
      layout.panel(
        common.timeseriesPanel(
          title='Latency Attribution (p95)',
          targets=common.lifecycleTargets,
          unit='s',
          overrides=common.lifecycleOverrides,
        ) + {
          fieldConfig+: {
            defaults+: {
              links: [
                links.toDashboard('System Health', links.dashboardUids.systemHealth),
                links.toDashboard('System Health', links.dashboardUids.systemHealth),
                links.toDashboard('System Health', links.dashboardUids.systemHealth),
              ],
            },
          },
        },
        w=24, h=8,
        description='Normal: DB + game logic + WS sum to roughly HTTP total; event handler runs independently after response. Watch for: one boundary dominating (e.g. DB > 70% of HTTP) or event handler exceeding HTTP total. Check next: Database dashboard if DB dominates, WebSocket dashboard if WS dominates.',
      ),
    ],

    // ================================================================
    // DECIDE — Where's the bottleneck? (7 always-visible + 2 collapsed)
    // ================================================================
    decide=[
      // E2E Move Latency P50/P95/P99 (w=12, h=8, SLO threshold overlay)
      layout.panel(
        common.timeseriesPanel(
          title='E2E Move Latency (P50/P95/P99)',
          targets=common.histogramQuantileTargetsWithExemplars(
            'perftest_e2e_duration_seconds_bucket',
            [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
            'perftest',
          ),
          unit='s',
          color=colors.fixedColor(colors.client),
        ) + common.withSloThreshold(thresholds.e2eP95),
        w=12, h=8,
        description='Normal: P95 < 500ms, P99 < 1s. Watch for: P95 crossing the threshold line (SLO breach). Check next: REST Latency by Action to identify which move type is slowest.',
      ),

      // WS Delivery Latency P50/P95/P99 (w=12, h=8, SLO threshold overlay)
      layout.panel(
        common.timeseriesPanel(
          title='WS Delivery Latency (P50/P95/P99)',
          targets=common.histogramQuantileTargetsWithExemplars(
            'perftest_ws_delivery_duration_seconds_bucket',
            [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
            'perftest',
          ),
          unit='s',
          color=colors.fixedColor(colors.ws),
        ) + common.withSloThreshold(thresholds.wsDeliveryP95) + {
          fieldConfig+: {
            defaults+: {
              links: [links.toDashboard('System Health', links.dashboardUids.systemHealth)],
            },
          },
        },
        w=12, h=8,
        description='Normal: P95 < 200ms. Watch for: P95 crossing the threshold line (WS delivery SLO breach). Check next: WebSocket dashboard for connection and broadcast details.',
      ),

      // REST Latency by Action (w=12, h=8)
      layout.panel(
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
        ),
        w=12, h=8,
        description='Normal: all actions < 200ms at P95. Watch for: single action diverging (e.g. attack P95 spiking while others stay flat). Check next: Database dashboard for query-level latency.',
      ),

      // Conflicts/sec (w=12, h=8)
      layout.panel(
        common.timeseriesPanel(
          title='Conflicts/sec',
          targets=[{
            expr: 'rate(perftest_conflicts_total{service_name="perftest"}[30s])',
            legendFormat: 'conflicts/s',
            refId: 'A',
          }],
          unit='ops',
          color=colors.fixedColor(colors.http),
        ) + {
          fieldConfig+: {
            defaults+: {
              custom+: { fillOpacity: 15 },
            },
          },
        },
        w=12, h=8,
        description='Normal: low, proportional to concurrency. Watch for: conflicts/s exceeding moves/s (more retries than successes, contention too high). Check next: E2E Latency vs Concurrency to correlate conflict rate with load.',
      ),

      // E2E Latency vs Concurrency (w=12, h=8, dual-axis)
      layout.panel(
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
        ),
        w=12, h=8,
        description='Normal: latency stays flat as concurrency increases. Watch for: inflection point where latency climbs sharply with concurrency (saturation). Check next: Server Latency vs Concurrency to see if server-side shows the same knee.',
      ),

      // Server Latency vs Concurrency (w=12, h=8, dual-axis)
      layout.panel(
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
          fieldConfig+: {
            defaults+: {
              links: [
                links.toDashboard('System Health', links.dashboardUids.systemHealth),
                crossLinksSystemHealth,
              ],
            },
          },
        },
        w=12, h=8,
        description='Normal: HTTP P95 < 100ms regardless of concurrency. Watch for: server latency rising before client E2E does (server is the bottleneck). Check next: Latency Attribution in Orient to identify which server boundary dominates.',
      ),

      // Effective Concurrency (w=12, h=8)
      layout.panel(
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
        ),
        w=12, h=8,
        description='Normal: effective approximates active. Watch for: growing gap means games are getting stuck (stalled/zombie). Check next: Health Breakdown collapsed row for per-state waterfall.',
      ),
    ],

    decideDepth={
      // ── Collapsed: Client Phase Latency (1 panel) ──
      'Client Phase Latency': [
        layout.panel(
          common.timeseriesPanel(
            title='Phase Duration by Phase',
            targets=[{
              expr: 'histogram_quantile(0.95, sum(rate(perftest_phase_duration_seconds_bucket{service_name="perftest"}[1m])) by (le, phase))',
              legendFormat: '{{phase}} p95',
              refId: 'A',
            }],
            unit='s',
          ),
          w=24, h=8,
          description='Normal: deploy and attack phases are longest. Watch for: single phase P95 spiking while others stay flat. Check next: REST Latency by Action for per-action HTTP latency comparison.',
        ),
      ],

      // ── Collapsed: Health Breakdown (1 showcase panel) ──
      'Health Breakdown': [
        // Health State Waterfall SHOWCASE
        layout.panel(
          common.timeseriesPanel(
            title='Health State Waterfall',
            targets=[
              {
                expr: 'perftest_health_healthy{service_name="perftest"}',
                legendFormat: 'healthy',
                refId: 'A',
              },
              {
                expr: 'perftest_health_slow{service_name="perftest"}',
                legendFormat: 'slow',
                refId: 'B',
              },
              {
                expr: 'perftest_health_stalled{service_name="perftest"}',
                legendFormat: 'stalled',
                refId: 'C',
              },
              {
                expr: 'perftest_health_zombie{service_name="perftest"}',
                legendFormat: 'zombie',
                refId: 'D',
              },
            ],
            unit='short',
          )
          + modifiers.withStackedArea(40, 'scheme')
          + modifiers.withSeriesColors({
            healthy: '#56A64B',
            slow: '#FADE2A',
            stalled: '#FF9830',
            zombie: '#E02F44',
          }) + {
            options+: {
              legend+: {
                displayMode: 'table',
                placement: 'bottom',
                calcs: ['mean', 'last'],
              },
            },
          },
          w=24, h=10,
          description='Normal: healthy dominates the stack, others near zero. Watch for: slow/stalled/zombie layers growing (games not making progress). Check next: Effective Concurrency panel for the gap metric, System Health for DB pool saturation.',
        ),
      ],
    },

    // ================================================================
    // ACT — What's the evidence? (4 panels)
    // ================================================================
    act=[
      // Error Rate by Type (timeseries stacked, w=8)
      layout.panel(
        common.timeseriesPanel(
          title='Error Rate by Type',
          targets=[{
            expr: 'sum(rate(perftest_errors_total{service_name="perftest"}[1m])) by (type)',
            legendFormat: '{{type}}',
            refId: 'A',
          }],
          unit='ops',
        ) + {
          fieldConfig+: {
            defaults+: {
              color: { mode: 'palette-classic' },
              custom+: {
                stacking: { group: 'A', mode: 'normal' },
              },
            },
          },
        },
        w=8, h=8,
        description='Normal: zero or near-zero. Watch for: any sustained error rate, especially new error types appearing. Check next: Game Completion panel to see if errors cause game failures.',
      ),

      // Game Duration P50/P95 (timeseries, w=8)
      layout.panel(
        common.timeseriesPanel(
          title='Game Duration (P50/P95)',
          targets=common.histogramQuantileTargetsWithExemplars(
            'perftest_game_duration_seconds_bucket',
            [['0.5', 'p50'], ['0.95', 'p95']],
            'perftest',
          ),
          unit='s',
          color=colors.fixedColor(colors.gameLogic),
        ),
        w=8, h=8,
        description='Normal: consistent across the test run. Watch for: game duration growing over time (server slowing down under sustained load). Check next: Moves per Game to check if duration increase is from more moves or slower moves.',
      ),

      // Moves per Game P50/P95 (timeseries, w=8)
      layout.panel(
        common.timeseriesPanel(
          title='Moves per Game (P50/P95)',
          targets=common.histogramQuantileTargetsWithExemplars(
            'perftest_game_moves_bucket',
            [['0.5', 'p50'], ['0.95', 'p95']],
            'perftest',
          ),
          unit='short',
          color=colors.fixedColor(colors.client),
        ),
        w=8, h=8,
        description='Normal: stable distribution determined by game logic, not server performance. Watch for: sudden changes in move count (game logic bug or strategy change). Check next: Game Duration to correlate move count with total time.',
      ),

      // Game Completion cumulative (timeseries, w=24)
      layout.panel(
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
          options+: {
            legend+: { calcs: ['lastNotNull'] },
          },
        },
        w=24, h=8,
        description='Normal: completed grows steadily, others stay flat. Watch for: timed-out or fatal counts climbing (server cannot finish games in time). Check next: Error Rate by Type for error categorization.',
      ),
    ],
  ),
)
