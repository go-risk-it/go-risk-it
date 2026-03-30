// Game Engine dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/game-engine.jsonnet
// Regenerate: make dashboards
local colors = import 'colors.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local layout = import 'layout.libsonnet';
local links = import 'links.libsonnet';
local modifiers = import 'modifiers.libsonnet';
local panels = import 'panels.libsonnet';
local targets = import 'targets.libsonnet';
local thresholds = import 'thresholds.libsonnet';

local svc = targets.serviceName;

// Template variable for game-scoped log filtering.
local gameIdVar = {
  name: 'gameId',
  type: 'textbox',
  label: 'Game ID',
  current: { value: '' },
};

// Shared data links to cross-dashboard navigation.
local crossLinks = [
  links.toDashboard('System Health', links.dashboardUids.systemHealth),
  links.toDashboard('Perf Test', links.dashboardUids.perfTest),
];

dashboard.new(
  uid='game-engine',
  title='Game Engine',
  description='Game engine observability: game lifecycle, phase timing, event bus health, and move analytics',
  tags=['risk-it', 'game-engine'],
  templating={ list: [gameIdVar] },
  panels=layout.ooda(

    // ════════════════════════════════════════════════════════════════
    // OBSERVE — Am I OK? (6 panels: 4 stats + 2 timeseries)
    // ════════════════════════════════════════════════════════════════
    observe=[
      // ── Row 1: Stat summary bar (w=6 each) ──

      // Active Games (stat)
      layout.panel(
        panels.statPanel(
          title='Active Games',
          targets=[targets.target('game_active{service_name="%s"}' % svc, 'Active Games')],
          thresholds=thresholds.activeGames,
        ) + modifiers.withLinks([links.toDashboard('Perf Test', links.dashboardUids.perfTest)]),
        w=6, h=8,
        description='Normal: < 50 active games (green). Watch for: sustained > 100 (red) may indicate stuck phases or games not finishing. Check next: Games Created/Finished Rate for flow balance.',
      ),

      // Total Events (stat)
      layout.panel(
        panels.statPanel(
          title='Total Events',
          targets=[targets.target(
            'sum(event_bus_events_total{service_name="%s"})' % svc,
            'total events',
          )],
          thresholds={
            mode: 'absolute',
            steps: [{ color: 'green', value: null }],
          },
        ),
        w=6, h=8,
        description='Normal: growing counter proportional to game activity. Watch for: counter stalling (bus stopped processing). Check next: Event Heartbeat for rate view.',
      ),

      // Handler Error Rate (stat — surfaced error indicator)
      layout.panel(
        panels.statPanel(
          title='Handler Error Rate',
          targets=[targets.target(
            'sum(rate(event_handler_errors_total{service_name="%s"}[5m]))' % svc,
            'errors/s',
          )],
          thresholds=thresholds.handlerErrorRate,
          unit='ops',
        ),
        w=6, h=8,
        description='Normal: 0 errors (green). Watch for: any non-zero value (red) indicates handler failures. Check next: Event Bus Detail collapsed row for per-handler error breakdown.',
      ),

      // Game Duration P50 (stat — median game length)
      layout.panel(
        panels.statPanel(
          title='Game Duration P50',
          targets=[targets.target(
            'histogram_quantile(0.5, sum(rate(game_duration_seconds_bucket{service_name="%s"}[5m])) by (le))' % svc,
            'median',
          )],
          thresholds={
            mode: 'absolute',
            steps: [{ color: 'green', value: null }],
          },
          unit='s',
        ),
        w=6, h=8,
        description='Normal: consistent median game duration. Watch for: rising median means games are getting slower. Check next: Game Duration P50/P95 in Decide for trend analysis.',
      ),

      // ── Row 2: Timeseries (w=12 each) ──

      // Games Created/Finished Rate (timeseries)
      layout.panel(
        panels.timeseriesPanel(
          title='Games Created/Finished Rate',
          targets=[
            targets.target('rate(game_created_total{service_name="%s"}[1m])' % svc, 'Created'),
            targets.target('rate(game_finished_total{service_name="%s"}[1m])' % svc, 'Finished', 'B'),
          ],
          unit='ops',
        ) + modifiers.withSeriesColors({ Created: colors.gameLogic, Finished: colors.errors }),
        w=12, h=8,
        description='Normal: created and finished rates roughly balanced. Watch for: created >> finished means games accumulating (stuck phases). Check next: Active Games stat for current count.',
      ),

      // SHOWCASE: Game Event Heartbeat (stacked area, event type colors)
      layout.panel(
        panels.timeseriesPanel(
          title='Game Event Heartbeat',
          targets=[targets.target(
            'sum(rate(event_bus_events_total{service_name="%s"}[1m])) by (event_type)' % svc,
            '{{event_type}}',
          )],
          unit='ops',
        ) + modifiers.withStackedArea(25, 'opacity') + modifiers.withSeriesColors(colors.eventTypes),
        w=12, h=8,
        description='Normal: move_executed dominates, all 8 event types visible. Watch for: any event type dropping to zero while others continue. Check next: Event Handler Latency in Decide for handler-level bottlenecks.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // ORIENT — What's the shape? (4 always-visible, 2 rows of 2)
    // ════════════════════════════════════════════════════════════════
    orient=[
      // Phase Duration Heatmap (Oranges)
      layout.panel(
        panels.heatmapPanel(
          title='Phase Duration Heatmap',
          targets=[targets.heatmapTarget(
            'sum(rate(game_phase_duration_seconds_bucket{service_name="%s"}[1m])) by (le)' % svc,
          )],
          unit='s',
          colorScheme='Oranges',
          colorFill='dark-orange',
        ),
        w=12, h=8,
        description='Normal: dense band in low-seconds range (fast phases). Watch for: heat spreading to higher buckets over time. Check next: Phase Latency P50/P99 for per-phase breakdown.',
      ),

      // Game Duration Heatmap (Blues)
      layout.panel(
        panels.heatmapPanel(
          title='Game Duration Heatmap',
          targets=[targets.heatmapTarget(
            'sum(rate(game_duration_seconds_bucket{service_name="%s"}[1m])) by (le)' % svc,
          )],
          unit='s',
          colorScheme='Blues',
          colorFill='dark-blue',
        ),
        w=12, h=8,
        description='Normal: concentrated distribution showing consistent game lengths. Watch for: bimodal distribution (fast + very slow games). Check next: Game Duration P50/P95 in Decide for percentile trends.',
      ),

      // Phase Latency P50/P99 by phase (merged: replaces Phase Duration P50/P95 + Move Timing)
      layout.panel(
        panels.timeseriesPanel(
          title='Phase Latency P50/P99',
          targets=[
            targets.target(
              'histogram_quantile(0.5, sum(rate(game_phase_duration_seconds_bucket{service_name="%s"}[1m])) by (le, phase))' % svc,
              '{{phase}} P50',
            ),
            targets.target(
              'histogram_quantile(0.99, sum(rate(game_phase_duration_seconds_bucket{service_name="%s"}[1m])) by (le, phase))' % svc,
              '{{phase}} P99',
              'B',
            ),
          ],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
        w=12, h=8,
        description='Normal: deploy and attack phases are longest; reinforce is shortest. Watch for: P99 diverging from P50 (tail latency outliers), a single phase spiking while others stay flat. Check next: Game Duration P50/P95 in Decide.',
      ),

      // Headline Event Rates (merged: Event Cascade + Player Elimination)
      layout.panel(
        panels.timeseriesPanel(
          title='Headline Event Rates',
          targets=[
            targets.target(
              'sum(rate(event_bus_events_total{service_name="%s",event_type="continent_captured"}[1m]))' % svc,
              'Captured',
            ),
            targets.target(
              'sum(rate(event_bus_events_total{service_name="%s",event_type="continent_lost"}[1m]))' % svc,
              'Lost',
              'B',
            ),
            targets.target(
              'sum(rate(event_bus_events_total{service_name="%s",event_type="player_eliminated"}[1m]))' % svc,
              'Eliminated',
              'C',
            ),
          ],
          unit='ops',
        ) + modifiers.withSeriesColors({
          Captured: colors.http,
          Lost: '#FF780A',
          Eliminated: colors.errors,
        }),
        w=12, h=8,
        description='Normal: captured and lost rates correlated with game activity; elimination is lower steady rate. Watch for: captured >> lost (map consolidation), sudden elimination spikes (many games ending). Check next: Game Duration Heatmap for game length distribution.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // DECIDE — Where's the bottleneck? (4 always-visible + 2 collapsed rows)
    // ════════════════════════════════════════════════════════════════
    decide=[
      // Moves per Second by Phase
      layout.panel(
        panels.timeseriesPanel(
          title='Moves per Second by Phase',
          targets=[targets.target(
            'sum(rate(game_moves_total{service_name="%s"}[1m])) by (phase)' % svc,
            '{{phase}}',
          )],
          unit='ops',
        ) + modifiers.withLinks(crossLinks),
        w=12, h=8,
        description='Normal: deploy most frequent, then attack, then reinforce/conquer. Watch for: a phase dropping to zero while others continue (phase blocked). Check next: Event Handler Latency for downstream processing speed.',
      ),

      // Game Duration P50/P95
      layout.panel(
        panels.timeseriesPanel(
          title='Game Duration P50/P95',
          targets=targets.histogramQuantileTargetsWithExemplars(
            'game_duration_seconds_bucket',
            [['0.5', 'P50'], ['0.95', 'P95']],
          ),
          unit='s',
          color=colors.fixedColor(colors.gameLogic),
        ) + modifiers.withPercentileColors('gameLogic'),
        w=12, h=8,
        description='Normal: consistent P50 with P95 within 2-3x of P50. Watch for: P95 growing while P50 stays flat (subset of slow games). Check next: Phase Latency P50/P99 in Orient to identify which phase is slow.',
      ),

      // Event Handler Latency by handler (p95)
      layout.panel(
        panels.timeseriesPanel(
          title='Event Handler Latency p95',
          targets=[targets.target(
            'histogram_quantile(0.95, sum(rate(event_handler_duration_seconds_bucket{service_name="%s"}[1m])) by (le, handler))' % svc,
            '{{handler}} p95',
          )],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
        w=12, h=8,
        description='Normal: all handlers < 100ms p95. Watch for: individual handler p95 > 500ms (slow consumer bottleneck). Check next: Event Dispatch Duration for bus-level overhead.',
      ),

      // Event Dispatch Duration by event_type (p95)
      layout.panel(
        panels.timeseriesPanel(
          title='Event Dispatch Duration p95',
          targets=[targets.target(
            'histogram_quantile(0.95, sum(rate(event_bus_dispatch_duration_seconds_bucket{service_name="%s"}[1m])) by (le, event_type))' % svc,
            '{{event_type}}',
          )],
          unit='s',
        ) + modifiers.withSeriesColors(colors.eventTypes),
        w=12, h=8,
        description='Normal: all event types dispatched < 50ms. Watch for: move_executed dispatch exceeding 100ms (many handlers subscribed). Check next: Event Bus Detail collapsed row for throughput and errors.',
      ),
    ],

    decideDepth={
      // ── Collapsed: Event Bus Detail (2 panels) ──
      'Event Bus Detail': [
        // Handler Throughput by event_type
        layout.panel(
          panels.timeseriesPanel(
            title='Handler Throughput',
            targets=[targets.target(
              'sum(rate(event_bus_events_total{service_name="%s"}[1m])) by (event_type)' % svc,
              '{{event_type}}',
            )],
            unit='ops',
          ) + modifiers.withSeriesColors(colors.eventTypes),
          w=12, h=8,
          description='Normal: move_executed dominates throughput. Watch for: unexpected event type surges. Check next: Event Bus Events Total for cumulative count.',
        ),

        // Event Bus Events Total (stat)
        layout.panel(
          panels.statPanel(
            title='Event Bus Events Total',
            targets=[targets.target(
              'sum(event_bus_events_total{service_name="%s"})' % svc,
              'total events',
            )],
            thresholds={
              mode: 'absolute',
              steps: [{ color: 'green', value: null }],
            },
          ),
          w=12, h=8,
          description='Normal: growing counter proportional to game activity. Watch for: counter stalling (bus stopped processing). Check next: Handler Throughput for rate view.',
        ),
      ],

      // ── Collapsed: Move Routes (2 panels) ──
      'Move Routes': [
        // Total Moves by Phase (bar gauge)
        layout.panel(
          panels.barGaugePanel(
            title='Total Moves by Phase',
            targets=[targets.target(
              'sum(game_moves_total{service_name="%s"}) by (phase)' % svc,
              '{{phase}}',
            )],
          ),
          w=12, h=8,
          description='Normal: deploy highest, attack second, conquer/reinforce lower. Watch for: unusual ratios (few conquer relative to attack). Check next: Moves per Second by Phase for rate view.',
        ),

        // Game HTTP Route Request Rate
        layout.panel(
          panels.timeseriesPanel(
            title='Game HTTP Route Request Rate',
            targets=[targets.target(
              'sum(rate(http_server_requests_total{service_name="%s",http_route=~".*games.*"}[1m])) by (http_route)' % svc,
              '{{http_route}}',
            )],
            unit='reqps',
          ),
          w=12, h=8,
          description='Normal: move endpoints dominate game traffic. Watch for: unexpected routes or disproportionate error rates. Check next: System Health for full HTTP breakdown.',
        ),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // ACT — What's the evidence? (1 panel)
    // ════════════════════════════════════════════════════════════════
    act=[
      // Game Event Logs (Loki)
      layout.panel(
        panels.logPanel(
          title='Game Event Logs',
          expr='{service_name="%s"} |= "game" ${gameId:pipe}' % svc,
        ),
        w=24, h=8,
        description='Normal: game creation, move execution, phase transitions. Watch for: error-level entries, panic recoveries. Check next: filter by Game ID using the $gameId variable above.',
      ),
    ],
  ),
)
