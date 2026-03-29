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

// Helper: per-phase histogram quantile targets for game_phase_duration_seconds_bucket.
local phaseLatencyTargets(phase) = [
  {
    expr: 'histogram_quantile(%s, sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it",phase="%s"}[1m])) by (le))' % [q[0], phase],
    legendFormat: q[1],
    refId: std.char(65 + i),
  }
  for i in std.range(0, 2)
  for q in [[['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']][i]]
];

// Helper: per-phase latency panel for the Move Timing collapsed row.
local phasePanel(phase, phaseName, description) =
  layout.panel(
    panels.timeseriesPanel(
      title='%s Phase Latency' % phaseName,
      targets=phaseLatencyTargets(phase),
      unit='s',
    ) + modifiers.withPercentileColors('gameLogic'),
    w=12, h=8,
    description=description,
  );

dashboard.new(
  uid='game-engine',
  title='Game Engine',
  description='Game engine observability: game lifecycle, phase timing, event bus health, and move analytics',
  tags=['risk-it', 'game-engine'],
  templating={ list: [gameIdVar] },
  annotations={ list: [dashboard.perfTestAnnotation] },
  panels=layout.ooda(

    // ════════════════════════════════════════════════════════════════
    // OBSERVE — Am I OK? (3 panels)
    // ════════════════════════════════════════════════════════════════
    observe=[
      // Active Games (stat) — migrated from P1
      layout.panel(
        panels.statPanel(
          title='Active Games',
          targets=[{
            expr: 'game_active{service_name="risk-it"}',
            legendFormat: 'Active Games',
            refId: 'A',
          }],
          thresholds=thresholds.activeGames,
        ) + modifiers.withLinks([links.toDashboard('Perf Test', links.dashboardUids.perfTest)]),
        w=8, h=8,
        description='Normal: < 50 active games (green). Watch for: sustained > 100 (red) may indicate stuck phases or games not finishing. Check next: Games Created/Finished Rate for flow balance.',
      ),

      // Games Created/Finished Rate (timeseries) — migrated from P2
      layout.panel(
        panels.timeseriesPanel(
          title='Games Created/Finished Rate',
          targets=[
            {
              expr: 'rate(game_created_total{service_name="risk-it"}[1m])',
              legendFormat: 'Created',
              refId: 'A',
            },
            {
              expr: 'rate(game_finished_total{service_name="risk-it"}[1m])',
              legendFormat: 'Finished',
              refId: 'B',
            },
          ],
          unit='ops',
        ),
        w=8, h=8,
        description='Normal: created and finished rates roughly balanced. Watch for: created >> finished means games accumulating (stuck phases). Check next: Active Games stat for current count.',
      ),

      // SHOWCASE: Game Event Heartbeat (stacked area, event type colors)
      layout.panel(
        panels.timeseriesPanel(
          title='Game Event Heartbeat',
          targets=[{
            expr: 'sum(rate(event_bus_events_total{service_name="risk-it"}[1m])) by (event_type)',
            legendFormat: '{{event_type}}',
            refId: 'A',
          }],
          unit='ops',
        ) + modifiers.withStackedArea(25, 'opacity') + modifiers.withSeriesColors(colors.eventTypes),
        w=8, h=8,
        description='Normal: move_executed dominates, all 8 event types visible. Watch for: any event type dropping to zero while others continue. Check next: Event Handler Latency in Decide for handler-level bottlenecks.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // ORIENT — What's the shape? (5 always-visible + 1 collapsed row)
    // ════════════════════════════════════════════════════════════════
    orient=[
      // Phase Duration Heatmap (Oranges) — migrated from P5
      layout.panel(
        panels.heatmapPanel(
          title='Phase Duration Heatmap',
          targets=[{
            expr: 'sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)',
            format: 'heatmap',
            legendFormat: '{{le}}',
            refId: 'A',
          }],
          unit='s',
          colorScheme='Oranges',
          colorFill='dark-orange',
        ),
        w=12, h=8,
        description='Normal: dense band in low-seconds range (fast phases). Watch for: heat spreading to higher buckets over time. Check next: Phase Duration P50/P95 for per-phase breakdown.',
      ),

      // Game Duration Heatmap (Blues) — migrated from P8
      layout.panel(
        panels.heatmapPanel(
          title='Game Duration Heatmap',
          targets=[{
            expr: 'sum(rate(game_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)',
            format: 'heatmap',
            legendFormat: '{{le}}',
            refId: 'A',
          }],
          unit='s',
          colorScheme='Blues',
          colorFill='dark-blue',
        ),
        w=12, h=8,
        description='Normal: concentrated distribution showing consistent game lengths. Watch for: bimodal distribution (fast + very slow games). Check next: Game Duration P50/P95 in Decide for percentile trends.',
      ),

      // Phase Duration P50/P95 by phase — migrated from P4
      layout.panel(
        panels.timeseriesPanel(
          title='Phase Duration P50/P95',
          targets=[
            {
              expr: 'histogram_quantile(0.5, sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, phase))',
              legendFormat: '{{phase}} P50',
              refId: 'A',
            },
            {
              expr: 'histogram_quantile(0.95, sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, phase))',
              legendFormat: '{{phase}} P95',
              refId: 'B',
            },
          ],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
        w=12, h=8,
        description='Normal: deploy and attack phases are longest; reinforce is shortest. Watch for: a single phase P95 spiking while others stay flat. Check next: Move Timing collapsed row for per-phase latency bands.',
      ),

      // Event Cascade Rate (continent captured/lost)
      layout.panel(
        panels.timeseriesPanel(
          title='Event Cascade Rate',
          targets=[
            {
              expr: 'sum(rate(event_bus_events_total{service_name="risk-it",event_type="continent_captured"}[1m]))',
              legendFormat: 'Captured',
              refId: 'A',
            },
            {
              expr: 'sum(rate(event_bus_events_total{service_name="risk-it",event_type="continent_lost"}[1m]))',
              legendFormat: 'Lost',
              refId: 'B',
            },
          ],
          unit='ops',
          overrides=[
            {
              matcher: { id: 'byName', options: 'Captured' },
              properties: [
                { id: 'color', value: colors.fixedColor(colors.http) },
              ],
            },
            {
              matcher: { id: 'byName', options: 'Lost' },
              properties: [
                { id: 'color', value: colors.fixedColor(colors.errors) },
              ],
            },
          ],
        ),
        w=12, h=8,
        description='Normal: captured and lost rates correlated with game activity. Watch for: captured >> lost (map consolidation) or lost spikes (cascade collapses). Check next: Player Elimination Rate for downstream impact.',
      ),

      // Player Elimination Rate
      layout.panel(
        panels.timeseriesPanel(
          title='Player Elimination Rate',
          targets=[{
            expr: 'sum(rate(event_bus_events_total{service_name="risk-it",event_type="player_eliminated"}[1m]))',
            legendFormat: 'Eliminated',
            refId: 'A',
          }],
          unit='ops',
          color=colors.fixedColor(colors.errors),
        ),
        w=12, h=8,
        description='Normal: low, steady rate as games progress to endgame. Watch for: sudden spikes (many games ending simultaneously) or zero rate (no games reaching endgame). Check next: Game Duration Heatmap for game length distribution.',
      ),
    ],

    orientDepth={
      // ── Collapsed: Move Timing (5 per-phase latency panels) ──
      'Move Timing': [
        phasePanel('DEPLOY', 'Deploy', 'Normal: p95 < 500ms. Watch for: p95 diverging from p50 (slow outliers). Check next: Database dashboard for transaction contention.'),
        phasePanel('ATTACK', 'Attack', 'Normal: p95 < 1s (attack involves dice + region updates). Watch for: p99 spikes (complex multi-region attacks). Check next: Conquer Phase Latency for post-attack overhead.'),
        phasePanel('CONQUER', 'Conquer', 'Normal: fastest phase (single troop movement). Watch for: p95 > 200ms indicates DB contention on region updates. Check next: Reinforce Phase Latency.'),
        phasePanel('REINFORCE', 'Reinforce', 'Normal: fast, single troop redistribution. Watch for: p95 > 200ms. Check next: Cards Phase Latency.'),
        phasePanel('CARDS', 'Cards', 'Normal: fast card redemption. Watch for: p95 spikes when many players redeem simultaneously. Check next: Phase Duration P50/P95 for aggregate view.'),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // DECIDE — Where's the bottleneck? (4 always-visible + 2 collapsed rows)
    // ════════════════════════════════════════════════════════════════
    decide=[
      // Moves per Second by Phase — migrated from P3
      layout.panel(
        panels.timeseriesPanel(
          title='Moves per Second by Phase',
          targets=[{
            expr: 'sum(rate(game_moves_total{service_name="risk-it"}[1m])) by (phase)',
            legendFormat: '{{phase}}',
            refId: 'A',
          }],
          unit='ops',
        ) + modifiers.withLinks(crossLinks),
        w=12, h=8,
        description='Normal: deploy most frequent, then attack, then reinforce/conquer. Watch for: a phase dropping to zero while others continue (phase blocked). Check next: Event Handler Latency for downstream processing speed.',
      ),

      // Game Duration P50/P95 — migrated from P7
      layout.panel(
        panels.timeseriesPanel(
          title='Game Duration P50/P95',
          targets=targets.histogramQuantileTargetsWithExemplars(
            'game_duration_seconds_bucket',
            [['0.5', 'p50'], ['0.95', 'p95']],
          ),
          unit='s',
        ) + modifiers.withPercentileColors('gameLogic'),
        w=12, h=8,
        description='Normal: consistent P50 with P95 within 2-3x of P50. Watch for: P95 growing while P50 stays flat (subset of slow games). Check next: Phase Duration P50/P95 in Orient to identify which phase is slow.',
      ),

      // Event Handler Latency by handler (p95)
      layout.panel(
        panels.timeseriesPanel(
          title='Event Handler Latency p95',
          targets=[{
            expr: 'histogram_quantile(0.95, sum(rate(event_handler_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, handler))',
            legendFormat: '{{handler}} p95',
            refId: 'A',
          }],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
        w=12, h=8,
        description='Normal: all handlers < 100ms p95. Watch for: individual handler p95 > 500ms (slow consumer bottleneck). Check next: Event Dispatch Duration for bus-level overhead.',
      ),

      // Event Dispatch Duration by event_type (p95)
      layout.panel(
        panels.timeseriesPanel(
          title='Event Dispatch Duration p95',
          targets=[{
            expr: 'histogram_quantile(0.95, sum(rate(event_bus_dispatch_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le, event_type))',
            legendFormat: '{{event_type}} p95',
            refId: 'A',
          }],
          unit='s',
        ),
        w=12, h=8,
        description='Normal: all event types dispatched < 50ms. Watch for: move_executed dispatch exceeding 100ms (many handlers subscribed). Check next: Event Bus Detail collapsed row for throughput and errors.',
      ),
    ],

    decideDepth={
      // ── Collapsed: Event Bus Detail (3 panels) ──
      'Event Bus Detail': [
        // Handler Throughput by event_type
        layout.panel(
          panels.timeseriesPanel(
            title='Handler Throughput',
            targets=[{
              expr: 'sum(rate(event_bus_events_total{service_name="risk-it"}[1m])) by (event_type)',
              legendFormat: '{{event_type}}',
              refId: 'A',
            }],
            unit='ops',
          ),
          w=12, h=8,
          description='Normal: move_executed dominates throughput. Watch for: unexpected event type surges. Check next: Event Handler Errors.',
        ),

        // Event Handler Errors
        layout.panel(
          panels.timeseriesPanel(
            title='Event Handler Errors',
            targets=[{
              expr: 'sum(rate(event_handler_errors_total{service_name="risk-it"}[1m])) by (handler)',
              legendFormat: '{{handler}}',
              refId: 'A',
            }],
            unit='ops',
          ),
          w=24, h=8,
          description='Normal: 0 errors. Watch for: any sustained error rate indicates handler panics or logic failures. Check next: Game Event Logs in Act for error details.',
        ),
      ],

      // ── Collapsed: Move Routes (2 panels) ──
      'Move Routes': [
        // Total Moves by Phase (bar gauge) — migrated from P6
        layout.panel(
          panels.barGaugePanel(
            title='Total Moves by Phase',
            targets=[{
              expr: 'sum(game_moves_total{service_name="risk-it"}) by (phase)',
              legendFormat: '{{phase}}',
              refId: 'A',
            }],
          ),
          w=12, h=8,
          description='Normal: deploy highest, attack second, conquer/reinforce lower. Watch for: unusual ratios (few conquer relative to attack). Check next: Moves per Second by Phase for rate view.',
        ),

        // Game HTTP Route Request Rate
        layout.panel(
          panels.timeseriesPanel(
            title='Game HTTP Route Request Rate',
            targets=[{
              expr: 'sum(rate(http_server_requests_total{service_name="risk-it",http_route=~".*games.*"}[1m])) by (http_route)',
              legendFormat: '{{http_route}}',
              refId: 'A',
            }],
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
      // Game Event Logs (Loki) — migrated from P9, improved with gameId variable
      layout.panel(
        panels.logPanel(
          title='Game Event Logs',
          expr='{service_name="risk-it"} |= "game" ${gameId:pipe}',
        ),
        w=24, h=8,
        description='Normal: game creation, move execution, phase transitions. Watch for: error-level entries, panic recoveries. Check next: filter by Game ID using the $gameId variable above.',
      ),
    ],
  ),
)
