// WebSocket dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/websocket.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local links = import 'links.libsonnet';
local ooda = import 'ooda.libsonnet';
local thresholds = import 'thresholds.libsonnet';

dashboard.new(
  uid='websocket',
  title='WebSocket',
  panels=[
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
      description: 'Normal: < 100 connections (green). Watch for: sustained > 500 (red) may indicate connection leaks or missing disconnects. Check next: server-golden-signals for overall resource pressure.',
      gridPos: { h: 8, w: 12, x: 0, y: 1 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter)],
        },
      },
    },

    // ── Orient — What's the shape? ──────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 9 } },

    // Panel 3: Broadcast Latency P50/P95/P99 (timeseries + SLO threshold line)
    common.timeseriesPanel(
      title='Broadcast Latency P50/P95/P99',
      targets=common.histogramQuantileTargetsWithExemplars(
        'ws_broadcast_duration_seconds_bucket',
        [['0.5', 'P50'], ['0.95', 'P95'], ['0.99', 'P99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.ws),
    ) + common.withSloThreshold(thresholds.wsDeliveryP95) + {
      id: 3,
      description: 'Normal: p95 < 200ms (green SLO line). Watch for: p95 crossing 200ms or p99 diverging sharply from p95 (tail latency). Check next: fan-out panel below to distinguish slow writes from high connection counts.',
      gridPos: { h: 8, w: 12, x: 0, y: 10 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter)],
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
      description: 'Normal: proportional to active games x players (~4 msgs per move). Watch for: sudden drops (broadcast failures) or spikes without matching game activity. Check next: game-engine for move rate correlation.',
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
      description: 'Normal: ~4 connections per broadcast (one per player). Watch for: P95 diverging from P50 indicates uneven game sizes or stale connections inflating fan-out. Check next: broadcast latency above — high fan-out with high latency means per-write cost is the bottleneck.',
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
      description: 'Normal: 0 errors/sec. Watch for: any sustained errors indicate broken connections not being cleaned up. Check next: active connections stat — errors without connection count dropping suggests a leak.',
      gridPos: { h: 8, w: 12, x: 0, y: 28 },
    },

    // Panel 6: WebSocket Broadcast Logs (Loki)
    common.logPanel(
      title='WebSocket Broadcast Logs',
      expr='{service_name="risk-it"} |= "broadcast" or |= "websocket" or |= "ws"',
    ) + {
      id: 6,
      description: 'Normal: Broadcast operations. Watch for: Error-level entries, panic recoveries.',
      gridPos: { h: 8, w: 12, x: 12, y: 28 },
    },
  ],
)
