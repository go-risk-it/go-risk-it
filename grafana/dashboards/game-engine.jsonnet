// Game Engine dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/game-engine.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
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
      gridPos: { h: 8, w: 12, x: 0, y: 1 },
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
      gridPos: { h: 8, w: 12, x: 0, y: 18 },
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
      gridPos: { h: 8, w: 12, x: 0, y: 27 },
    },

    // Panel 7: Game Duration P50/P95 (timeseries, fixed green)
    common.timeseriesPanel(
      title='Game Duration P50/P95',
      targets=common.histogramQuantileTargets(
        'game_duration_seconds_bucket',
        [['0.5', 'P50'], ['0.95', 'P95']],
      ),
      unit='s',
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 7,
      gridPos: { h: 8, w: 12, x: 12, y: 27 },
    },

    // ── Act — What's the evidence? ──────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 35 } },

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
      gridPos: { h: 8, w: 12, x: 12, y: 36 },
    },
  ],
}
