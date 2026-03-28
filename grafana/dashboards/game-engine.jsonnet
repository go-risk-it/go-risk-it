// Game Engine dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/game-engine.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local links = import 'links.libsonnet';
local thresholds = import 'thresholds.libsonnet';
local ooda = import 'ooda.libsonnet';

{
  uid: 'game-engine',
  title: 'Game Engine',
  schemaVersion: 39,
  version: 1,
  timezone: 'browser',
  editable: true,
  time: { from: 'now-15m', to: 'now' },
  refresh: '10s',
  templating: { list: [] },
  annotations: { list: [] },

  panels: [
    // ── Observe — Am I OK? ──────────────────────────────────────────
    ooda.observeRow() + { gridPos: { h: 1, w: 24, x: 0, y: 0 } },

    // Panel 1: Active Games (stat)
    common.statPanel(
      title='Active Games',
      targets=[
        {
          expr: 'game_active{service_name="risk-it"}',
          legendFormat: 'Active Games',
          refId: 'A',
        },
      ],
      thresholds=thresholds.activeGames,
    ) + {
      id: 1,
      description: 'Normal: < 50 active games (green). Watch for: sustained > 100 (red) may indicate games not finishing or accumulating due to stuck phases. Check next: games created/finished rate for flow balance.',
      gridPos: { h: 8, w: 12, x: 0, y: 1 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter)],
        },
      },
    },

    // Panel 2: Games Created/Finished Rate (timeseries, fixed green)
    common.timeseriesPanel(
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
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 2,
      description: 'Normal: created and finished rates roughly balanced over time. Watch for: created >> finished means games are accumulating (check for stuck phases); finished >> created means backlog is draining. Check next: active games stat for current count.',
      gridPos: { h: 8, w: 12, x: 12, y: 1 },
    },

    // ── Orient — What's the shape? ──────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 9 } },

    // Panel 5: Phase Duration Heatmap (Oranges scheme)
    common.heatmapPanel(
      title='Phase Duration Heatmap',
      targets=[
        {
          expr: 'sum(rate(game_phase_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)',
          format: 'heatmap',
          legendFormat: '{{le}}',
          refId: 'A',
        },
      ],
      unit='s',
      colorScheme='Oranges',
      colorFill='dark-orange',
    ) + {
      id: 5,
      description: 'Normal: dense band in low-seconds range (fast phase transitions). Watch for: heat spreading to higher buckets over time — phases are taking longer, possibly due to slow moves or DB contention. Check next: phase duration P50/P95 for per-phase breakdown.',
      gridPos: { h: 8, w: 12, x: 0, y: 10 },
    },

    // Panel 8: Game Duration Heatmap (Blues scheme)
    common.heatmapPanel(
      title='Game Duration Heatmap',
      targets=[
        {
          expr: 'sum(rate(game_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)',
          format: 'heatmap',
          legendFormat: '{{le}}',
          refId: 'A',
        },
      ],
      unit='s',
      colorScheme='Blues',
      colorFill='dark-blue',
    ) + {
      id: 8,
      description: 'Normal: concentrated distribution showing consistent game lengths. Watch for: bimodal distribution (fast + very slow games) suggests some games are getting stuck mid-play. Check next: game duration P50/P95 for percentile trends over time.',
      gridPos: { h: 8, w: 12, x: 12, y: 10 },
    },

    // Panel 4: Phase Duration P50/P95 (timeseries, palette-classic for per-phase breakdown)
    common.timeseriesPanel(
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
    ) + {
      id: 4,
      description: 'Normal: deploy and attack phases are longest; reinforce is shortest. Watch for: a single phase P95 spiking while others stay flat — isolates which game phase has the bottleneck. Check next: database transaction duration to see if DB latency is the cause.',
      gridPos: { h: 8, w: 12, x: 0, y: 18 },
      fieldConfig+: {
        defaults+: {
          links: [
            links.toDashboard('Database', links.dashboardUids.database),
            links.toDashboard('Request Lifecycle', links.dashboardUids.requestLifecycle),
          ],
        },
      },
    },

    // ── Decide — Where's the bottleneck? ────────────────────────────
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 26 } },

    // Panel 3: Moves per Second by Phase (timeseries, palette-classic for dynamic labels)
    common.timeseriesPanel(
      title='Moves per Second by Phase',
      targets=[
        {
          expr: 'sum(rate(game_moves_total{service_name="risk-it"}[1m])) by (phase)',
          legendFormat: '{{phase}}',
          refId: 'A',
        },
      ],
      unit='ops',
    ) + {
      id: 3,
      description: 'Normal: deploy moves are most frequent, followed by attack, then reinforce/conquer. Watch for: a phase dropping to zero while others continue — that phase may be blocked or erroring. Check next: server-golden-signals HTTP error rate for failed move requests.',
      gridPos: { h: 8, w: 12, x: 0, y: 27 },
    },

    // Panel 7: Game Duration P50/P95 (timeseries, fixed green)
    common.timeseriesPanel(
      title='Game Duration P50/P95',
      targets=common.histogramQuantileTargetsWithExemplars(
        'game_duration_seconds_bucket',
        [['0.5', 'P50'], ['0.95', 'P95']],
      ),
      unit='s',
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 7,
      description: 'Normal: consistent P50 with P95 within 2-3x of P50. Watch for: P95 growing while P50 stays flat — a subset of games are taking much longer than typical. Check next: phase duration P50/P95 to identify which phase is slow in outlier games.',
      gridPos: { h: 8, w: 12, x: 12, y: 27 },
    },

    // ── Act — What's the evidence? ──────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 35 } },

    // Panel 9: Game Event Logs (Loki)
    {
      id: 9,
      title: 'Game Event Logs',
      description: 'Normal: Game creation, move execution, phase transitions. Watch for: Error-level entries, panic recoveries.',
      type: 'logs',
      datasource: { type: 'loki', uid: 'loki' },
      targets: [{
        refId: 'A',
        expr: '{service_name="risk-it"} |= "game"',
      }],
      gridPos: { h: 8, w: 12, x: 0, y: 36 },
      options: {
        showTime: true,
        showLabels: false,
        showCommonLabels: false,
        wrapLogMessage: true,
        prettifyLogMessage: false,
        enableLogDetails: true,
        sortOrder: 'Descending',
        dedupStrategy: 'none',
      },
    },

    // Panel 6: Total Moves by Phase (bar gauge, palette-classic)
    common.barGaugePanel(
      title='Total Moves by Phase',
      targets=[
        {
          expr: 'sum(game_moves_total{service_name="risk-it"}) by (phase)',
          legendFormat: '{{phase}}',
          refId: 'A',
        },
      ],
    ) + {
      id: 6,
      description: 'Normal: deploy has the highest total, attack second, conquer/reinforce lower. Watch for: unusual ratios — very few conquer moves relative to attack may indicate game logic issues. Check next: moves per second by phase for the rate view.',
      gridPos: { h: 8, w: 12, x: 12, y: 44 },
    },
  ],
}
