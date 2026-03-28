// Server Golden Signals dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/server-golden-signals.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local links = import 'links.libsonnet';
local thresholds = import 'thresholds.libsonnet';
local ooda = import 'ooda.libsonnet';

{
  uid: 'server-golden-signals',
  title: 'Server Golden Signals',
  description: 'Golden signals for the risk-it Go backend service',
  schemaVersion: 39,
  version: 1,
  timezone: 'browser',
  editable: true,
  time: { from: 'now-15m', to: 'now' },
  refresh: '10s',
  tags: ['go', 'otel', 'golden-signals'],
  templating: { list: [] },
  annotations: { list: [] },

  panels: [
    // ── Observe — Am I OK? ──────────────────────────────────────────
    ooda.observeRow() + { gridPos: { h: 1, w: 24, x: 0, y: 0 } },

    // Panel 1: HTTP Request Rate
    common.timeseriesPanel(
      title='HTTP Request Rate',
      targets=[
        {
          expr: 'sum(rate(http_server_requests_total{service_name="risk-it"}[1m]))',
          legendFormat: 'req/s',
          refId: 'A',
        },
      ],
      unit='reqps',
      color=colors.fixedColor(colors.http),
    ) + {
      id: 1,
      description: 'Normal: steady rate proportional to active games and players. Watch for: sudden drops (service unreachable) or unexpected spikes (retry storms, bot activity). Check next: HTTP request rate by route to identify which endpoints changed.',
      gridPos: { h: 8, w: 12, x: 0, y: 1 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter)],
        },
      },
    },

    // Panel 9: HTTP Error Rate % (fixed red, threshold line+area)
    common.timeseriesPanel(
      title='HTTP Error Rate %',
      targets=[
        {
          expr: 'sum(rate(http_server_requests_total{service_name="risk-it",http_status_code=~"[45].."}[1m])) / sum(rate(http_server_requests_total{service_name="risk-it"}[1m])) * 100',
          legendFormat: 'error %',
          refId: 'A',
        },
      ],
      unit='percent',
      color=colors.fixedColor(colors.errors),
    ) + {
      id: 9,
      description: 'Normal: < 1% error rate (green). Watch for: sustained > 5% (red) indicates systemic failures — distinguish 4xx (client validation) from 5xx (server bugs). Check next: database for connection/transaction errors if 5xx-heavy.',
      gridPos: { h: 8, w: 12, x: 12, y: 1 },
      fieldConfig+: {
        defaults+: {
          min: 0,
          thresholds: thresholds.httpError,
          custom+: {
            fillOpacity: 15,
            thresholdsStyle: { mode: 'line+area' },
          },
          links: [links.toDashboard('Game Engine', links.dashboardUids.gameEngine)],
        },
      },
    },

    // ── Orient — What's the shape? ──────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 9 } },

    // Panel 3: HTTP Latency P50/P95/P99 (histogram_quantile + SLO threshold)
    common.timeseriesPanel(
      title='HTTP Latency P50/P95/P99',
      targets=common.histogramQuantileTargetsWithExemplars(
        'http_server_request_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.http),
    ) + {
      id: 3,
      description: 'Normal: p95 < 500ms (green SLO line). Watch for: p95 crossing 500ms or p99 > 1s (red) — latency is user-visible at this point. Check next: database transaction duration to determine if DB is the latency bottleneck.',
      gridPos: { h: 8, w: 12, x: 0, y: 10 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.e2eP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
          links: [links.toDashboard('Database', links.dashboardUids.database)],
        },
      },
    },

    // Panel 10: Scheduler Latency P50/P95/P99 (histogram_quantile)
    common.timeseriesPanel(
      title='Scheduler Latency P50/P95/P99',
      targets=common.histogramQuantileTargetsWithExemplars(
        'go_schedule_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 10,
      description: 'Normal: p95 < 1ms (goroutines scheduled quickly). Watch for: p95 > 10ms indicates CPU saturation — goroutines are waiting in the run queue. Check next: process CPU usage to confirm CPU is the constraint.',
      gridPos: { h: 8, w: 12, x: 12, y: 10 },
    },

    // ── Decide — Where's the bottleneck? ────────────────────────────
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 18 } },

    // Panel 2: HTTP Request Rate by Route (multi-series, palette-classic)
    common.timeseriesPanel(
      title='HTTP Request Rate by Route',
      targets=[
        {
          expr: 'sum(rate(http_server_requests_total{service_name="risk-it"}[1m])) by (http_method, http_route, http_status_code)',
          legendFormat: '{{http_method}} {{http_route}} [{{http_status_code}}]',
          refId: 'A',
        },
      ],
      unit='reqps',
    ) + {
      id: 2,
      description: 'Normal: move endpoints dominate traffic; game creation is low-frequency. Watch for: unexpected routes appearing, single route with high error codes, or route disappearing from traffic. Check next: game-engine moves per second for game-level correlation.',
      gridPos: { h: 8, w: 12, x: 0, y: 19 },
    },

    // Panel 7: Process CPU Usage (stacked, per cpu_mode)
    common.timeseriesPanel(
      title='Process CPU Usage',
      targets=[
        {
          expr: 'rate(process_cpu_time_seconds_total{service_name="risk-it"}[1m])',
          legendFormat: '{{cpu_mode}}',
          refId: 'A',
        },
      ],
      unit='percentunit',
    ) + {
      id: 7,
      description: 'Normal: total CPU < 50% of available cores. Watch for: user CPU > 80% (compute-bound) or system CPU rising (excessive syscalls, GC pressure). Check next: scheduler latency — high CPU causes goroutine scheduling delays.',
      gridPos: { h: 8, w: 12, x: 12, y: 19 },
      fieldConfig+: {
        defaults+: {
          min: 0,
          custom+: {
            stacking+: { mode: 'normal' },
            fillOpacity: 20,
          },
        },
      },
    },

    // ── Act — What's the evidence? ──────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 27 } },

    // Panel 4: Goroutines (server runtime metric)
    common.timeseriesPanel(
      title='Goroutines',
      targets=[
        {
          expr: 'go_goroutine_count{service_name="risk-it"}',
          legendFormat: 'goroutines',
          refId: 'A',
        },
      ],
      unit='short',
      color=colors.fixedColor(colors.gameLogic),
    ) + {
      id: 4,
      description: 'Normal: stable count proportional to active games + connections. Watch for: monotonically increasing count (goroutine leak) or sudden spikes (broadcast storms, stuck handlers). Check next: websocket active connections — each connection holds a goroutine.',
      gridPos: { h: 8, w: 12, x: 0, y: 28 },
    },

    // Panel 5: Heap Memory (2 series: heap + stack)
    common.timeseriesPanel(
      title='Heap Memory',
      targets=[
        {
          expr: 'go_memory_used_bytes{service_name="risk-it",go_memory_type="other"}',
          legendFormat: 'heap',
          refId: 'A',
        },
        {
          expr: 'go_memory_used_bytes{service_name="risk-it",go_memory_type="stack"}',
          legendFormat: 'stack',
          refId: 'B',
        },
      ],
      unit='bytes',
    ) + {
      id: 5,
      description: 'Normal: heap stable with sawtooth GC pattern; stack proportional to goroutine count. Watch for: heap growing without GC reclaiming (memory leak) or stack growing unbounded (goroutine leak). Check next: GC goal panel for allocation pressure.',
      gridPos: { h: 8, w: 12, x: 12, y: 28 },
    },

    // Panel 6: GC Goal & Allocation Rate (2 series)
    common.timeseriesPanel(
      title='GC Goal & Allocation Rate',
      targets=[
        {
          expr: 'go_memory_gc_goal_bytes{service_name="risk-it"}',
          legendFormat: 'gc goal',
          refId: 'A',
        },
        {
          expr: 'rate(go_memory_allocated_bytes_total{service_name="risk-it"}[1m])',
          legendFormat: 'alloc rate',
          refId: 'B',
        },
      ],
      unit='Bps',
    ) + {
      id: 6,
      description: 'Normal: GC goal stable, allocation rate proportional to throughput. Watch for: GC goal ratcheting up (GOGC struggling to keep up) or allocation rate spiking without traffic increase (inefficient code path). Check next: heap memory for actual memory usage trend.',
      gridPos: { h: 8, w: 12, x: 0, y: 36 },
    },

    // Panel 8: Process Memory (2 series: heap + gc goal)
    common.timeseriesPanel(
      title='Process Memory',
      targets=[
        {
          expr: 'go_memory_used_bytes{service_name="risk-it"}',
          legendFormat: 'heap',
          refId: 'A',
        },
        {
          expr: 'go_memory_gc_goal_bytes{service_name="risk-it"}',
          legendFormat: 'gc goal',
          refId: 'B',
        },
      ],
      unit='bytes',
    ) + {
      id: 8,
      description: 'Normal: heap usage stays below GC goal with periodic collections. Watch for: heap consistently at or above GC goal — GC is running frequently and may cause latency spikes. Check next: HTTP latency for GC-induced tail latency.',
      gridPos: { h: 8, w: 12, x: 12, y: 36 },
    },
  ],
}
