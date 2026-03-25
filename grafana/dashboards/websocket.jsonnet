// WebSocket dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/websocket.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local ooda = import 'ooda.libsonnet';
local thresholds = import 'thresholds.libsonnet';

{
  uid: 'websocket',
  title: 'WebSocket',
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

    // Panel 1: Active Connections (stat)
    common.statPanel(
      title='Active Connections',
      targets=[
        {
          expr: 'ws_connections_active{service_name="risk-it"}',
          legendFormat: '{{instance}}',
          refId: 'A',
        },
      ],
      thresholds=thresholds.wsConnections,
    ) + {
      id: 1,
      gridPos: { h: 8, w: 12, x: 0, y: 1 },
    },

    // ── Orient — What's the shape? ──────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 9 } },

    // Panel 3: Broadcast Latency P50/P95/P99 (timeseries + SLO threshold line)
    common.timeseriesPanel(
      title='Broadcast Latency P50/P95/P99',
      targets=common.histogramQuantileTargets(
        'ws_broadcast_duration_seconds_bucket',
        [['0.5', 'P50'], ['0.95', 'P95'], ['0.99', 'P99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.ws),
    ) + {
      id: 3,
      gridPos: { h: 8, w: 12, x: 0, y: 10 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.wsDeliveryP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
        },
      },
    },

    // ── Decide — Where's the bottleneck? ────────────────────────────
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 18 } },

    // Panel 2: Messages Sent Rate (timeseries)
    common.timeseriesPanel(
      title='Messages Sent Rate',
      targets=[
        {
          expr: 'rate(ws_messages_sent_total{service_name="risk-it"}[1m])',
          legendFormat: '{{instance}}',
          refId: 'A',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.ws),
    ) + {
      id: 2,
      gridPos: { h: 8, w: 12, x: 0, y: 19 },
    },

    // Panel 5: Broadcast Fan-Out (timeseries)
    common.timeseriesPanel(
      title='Broadcast Fan-Out',
      targets=[
        {
          expr: 'histogram_quantile(0.5, sum(rate(ws_broadcast_fanout_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'P50',
          refId: 'A',
        },
        {
          expr: 'histogram_quantile(0.95, sum(rate(ws_broadcast_fanout_bucket{service_name="risk-it"}[1m])) by (le))',
          legendFormat: 'P95',
          refId: 'B',
        },
      ],
      unit='short',
      color=colors.fixedColor(colors.ws),
    ) + {
      id: 5,
      description: "Number of connections per broadcast \u2014 distinguishes 'each write is slow' from 'too many connections'",
      gridPos: { h: 8, w: 12, x: 12, y: 19 },
    },

    // ── Act — What's the evidence? ──────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 27 } },

    // Panel 4: Broadcast Errors Rate (timeseries, fixed red)
    common.timeseriesPanel(
      title='Broadcast Errors Rate',
      targets=[
        {
          expr: 'rate(ws_broadcast_errors_total{service_name="risk-it"}[1m])',
          legendFormat: '{{instance}}',
          refId: 'A',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.errors),
    ) + {
      id: 4,
      gridPos: { h: 8, w: 12, x: 0, y: 28 },
    },
  ],
}
