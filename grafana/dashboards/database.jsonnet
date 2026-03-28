// Database dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/database.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
local links = import 'links.libsonnet';
local thresholds = import 'thresholds.libsonnet';
local ooda = import 'ooda.libsonnet';

{
  uid: 'database',
  title: 'Database',
  description: 'Database connection pool and query metrics',
  schemaVersion: 39,
  version: 1,
  tags: ['database', 'risk-it', 'pool'],
  timezone: 'browser',
  editable: true,
  time: { from: 'now-15m', to: 'now' },
  refresh: '10s',
  templating: { list: [] },
  annotations: { list: [] },

  panels: [
    // ── Observe — Am I OK? ─────────────────────────────────────────────
    ooda.observeRow() + { gridPos: { h: 1, w: 24, x: 0, y: 0 } },

    // Panel 1: Pool Utilization % (gauge)
    common.gaugePanel(
      title='Pool Utilization %',
      targets=[
        {
          expr: 'db_pool_active{service_name="risk-it"} / db_pool_total{service_name="risk-it"} * 100',
          legendFormat: 'utilization',
          refId: 'A',
        },
      ],
      thresholds=thresholds.poolUtil,
      unit='percent',
    ) + {
      id: 1,
      description: 'Normal: < 60% (green). Watch for: sustained > 80% (red) means pool is near exhaustion and requests will start queuing. Check next: pool saturation panel for empty acquire rate.',
      gridPos: { h: 8, w: 8, x: 0, y: 1 },
      fieldConfig+: {
        defaults+: {
          links: [links.toDashboard('Command Center', links.dashboardUids.perfTestCommandCenter)],
        },
      },
    },

    // Panel 2: Canceled Acquires (stat)
    common.statPanel(
      title='Canceled Acquires',
      targets=[
        {
          expr: 'rate(db_pool_canceled_acquires_total{service_name="risk-it"}[1m])',
          legendFormat: 'canceled/sec',
          refId: 'A',
        },
      ],
      thresholds=thresholds.canceledAcquires,
    ) + {
      id: 2,
      description: 'Normal: 0 canceled/sec (green). Watch for: any value > 0 (red) means requests are timing out waiting for a connection — pool is fully saturated. Check next: pool utilization gauge and connection pool usage timeseries.',
      gridPos: { h: 8, w: 8, x: 8, y: 1 },
    },

    // Panel 3: Postgres Cache Hit Rate (gauge, inverted threshold)
    common.gaugePanel(
      title='Postgres Cache Hit Rate',
      targets=[
        {
          expr: 'pg_stat_database_blks_hit_total{datname=~"postgres"} / (pg_stat_database_blks_hit_total{datname=~"postgres"} + pg_stat_database_blks_read_total{datname=~"postgres"}) * 100',
          legendFormat: 'cache hit %',
          refId: 'A',
        },
      ],
      thresholds=thresholds.cacheHit,
      unit='percent',
    ) + {
      id: 3,
      description: 'Normal: > 99% (green). Watch for: drops below 90% (red) indicate working set exceeds shared_buffers or new query patterns causing cold cache reads. Check next: rows read/written panel for query volume changes.',
      gridPos: { h: 8, w: 8, x: 16, y: 1 },
    },

    // ── Orient — What's the shape? ─────────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 9 } },

    // Panel 4: Transaction Duration P50/P95/P99 (timeseries + SLO threshold line)
    common.timeseriesPanel(
      title='Transaction Duration P50/P95/P99',
      targets=common.histogramQuantileTargetsWithExemplars(
        'db_transaction_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 4,
      description: 'Normal: p95 < 50ms (green SLO line). Watch for: p95 crossing 50ms or p99 > 100ms (red) — likely lock contention or complex queries. Check next: server-golden-signals HTTP latency to see if DB slowness is the dominant contributor.',
      gridPos: { h: 8, w: 12, x: 0, y: 10 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.dbTxnP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
          links: [links.toDashboard('Server Golden Signals', links.dashboardUids.serverGoldenSignals)],
        },
      },
    },

    // Panel 5: Connection Pool Usage (timeseries, 3 series — "total" gets dashed line override)
    common.timeseriesPanel(
      title='Connection Pool Usage',
      targets=[
        {
          expr: 'db_pool_active{service_name="risk-it"}',
          legendFormat: 'active (in-use)',
          refId: 'A',
        },
        {
          expr: 'db_pool_idle{service_name="risk-it"}',
          legendFormat: 'idle',
          refId: 'B',
        },
        {
          expr: 'db_pool_total{service_name="risk-it"}',
          legendFormat: 'total',
          refId: 'C',
        },
      ],
      unit='short',
      overrides=[
        {
          matcher: { id: 'byName', options: 'total' },
          properties: [
            { id: 'custom.lineStyle', value: { fill: 'dash', dash: [10, 10] } },
            { id: 'color', value: colors.fixedColor(colors.db) },
          ],
        },
      ],
      color=colors.fixedColor(colors.db),
    ) + {
      id: 5,
      description: 'Normal: active well below total (dashed line), healthy idle buffer. Watch for: active approaching total with idle dropping to 0 — pool exhaustion imminent. Check next: pool utilization gauge for the percentage view.',
      gridPos: { h: 8, w: 12, x: 12, y: 10 },
    },

    // ── Decide — Where's the bottleneck? ───────────────────────────────
    ooda.decideRow() + { gridPos: { h: 1, w: 24, x: 0, y: 18 } },

    // Panel 6: Avg Connection Acquire Time (timeseries + threshold line)
    common.timeseriesPanel(
      title='Avg Connection Acquire Time',
      targets=[
        {
          expr: 'db_pool_acquire_duration_seconds{service_name="risk-it"}',
          legendFormat: 'avg acquire time',
          refId: 'A',
        },
      ],
      unit='s',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 6,
      description: 'Normal: < 10ms (green SLO line). Watch for: spikes above 100ms (red) mean requests are waiting for connections — pool is undersized or transactions are held too long. Check next: pool saturation panel for empty acquire correlation.',
      gridPos: { h: 8, w: 8, x: 0, y: 19 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.dbAcquireTime,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
        },
      },
    },

    // Panel 7: Pool Saturation (timeseries, right-axis override for utilization %)
    common.timeseriesPanel(
      title='Pool Saturation (Empty Acquires)',
      targets=[
        {
          expr: 'rate(db_pool_empty_acquires_total{service_name="risk-it"}[1m])',
          legendFormat: 'empty acquires/sec',
          refId: 'A',
        },
        {
          expr: 'db_pool_active{service_name="risk-it"} / db_pool_max_conns{service_name="risk-it"} * 100',
          legendFormat: 'pool utilization %',
          refId: 'B',
        },
      ],
      unit='ops',
      overrides=[
        {
          matcher: { id: 'byName', options: 'pool utilization %' },
          properties: [
            { id: 'custom.axisPlacement', value: 'right' },
            { id: 'unit', value: 'percent' },
            { id: 'color', value: colors.fixedColor(colors.http) },
          ],
        },
      ],
      color=colors.fixedColor(colors.db),
    ) + {
      id: 7,
      description: 'Normal: 0 empty acquires/sec with utilization < 60%. Watch for: rising empty acquire rate — leading indicator that pool will saturate before canceled acquires appear. Check next: canceled acquires stat for confirmation of full exhaustion.',
      gridPos: { h: 8, w: 8, x: 8, y: 19 },
    },

    // Panel 8: Connection Acquire Rate (timeseries)
    common.timeseriesPanel(
      title='Connection Acquire Rate',
      targets=[
        {
          expr: 'rate(db_pool_acquires_total{service_name="risk-it"}[1m])',
          legendFormat: 'acquires/sec',
          refId: 'A',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 8,
      description: 'Normal: stable rate proportional to request throughput. Watch for: sudden spikes (retry storms) or drops (upstream failures stopping DB calls). Check next: server-golden-signals HTTP request rate for correlation.',
      gridPos: { h: 8, w: 8, x: 16, y: 19 },
    },

    // ── Act — What's the evidence? ─────────────────────────────────────
    ooda.actRow() + { gridPos: { h: 1, w: 24, x: 0, y: 27 } },

    // Panel 9: Transaction Rollbacks/sec (timeseries, fixed red)
    common.timeseriesPanel(
      title='Transaction Rollbacks/sec',
      targets=[
        {
          expr: 'rate(db_transaction_rollbacks_total{service_name="risk-it"}[1m])',
          legendFormat: 'rollbacks/sec',
          refId: 'A',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.errors),
    ) + {
      id: 9,
      description: 'Normal: 0 or near-zero rollbacks/sec. Watch for: sustained rollbacks indicate serialization conflicts, deadlocks, or application errors aborting transactions. Check next: server-golden-signals HTTP error rate for correlated 5xx responses.',
      gridPos: { h: 8, w: 12, x: 0, y: 28 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
        },
      },
    },

    // Panel 10: Postgres Active Connections (timeseries, fillOpacity=15)
    common.timeseriesPanel(
      title='Postgres Active Connections',
      targets=[
        {
          expr: 'pg_stat_database_numbackends{datname="postgres"}',
          legendFormat: 'active backends',
          refId: 'A',
        },
      ],
      unit='short',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 10,
      description: 'Normal: matches pool active count from application side. Watch for: backends exceeding pool max_conns — indicates leaked connections or external tools connecting directly. Check next: connection pool usage for the application-side view.',
      gridPos: { h: 8, w: 12, x: 12, y: 28 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
        },
      },
    },

    // Panel 11: Postgres Transactions/sec (timeseries, 2 series)
    common.timeseriesPanel(
      title='Postgres Transactions/sec',
      targets=[
        {
          expr: 'rate(pg_stat_database_xact_commit_total{datname=~"postgres"}[1m])',
          legendFormat: 'commits/sec',
          refId: 'A',
        },
        {
          expr: 'rate(pg_stat_database_xact_rollback_total{datname=~"postgres"}[1m])',
          legendFormat: 'rollbacks/sec',
          refId: 'B',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 11,
      description: 'Normal: steady commit rate with rollbacks near zero. Watch for: rollback rate climbing relative to commits — indicates contention or application errors. Check next: transaction rollbacks panel above for the application-level view.',
      gridPos: { h: 8, w: 12, x: 0, y: 36 },
    },

    // Panel 12: Postgres Rows Read/Written (timeseries, 3 series)
    common.timeseriesPanel(
      title='Postgres Rows Read/Written',
      targets=[
        {
          expr: 'rate(pg_stat_database_tup_fetched_total{datname=~"postgres"}[1m])',
          legendFormat: 'rows fetched/sec',
          refId: 'A',
        },
        {
          expr: 'rate(pg_stat_database_tup_inserted_total{datname=~"postgres"}[1m])',
          legendFormat: 'rows inserted/sec',
          refId: 'B',
        },
        {
          expr: 'rate(pg_stat_database_tup_updated_total{datname=~"postgres"}[1m])',
          legendFormat: 'rows updated/sec',
          refId: 'C',
        },
      ],
      unit='ops',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 12,
      description: 'Normal: fetched >> inserted/updated (read-heavy game state queries). Watch for: sudden ratio changes or spikes in updates — may indicate inefficient queries or unexpected write patterns. Check next: cache hit rate for I/O impact of changing query patterns.',
      gridPos: { h: 8, w: 12, x: 12, y: 36 },
    },
  ],
}
