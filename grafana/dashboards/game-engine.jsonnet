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

// Link to Game Theater with $gameId pre-populated.
local gameTheaterLink = links.toDashboardWithVar(
  'Game Theater', links.dashboardUids.gameTheater, 'gameId', '${gameId}'
);

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

      // Active Games (stat) — gauge, keep manual metric
      layout.panel(
        panels.statPanel(
          title='Active Games',
          targets=[targets.target('game_active{service_name="%s"}' % svc, 'Active Games')],
          thresholds=thresholds.activeGames,
        ) + modifiers.withLinks([links.toDashboard('Perf Test', links.dashboardUids.perfTest)]),
        w=6, h=8,
        description='Normal: < 50 active games (green). Watch for: sustained > 100 (red) may indicate stuck phases or games not finishing. Check next: Games Created/Finished Rate for flow balance.',
      ),

      // Total Events (stat) — spanmetrics calls for bus dispatch spans
      layout.panel(
        panels.statPanel(
          title='Total Events',
          targets=[targets.target(
            'sum(%s{service_name="%s", span_name=~"%s"})' % [targets.spanmetricsMetric.calls, svc, targets.spans.busDispatch],
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

      // Handler Error Rate (stat — spanmetrics error rate for event handlers)
      layout.panel(
        panels.statPanel(
          title='Handler Error Rate',
          targets=[targets.spanErrorRate(targets.spans.eventHandler, 'errors/s')],
          thresholds=thresholds.handlerErrorRate,
          unit='ops',
        ),
        w=6, h=8,
        description='Normal: 0 errors (green). Watch for: any non-zero value (red) indicates handler failures. Check next: Event Bus Detail collapsed row for per-handler error breakdown.',
      ),

      // Game Duration P50 (stat — manual metric, game lifecycle)
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

      // Games Created/Finished Rate — spanmetrics for specific bus events
      layout.panel(
        panels.timeseriesPanel(
          title='Games Created/Finished Rate',
          targets=[
            targets.spanRate('bus:game_created', 'Created'),
            targets.spanRate('bus:game_completed', 'Finished') { refId: 'B' },
          ],
          unit='ops',
        ) + modifiers.withSeriesColors({ Created: colors.gameLogic, Finished: colors.errors }),
        w=12, h=8,
        description='Normal: created and finished rates roughly balanced. Watch for: created >> finished means games accumulating (stuck phases). Check next: Active Games stat for current count.',
      ),

      // Game Event Heartbeat (stacked area, semantic event colors)
      // Uses spanmetrics calls grouped by span_name for bus dispatch spans.
      // Colors: move_executed=muted gray (dominant baseline), headlines=bright, lifecycle=info blue.
      layout.panel(
        panels.timeseriesPanel(
          title='Game Event Heartbeat',
          targets=[targets.spanRateBy(targets.spans.busDispatch, 'span_name', '{{span_name}}')],
          unit='ops',
        ) + modifiers.withStackedArea(25, 'opacity')
          + modifiers.withSeriesColors({
            'bus:move_executed': colors.signal.muted,
            'bus:phase_transitioned': colors.eventTypes.phase_transitioned,
            'bus:game_created': colors.eventTypes.game_created,
            'bus:game_completed': colors.signal.info,
            'bus:player_connected': colors.eventTypes.player_connected,
            'bus:player_eliminated': colors.headline.elimination,
            'bus:continent_captured': colors.headline.continent,
            'bus:continent_lost': colors.eventTypes.continent_lost,
          }),
        w=12, h=8,
        description='Normal: bus:move_executed (gray) dominates, headline events (gold/red) visible as thin bands. Watch for: any event type dropping to zero while others continue. Check next: Headline Frequency stats or Event Handler Latency in Decide.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // ORIENT — What's the shape? (8 always-visible: hero + 2 heatmaps + phase P50/P99 + headline rates + 3 sparkline stats)
    // ════════════════════════════════════════════════════════════════
    orient=[
      // ── Row 1: Phase × Latency hero (full-width) ──
      // Orchestrate_move p95 broken down by phase — the single most diagnostic panel.
      layout.panel(
        panels.timeseriesPanel(
          title='Phase × Latency p95',
          targets=[targets.spanDurationBy(targets.spans.gameLogic, '0.95', 'phase', '{{phase}} p95', exemplars=true)],
          unit='s',
        ) + modifiers.withSeriesColors({
          'DEPLOY p95': colors.phase.deploy,
          'ATTACK p95': colors.phase.attack,
          'CARDS p95': colors.phase.cards,
          'CONQUER p95': colors.phase.conquer,
          'REINFORCE p95': colors.phase.reinforce,
        }) + modifiers.withLinks([gameTheaterLink]),
        w=24, h=8,
        description='Normal: DEPLOY and ATTACK p95 highest; REINFORCE and CONQUER lowest. Watch for: a single phase spiking while others stay flat (phase-specific regression). Check next: Phase Duration Heatmap for distribution shape, Moves per Second by Phase in Decide.',
      ),

      // ── Row 2: Heatmaps (w=12 each) ──

      // Phase Duration Heatmap (Oranges) — spanmetrics for game logic spans
      layout.panel(
        panels.heatmapPanel(
          title='Phase Duration Heatmap',
          targets=[targets.heatmapTarget(
            'sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) by (le)' % [targets.spanmetricsMetric.duration, svc, targets.spans.gameLogic],
          )],
          unit='ms',
          colorScheme='Oranges',
          colorFill='dark-orange',
        ),
        w=12, h=8,
        description='Normal: dense band in low-seconds range (fast phases). Watch for: heat spreading to higher buckets over time. Check next: Phase × Latency p95 hero for per-phase breakdown.',
      ),

      // Game Duration Heatmap (Blues) — manual metric, game lifecycle
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

      // ── Row 3: Phase P50/P99 + Headline rates (w=12 each) ──

      // Phase Latency P50/P99 by phase — spanmetrics grouped by phase
      layout.panel(
        panels.timeseriesPanel(
          title='Phase Latency P50/P99',
          targets=[
            targets.spanDurationBy(targets.spans.gameLogic, '0.5', 'phase', '{{phase}} P50', exemplars=true),
            targets.spanDurationBy(targets.spans.gameLogic, '0.99', 'phase', '{{phase}} P99', exemplars=true) { refId: 'B' },
          ],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
        w=12, h=8,
        description='Normal: deploy and attack phases are longest; reinforce is shortest. Watch for: P99 diverging from P50 (tail latency outliers), a single phase spiking while others stay flat. Check next: Game Duration P50/P95 in Decide.',
      ),

      // Headline Event Rates — spanmetrics for specific bus event spans
      layout.panel(
        panels.timeseriesPanel(
          title='Headline Event Rates',
          targets=[
            targets.spanRate('bus:continent_captured', 'Captured'),
            targets.spanRate('bus:continent_lost', 'Lost') { refId: 'B' },
            targets.spanRate('bus:player_eliminated', 'Eliminated') { refId: 'C' },
          ],
          unit='ops',
        ) + modifiers.withSeriesColors({
          Captured: colors.headline.continent,
          Lost: colors.eventTypes.continent_lost,
          Eliminated: colors.headline.elimination,
        }),
        w=12, h=8,
        description='Normal: captured and lost rates correlated with game activity; elimination is lower steady rate. Watch for: captured >> lost (map consolidation), sudden elimination spikes (many games ending). Check next: Headline Frequency sparklines below.',
      ),

      // ── Row 4: Headline Frequency sparkline stats (3x w=8) ──

      // Continent Captured rate sparkline
      layout.panel(
        panels.statPanel(
          title='Continent Captured',
          targets=[targets.spanRate('bus:continent_captured', 'captured/s')],
          thresholds={
            mode: 'absolute',
            steps: [{ color: colors.headline.continent, value: null }],
          },
          unit='ops',
        ),
        w=8, h=6,
        description='Normal: steady rate proportional to active games. Watch for: rate dropping to zero (no map consolidation happening). Check next: Headline Event Rates timeseries for trend.',
      ),

      // Player Eliminated rate sparkline
      layout.panel(
        panels.statPanel(
          title='Player Eliminated',
          targets=[targets.spanRate('bus:player_eliminated', 'eliminated/s')],
          thresholds={
            mode: 'absolute',
            steps: [{ color: colors.headline.elimination, value: null }],
          },
          unit='ops',
        ),
        w=8, h=6,
        description='Normal: low steady rate (eliminations are rare events). Watch for: sudden spikes (many games ending simultaneously). Check next: Game Duration Heatmap for game length distribution.',
      ),

      // Game Completed rate sparkline
      layout.panel(
        panels.statPanel(
          title='Game Completed',
          targets=[targets.spanRate('bus:game_completed', 'completed/s')],
          thresholds={
            mode: 'absolute',
            steps: [{ color: colors.signal.info, value: null }],
          },
          unit='ops',
        ),
        w=8, h=6,
        description='Normal: steady completions matching game creation rate. Watch for: rate dropping to zero while active games remain (stuck games). Check next: Active Games stat and Games Created/Finished Rate in Observe.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // DECIDE — Where's the bottleneck? (5 always-visible + 2 collapsed rows)
    // ════════════════════════════════════════════════════════════════
    decide=[
      // Moves per Second by Phase — spanmetrics for game logic grouped by phase
      layout.panel(
        panels.timeseriesPanel(
          title='Moves per Second by Phase',
          targets=[targets.spanRateBy(targets.spans.gameLogic, 'phase', '{{phase}}')],
          unit='ops',
        ) + modifiers.withSeriesColors({
          DEPLOY: colors.phase.deploy,
          ATTACK: colors.phase.attack,
          CARDS: colors.phase.cards,
          CONQUER: colors.phase.conquer,
          REINFORCE: colors.phase.reinforce,
        }) + modifiers.withLinks(crossLinks),
        w=12, h=8,
        description='Normal: deploy most frequent, then attack, then reinforce/conquer. Watch for: a phase dropping to zero while others continue (phase blocked). Check next: Event Handler Latency for downstream processing speed.',
      ),

      // Game Duration P50/P95 — manual metric, game lifecycle
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
        description='Normal: consistent P50 with P95 within 2-3x of P50. Watch for: P95 growing while P50 stays flat (subset of slow games). Check next: Phase × Latency p95 in Orient to identify which phase is slow.',
      ),

      // Event Handler Latency by handler (p95) — spanmetrics grouped by handler/span_name
      layout.panel(
        panels.timeseriesPanel(
          title='Event Handler Latency p95',
          targets=[targets.spanDurationBy(targets.spans.eventHandler, '0.95', 'span_name', '{{span_name}} p95', exemplars=true)],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('System Health', links.dashboardUids.systemHealth)]),
        w=12, h=8,
        description='Normal: all handlers < 100ms p95. Watch for: individual handler p95 > 500ms (slow consumer bottleneck). Check next: Event Dispatch Duration for bus-level overhead.',
      ),

      // Event Dispatch Duration by event_type (p95) — spanmetrics for bus dispatch
      layout.panel(
        panels.timeseriesPanel(
          title='Event Dispatch Duration p95',
          targets=[targets.spanDurationBy(targets.spans.busDispatch, '0.95', 'span_name', '{{span_name}}', exemplars=true)],
          unit='s',
        ),
        w=12, h=8,
        description='Normal: all event types dispatched < 50ms. Watch for: bus:move_executed dispatch exceeding 100ms (many handlers subscribed). Check next: Event Bus Detail collapsed row for throughput and errors.',
      ),

      // WS Fanout Cost — compare broadcast duration at different player counts
      layout.panel(
        panels.timeseriesPanel(
          title='WS Fanout Cost',
          targets=[
            targets.spanDuration(targets.spans.wsBroadcast, [['0.95', 'fanout=3 p95']], exemplars=true, extraLabels=', ws_fanout="3"')[0],
            targets.spanDuration(targets.spans.wsBroadcast, [['0.95', 'fanout=4 p95']], exemplars=true, extraLabels=', ws_fanout="4"')[0] { refId: 'B' },
            targets.spanDuration(targets.spans.wsBroadcast, [['0.5', 'fanout=3 p50']], exemplars=true, extraLabels=', ws_fanout="3"')[0] { refId: 'C' },
            targets.spanDuration(targets.spans.wsBroadcast, [['0.5', 'fanout=4 p50']], exemplars=true, extraLabels=', ws_fanout="4"')[0] { refId: 'D' },
          ],
          unit='s',
        ) + modifiers.withSeriesColors({
          'fanout=3 p95': colors.ws,
          'fanout=4 p95': colors.eventBus,
          'fanout=3 p50': colors.shades.ws.light,
          'fanout=4 p50': colors.shades.eventBus.light,
        }),
        w=24, h=8,
        description='Normal: fanout=4 slightly higher than fanout=3 (one extra WS write). Watch for: fanout=4 p95 disproportionately higher than fanout=3 (non-linear cost scaling). Check next: System Health for WS connection pool and DB pool pressure.',
      ),
    ],

    decideDepth={
      // ── Collapsed: Event Bus Detail (2 panels) ──
      'Event Bus Detail': [
        // Handler Throughput by event_type — spanmetrics bus dispatch
        layout.panel(
          panels.timeseriesPanel(
            title='Handler Throughput',
            targets=[targets.spanRateBy(targets.spans.busDispatch, 'span_name', '{{span_name}}')],
            unit='ops',
          ),
          w=12, h=8,
          description='Normal: bus:move_executed dominates throughput. Watch for: unexpected event type surges. Check next: Event Bus Events Total for cumulative count.',
        ),

        // Event Bus Events Total (stat) — spanmetrics calls cumulative
        layout.panel(
          panels.statPanel(
            title='Event Bus Events Total',
            targets=[targets.target(
              'sum(%s{service_name="%s", span_name=~"%s"})' % [targets.spanmetricsMetric.calls, svc, targets.spans.busDispatch],
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
        // Total Moves by Phase (bar gauge) — spanmetrics calls for game logic
        layout.panel(
          panels.barGaugePanel(
            title='Total Moves by Phase',
            targets=[targets.target(
              'sum(%s{service_name="%s", span_name=~"%s"}) by (phase)' % [targets.spanmetricsMetric.calls, svc, targets.spans.gameLogic],
              '{{phase}}',
            )],
          ),
          w=12, h=8,
          description='Normal: deploy highest, attack second, conquer/reinforce lower. Watch for: unusual ratios (few conquer relative to attack). Check next: Moves per Second by Phase for rate view.',
        ),

        // Game HTTP Route Request Rate — spanmetrics calls for HTTP spans by span_name
        layout.panel(
          panels.timeseriesPanel(
            title='Game HTTP Route Request Rate',
            targets=[targets.target(
              'sum(rate(%s{service_name="%s", span_name=~"(GET|POST|PUT|DELETE) .*/games.*"}[1m])) by (span_name)' % [targets.spanmetricsMetric.calls, svc],
              '{{span_name}}',
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
