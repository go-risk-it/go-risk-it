// Database dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/database.jsonnet
// Regenerate: make dashboards
local common = import 'common.libsonnet';
local colors = import 'colors.libsonnet';
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
      gridPos: { h: 8, w: 8, x: 0, y: 1 },
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
      description: 'Connection acquires canceled while waiting \u2014 should be 0 under normal operation',
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
      gridPos: { h: 8, w: 8, x: 16, y: 1 },
    },

    // ── Orient — What's the shape? ─────────────────────────────────────
    ooda.orientRow() + { gridPos: { h: 1, w: 24, x: 0, y: 9 } },

    // Panel 4: Transaction Duration P50/P95/P99 (timeseries + SLO threshold line)
    common.timeseriesPanel(
      title='Transaction Duration P50/P95/P99',
      targets=common.histogramQuantileTargets(
        'db_transaction_duration_seconds_bucket',
        [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
      ),
      unit='s',
      color=colors.fixedColor(colors.db),
    ) + {
      id: 4,
      gridPos: { h: 8, w: 12, x: 0, y: 10 },
      fieldConfig+: {
        defaults+: {
          thresholds: thresholds.dbTxnP95,
          custom+: {
            thresholdsStyle: { mode: 'line+area' },
          },
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
      description: 'Rate of connection acquires that had to wait because the pool was empty \u2014 leading indicator of saturation',
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
      gridPos: { h: 8, w: 12, x: 12, y: 36 },
    },
  ],
}
