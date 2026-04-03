// System Health dashboard — generated from Jsonnet.
// Source of truth: grafana/dashboards/system-health.jsonnet
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

// Template variable for trace investigation.
local traceIdVar = {
  name: 'traceId',
  type: 'textbox',
  label: 'Trace ID',
  current: { value: '' },
};

// Shared data links for Observe tiles.
local observeLinks = [
  links.toDashboard('Game Engine', links.dashboardUids.gameEngine),
  links.toDashboard('Perf Test', links.dashboardUids.perfTest),
];

dashboard.new(
  uid='system-health',
  title='System Health',
  description='Consolidated system health: HTTP, DB, WS, runtime + Postgres internals, query performance. Progressive disclosure via collapsed rows.',
  tags=['risk-it', 'system-health'],
  templating={ list: [traceIdVar] },
  panels=layout.ooda(

    // ════════════════════════════════════════════════════════════════
    // OBSERVE — Am I OK? (7 panels: 4x stat bg + 3x stat)
    // ════════════════════════════════════════════════════════════════
    observe=[
      // Row 1: 4x w=6 stat panels with background color

      // HTTP Latency p95 — spanmetrics
      layout.panel(
        panels.statPanel(
          title='HTTP Latency p95',
          targets=targets.spanDuration(targets.spans.http, [['0.95', 'p95']]),
          thresholds=thresholds.e2eP95,
          unit='s',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 500ms p95 (green). Watch for: sustained yellow (> 500ms) or red (> 1s). Check next: Lifecycle Timing in Orient for breakdown.',
      ),

      // DB Transaction p95 — spanmetrics
      layout.panel(
        panels.statPanel(
          title='DB Transaction p95',
          targets=targets.spanDuration(targets.spans.db, [['0.95', 'p95']]),
          thresholds=thresholds.dbTxnP95,
          unit='s',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 50ms p95 (green). Watch for: yellow (> 50ms) signals pool pressure, red (> 100ms) signals contention. Check next: Database collapsed row in Orient.',
      ),

      // HTTP Error Rate % — spanmetrics calls with status_code filter
      layout.panel(
        panels.statPanel(
          title='HTTP Error Rate %',
          targets=[targets.target(
            'sum(rate(%s{service_name="%s", span_name=~"%s", status_code="STATUS_CODE_ERROR"}[1m])) / sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) * 100' % [targets.spanmetricsMetric.calls, svc, targets.spans.http, targets.spanmetricsMetric.calls, svc, targets.spans.http],
            'error %',
            'A',
          )],
          thresholds=thresholds.httpError,
          unit='percent',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 1% errors (green). Watch for: yellow (> 1%) may be transient; red (> 5%) indicates systematic failures. Check next: HTTP Error Rate timeseries in Server & HTTP.',
      ),

      // Pool Utilization — manual gauge, keep
      layout.panel(
        panels.statPanel(
          title='Pool Utilization',
          targets=[targets.target(
            'db_pool_active{service_name="%s"} / db_pool_total{service_name="%s"} * 100' % [svc, svc],
            'utilization',
            'A',
          )],
          thresholds=thresholds.poolUtil,
          unit='percent',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 60% utilization (green). Watch for: yellow (> 60%) under moderate load, red (> 80%) approaching saturation. Check next: Pool Usage timeseries in Database collapsed row.',
      ),

      // Row 2: 3x w=8 stat panels

      // Canceled Acquires — manual counter, keep
      layout.panel(
        panels.statPanel(
          title='Canceled Acquires',
          targets=[targets.target(
            'rate(db_pool_canceled_acquires_total{service_name="%s"}[1m])' % svc,
            'canceled/s',
            'A',
          )],
          thresholds=thresholds.canceledAcquires,
        ) + modifiers.withLinks(observeLinks),
        w=8, h=6,
        description='Normal: 0 canceled acquires (green). Watch for: any non-zero rate (red) means requests timing out waiting for DB connections. Check next: Pool Utilization and Acquire Time in Database row.',
      ),

      // Cache Hit Rate — postgres exporter, keep
      layout.panel(
        panels.statPanel(
          title='Cache Hit Rate',
          targets=[targets.target(
            'pg_stat_database_blks_hit_total{datname=~"postgres"} / (pg_stat_database_blks_hit_total{datname=~"postgres"} + pg_stat_database_blks_read_total{datname=~"postgres"}) * 100',
            'hit %',
            'A',
          )],
          thresholds=thresholds.cacheHit,
          unit='percent',
        ) + modifiers.withLinks(observeLinks),
        w=8, h=6,
        description='Normal: > 99% cache hit rate (green). Watch for: yellow (< 99%) indicates working set exceeds shared_buffers; red (< 90%) severe miss rate. Check next: Postgres Internals collapsed row.',
      ),

      // WS Connections — manual gauge, keep
      layout.panel(
        panels.statPanel(
          title='WS Connections',
          targets=[targets.target(
            'ws_connections_active{service_name="%s"}' % svc,
            'active',
            'A',
          )],
          thresholds=thresholds.wsConnections,
        ) + modifiers.withLinks(observeLinks),
        w=8, h=6,
        description='Normal: < 100 connections (green). Watch for: yellow (> 100) under moderate load; red (> 500) may indicate leak. Check next: WebSocket collapsed row for broadcast health.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // ORIENT — What's the shape? (2 always-visible + 3 collapsed rows)
    // ════════════════════════════════════════════════════════════════
    orient=[
      // Lifecycle Timing hero panel (w=24, h=10) — spanmetrics via spanLifecycleTargets
      layout.panel(
        panels.timeseriesPanel(
          title='Lifecycle Timing (p95)',
          targets=targets.spanLifecycleTargets,
          unit='s',
          overrides=targets.lifecycleOverrides,
        ),
        w=24, h=10,
        description='Normal: HTTP Total < 500ms, DB < 50ms, Game Logic < 200ms, WS < 100ms. Watch for: any boundary diverging from baseline. Check next: collapsed rows below for per-boundary detail.',
      ),

      // Percentile Bands panel (w=24, h=8) — spanmetrics
      layout.panel(
        panels.spanPercentileBandsPanel(
          title='HTTP Latency Percentile Bands',
          spanNameFilter=targets.spans.http,
          unit='s',
          exemplars=true,
        ),
        w=24, h=8,
        description='Normal: tight bands (p50-p95 close together). Watch for: p99 band widening (tail latency divergence). Check next: Server & HTTP collapsed row for per-route breakdown.',
      ),

      // Isolation Rhythm (w=12, h=8) — read committed vs repeatable read p95
      layout.panel(
        panels.timeseriesPanel(
          title='Isolation Rhythm',
          targets=[
            targets.spanDuration(targets.spans.db, [['0.95', 'read committed p95']], exemplars=true, extraLabels=', isolation="read committed"')[0],
            targets.spanDuration(targets.spans.db, [['0.95', 'repeatable read p95']], exemplars=true, extraLabels=', isolation="repeatable read"')[0] { refId: 'B' },
          ],
          unit='s',
        ) + modifiers.withSeriesColors({
          'read committed p95': colors.shades.db.light,
          'repeatable read p95': colors.shades.db.dark,
        }) + modifiers.withLinks([links.toDashboard('Game Engine', links.dashboardUids.gameEngine)]),
        w=12, h=8,
        description='Normal: both isolation levels < 50ms p95. Watch for: repeatable read diverging from read committed (serialization retries). Check next: Database collapsed row for pool pressure.',
      ),

      // Handler Performance (w=12, h=8) — p95 latency per bus handler
      layout.panel(
        panels.barGaugePanel(
          title='Handler Performance',
          targets=[targets.spanDurationBy(targets.spans.eventHandler, '0.95', 'span_name', '{{span_name}}')],
          unit='s',
        ) + modifiers.withLinks([links.toDashboard('Game Engine', links.dashboardUids.gameEngine)]),
        w=12, h=8,
        description='Normal: all handlers < 100ms p95. Watch for: individual handlers exceeding 200ms (slow consumer). Check next: Event Bus collapsed row for throughput.',
      ),
    ],

    orientDepth={
      // ── Collapsed: Database (~9 panels) ──
      Database: [
        // Txn Duration P50/P95/P99 — spanmetrics
        layout.panel(
          panels.timeseriesPanel(
            title='Txn Duration P50/P95/P99',
            targets=targets.spanDuration(targets.spans.db, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], exemplars=true),
            unit='s',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: p95 < 50ms, p99 < 100ms. Watch for: p99 diverging from p95 (long-tail transactions). Check next: Pool Usage for saturation.',
        ),

        // Pool Usage — manual gauges, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Pool Usage',
            targets=[
              targets.target('db_pool_active{service_name="%s"}' % svc, 'Active', 'A'),
              targets.target('db_pool_idle{service_name="%s"}' % svc, 'Idle', 'B'),
              targets.target('db_pool_total{service_name="%s"}' % svc, 'Total', 'C'),
            ],
            unit='short',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: active well below total, healthy idle buffer. Watch for: active approaching total (saturation). Check next: Acquire Time for queue pressure.',
        ),

        // Acquire Time — manual gauge, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Acquire Time (avg)',
            targets=[targets.target(
              'db_pool_acquire_duration_seconds{service_name="%s"}' % svc,
              'avg acquire time',
              'A',
            )],
            unit='s',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: < 1ms average acquire time. Watch for: sustained > 10ms (pool exhaustion starting). Check next: Canceled Acquires in Observe.',
        ),

        // Pool Saturation — manual gauge, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Pool Saturation',
            targets=[targets.target(
              'db_pool_active{service_name="%s"} / db_pool_total{service_name="%s"} * 100' % [svc, svc],
              'utilization %',
              'A',
            )],
            unit='percent',
          ) + modifiers.withSloThreshold(thresholds.poolUtil),
          w=12, h=8,
          description='Normal: < 60% with SLO line visible. Watch for: crossing yellow threshold under normal load. Check next: Acquire Rate for demand side.',
        ),

        // Acquire Rate — manual counter, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Acquire Rate',
            targets=[targets.target(
              'rate(db_pool_acquires_total{service_name="%s"}[1m])' % svc,
              'acquires/s',
              'A',
            )],
            unit='ops',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: proportional to request rate. Watch for: spikes not correlated with traffic. Check next: Rollbacks for failed transactions.',
        ),

        // Rollbacks — postgres exporter, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Rollbacks',
            targets=[targets.target(
              'rate(pg_stat_database_xact_rollback_total{datname=~"postgres"}[1m])',
              'rollbacks/s',
              'A',
            )],
            unit='ops',
            color=colors.fixedColor(colors.errors),
          ),
          w=12, h=8,
          description='Normal: near zero rollback rate. Watch for: sustained rollbacks indicating transaction conflicts. Check next: PG Connections for connection-level issues.',
        ),

        // PG Connections — postgres exporter, keep
        layout.panel(
          panels.timeseriesPanel(
            title='PG Connections',
            targets=[
              targets.target('pg_stat_database_numbackends{datname=~"postgres"}', 'backends', 'A'),
            ],
            unit='short',
            color=colors.fixedColor(colors.db),
          ),
          w=8, h=8,
          description='Normal: stable connection count matching pool size. Watch for: gradual increase (connection leak). Check next: PG Txns/sec for throughput.',
        ),

        // PG Txns/sec — postgres exporter, keep
        layout.panel(
          panels.timeseriesPanel(
            title='PG Txns/sec',
            targets=[
              targets.target('rate(pg_stat_database_xact_commit_total{datname=~"postgres"}[1m])', 'commits/s', 'A'),
              targets.target('rate(pg_stat_database_xact_rollback_total{datname=~"postgres"}[1m])', 'rollbacks/s', 'B'),
            ],
            unit='ops',
          ),
          w=8, h=8,
          description='Normal: commits >> rollbacks. Watch for: rollback rate approaching commit rate. Check next: PG Rows for row-level throughput.',
        ),

        // PG Rows — postgres exporter, keep
        layout.panel(
          panels.timeseriesPanel(
            title='PG Rows',
            targets=[
              targets.target('rate(pg_stat_database_tup_returned_total{datname=~"postgres"}[1m])', 'returned/s', 'A'),
              targets.target('rate(pg_stat_database_tup_fetched_total{datname=~"postgres"}[1m])', 'fetched/s', 'B'),
              targets.target('rate(pg_stat_database_tup_inserted_total{datname=~"postgres"}[1m])', 'inserted/s', 'C'),
              targets.target('rate(pg_stat_database_tup_updated_total{datname=~"postgres"}[1m])', 'updated/s', 'D'),
            ],
            unit='ops',
          ),
          w=8, h=8,
          description='Normal: returned/fetched dominate (read-heavy workload). Watch for: updated/inserted spikes not correlated with game activity. Check next: Postgres Internals in Decide.',
        ),
      ],

      // ── Collapsed: Server & HTTP (~10 panels) ──
      'Server & HTTP': [
        // HTTP Request Rate — spanmetrics
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Request Rate',
            targets=[targets.spanRate(targets.spans.http, 'req/s')],
            unit='reqps',
            color=colors.fixedColor(colors.http),
          ),
          w=12, h=8,
          description='Normal: stable request rate matching expected load. Watch for: sudden drops (upstream issue) or spikes (load surge). Check next: HTTP Error Rate % for failure ratio.',
        ),

        // HTTP Error Rate % timeseries — spanmetrics
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Error Rate %',
            targets=[targets.target(
              'sum(rate(%s{service_name="%s", span_name=~"%s", status_code="STATUS_CODE_ERROR"}[1m])) / sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) * 100' % [targets.spanmetricsMetric.calls, svc, targets.spans.http, targets.spanmetricsMetric.calls, svc, targets.spans.http],
              'error %',
              'A',
            )],
            unit='percent',
            color=colors.fixedColor(colors.errors),
          ) + modifiers.withSloThreshold(thresholds.httpError),
          w=12, h=8,
          description='Normal: < 1% with SLO line visible. Watch for: error rate crossing SLO threshold. Check next: HTTP Latency P50/P95/P99 for latency correlation.',
        ),

        // HTTP Latency P50/P95/P99 — spanmetrics
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Latency P50/P95/P99',
            targets=targets.spanDuration(targets.spans.http, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], exemplars=true),
            unit='s',
            color=colors.fixedColor(colors.http),
          ),
          w=24, h=8,
          description='Normal: p50 < 100ms, p95 < 500ms, p99 < 1s. Watch for: p99 diverging sharply from p95 (long tail). Check next: HTTP Rate by Route for per-endpoint breakdown.',
        ),

        // HTTP Rate by Route — spanmetrics grouped by span_name
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Rate by Route',
            targets=[targets.spanRateBy(targets.spans.http, 'span_name', '{{span_name}}')],
            unit='reqps',
            color=colors.fixedColor(colors.http),
          ),
          w=24, h=8,
          description='Normal: move endpoints dominate. Watch for: unusual routes or disproportionate rates. Check next: Game Engine dashboard for game-specific routes.',
        ),

        // Process CPU — runtime metric, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Process CPU',
            targets=[targets.target(
              'rate(process_cpu_time_seconds_total{service_name="%s"}[1m])' % svc,
              'CPU cores',
              'A',
            )],
            unit='short',
            color=colors.fixedColor(colors.gameLogic),
          ),
          w=8, h=8,
          description='Normal: < 1 core under normal load. Watch for: sustained increase correlated with game count. Check next: Scheduler Latency for GC pressure.',
        ),

        // Scheduler Latency — runtime metric, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Scheduler Latency',
            targets=[targets.target(
              'histogram_quantile(0.95, sum(rate(go_schedule_duration_seconds_bucket{service_name="%s"}[1m])) by (le))' % svc,
              'p95',
              'A',
              exemplar=true,
            )],
            unit='s',
          ),
          w=8, h=8,
          description='Normal: < 1ms scheduler latency. Watch for: spikes during GC pauses or high goroutine count. Check next: Goroutines for concurrency.',
        ),

        // Goroutines — runtime metric, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Goroutines',
            targets=[targets.target(
              'go_goroutine_count{service_name="%s"}' % svc,
              'goroutines',
              'A',
            )],
            unit='short',
            color=colors.fixedColor(colors.gameLogic),
          ),
          w=8, h=8,
          description='Normal: stable count proportional to active games. Watch for: monotonic increase (goroutine leak). Check next: Heap Memory for memory pressure.',
        ),

        // Heap Memory — runtime metric, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Heap Memory',
            targets=[
              targets.target('go_memstats_heap_alloc_bytes{service_name="%s"}' % svc, 'Heap Alloc', 'A'),
              targets.target('go_memstats_heap_sys_bytes{service_name="%s"}' % svc, 'Heap Sys', 'B'),
              targets.target('go_memory_used_bytes{service_name="%s"}' % svc, 'Used', 'C'),
            ],
            unit='bytes',
          ),
          w=8, h=8,
          description='Normal: heap alloc tracks used with sawtooth GC pattern. Watch for: used growing without release (heap pressure). Check next: GC Goal for tuning.',
        ),

        // GC Goal — runtime metric, keep
        layout.panel(
          panels.timeseriesPanel(
            title='GC Goal',
            targets=[
              targets.target('go_config_gogc_percent{service_name="%s"}' % svc, 'GOGC %', 'A'),
              targets.target('go_memory_limit_bytes{service_name="%s"}' % svc, 'GOMEMLIMIT', 'B'),
            ],
            unit='short',
          ),
          w=8, h=8,
          description='Normal: GOGC at 100%, GOMEMLIMIT stable or absent. Watch for: GOMEMLIMIT being approached by heap usage (OOM risk). Check next: Runtime Memory for total footprint.',
        ),

        // Runtime Memory — runtime metric, keep
        layout.panel(
          panels.timeseriesPanel(
            title='Runtime Memory',
            targets=[targets.target(
              'go_memory_used_bytes{service_name="%s"}' % svc,
              'Used',
              'A',
            )],
            unit='bytes',
          ),
          w=8, h=8,
          description='Normal: stable memory proportional to heap + stack. Watch for: monotonic growth (memory leak). Check next: Goroutines for stack memory contributors.',
        ),
      ],

      // ── Collapsed: WebSocket (~4 panels) ──
      WebSocket: [
        // Broadcast Latency P50/P95/P99 — spanmetrics
        layout.panel(
          panels.timeseriesPanel(
            title='Broadcast Latency P50/P95/P99',
            targets=targets.spanDuration(targets.spans.wsBroadcast, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], exemplars=true),
            unit='s',
            color=colors.fixedColor(colors.ws),
          ),
          w=12, h=8,
          description='Normal: p95 < 100ms. Watch for: p99 > 500ms (slow consumers blocking broadcast). Check next: Messages Rate for throughput.',
        ),

        // Messages Rate — spanmetrics: rate of ws.broadcast spans
        layout.panel(
          panels.timeseriesPanel(
            title='Messages Rate',
            targets=[targets.spanRate(targets.spans.wsBroadcast, 'broadcasts/s')],
            unit='ops',
            color=colors.fixedColor(colors.ws),
          ),
          w=12, h=8,
          description='Normal: steady rate proportional to active games. Watch for: rate dropping to zero while games are active. Check next: Fan-Out for amplification ratio.',
        ),

        // Fan-Out — average ws_fanout dimension from broadcast spans
        layout.panel(
          panels.timeseriesPanel(
            title='Fan-Out',
            targets=[targets.target(
              'avg(traces_span_metrics_duration_milliseconds_sum{service_name="%s", span_name=~"%s"}) by ()' % [svc, targets.spans.wsBroadcast],
              'avg fanout',
              'A',
            )],
            unit='short',
            color=colors.fixedColor(colors.ws),
          ),
          w=12, h=8,
          description='Normal: ~4 (one message per connected player per broadcast). Watch for: 0 (no connected players). Check next: WS Connections.',
        ),

        // WS Errors — spanmetrics: error rate of ws.broadcast spans
        layout.panel(
          panels.timeseriesPanel(
            title='WS Errors',
            targets=[targets.spanErrorRate(targets.spans.wsBroadcast, 'errors/s')],
            unit='ops',
            color=colors.fixedColor(colors.errors),
          ),
          w=12, h=8,
          description='Normal: 0 errors. Watch for: any sustained error rate (broken connections, write failures). Check next: WS Connections in Observe for connection count.',
        ),
      ],

      // ── Collapsed: Event Bus (~2 panels) ──
      'Event Bus': [
        // Handler Throughput — rate of consumer spans by handler
        layout.panel(
          panels.timeseriesPanel(
            title='Handler Throughput',
            targets=[targets.spanRateBy(targets.spans.eventHandler, 'span_name', '{{span_name}}')],
            unit='ops',
            color=colors.fixedColor(colors.eventBus),
          ),
          w=12, h=8,
          description='Normal: rate proportional to move rate (each move triggers multiple handlers). Watch for: rate dropping to zero while games are active. Check next: Handler Latency for slow consumers.',
        ),

        // Handler Latency P95 — p95 latency per handler
        layout.panel(
          panels.timeseriesPanel(
            title='Handler Latency P95',
            targets=[targets.spanDurationBy(targets.spans.eventHandler, '0.95', 'span_name', '{{span_name}}', exemplars=true)],
            unit='s',
            color=colors.fixedColor(colors.eventBus),
          ),
          w=12, h=8,
          description='Normal: all handlers < 100ms p95. Watch for: individual handlers diverging (slow snapshot fetch or WS broadcast). Check next: Handler Performance bar gauge in Orient.',
        ),
      ],

      // ── Collapsed: Postgres Advanced (~3 panels) ──
      'Postgres Advanced': [
        // Lock Activity — pg_locks by mode
        layout.panel(
          panels.timeseriesPanel(
            title='Lock Activity',
            targets=[targets.target(
              'sum(pg_locks_count) by (mode)',
              '{{mode}}',
              'A',
            )],
            unit='short',
          ) + modifiers.withSeriesColors(colors.lockModes),
          w=8, h=8,
          description='Normal: AccessShareLock dominates (read-heavy workload). Watch for: ExclusiveLock or RowExclusiveLock spikes (write contention). Check next: Postgres Internals in Decide for full lock breakdown.',
        ),

        // Dead Tuples — by table
        layout.panel(
          panels.timeseriesPanel(
            title='Dead Tuples',
            targets=[targets.target(
              'sum(pg_stat_user_tables_n_dead_tup) by (relname)',
              '{{relname}}',
              'A',
            )],
            unit='short',
          ),
          w=8, h=8,
          description='Normal: low dead tuple count (autovacuum keeping up). Watch for: monotonic growth on any table (vacuum falling behind). Check next: Postgres Internals in Decide for table scans.',
        ),

        // Cache Efficiency — per-table cache hit ratio
        layout.panel(
          panels.timeseriesPanel(
            title='Cache Efficiency',
            targets=[targets.target(
              'rate(pg_statio_user_tables_heap_blocks_hit_total[1m]) / clamp_min(rate(pg_statio_user_tables_heap_blocks_hit_total[1m]) + rate(pg_statio_user_tables_heap_blocks_read_total[1m]), 0.001) * 100',
              '{{relname}}',
              'A',
            )],
            unit='percent',
          ),
          w=8, h=8,
          description='Normal: > 99% per table. Watch for: individual tables dropping below 95% (hot table exceeding buffer cache). Check next: Per-Table Cache Hit in Postgres Internals for trend.',
        ),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // DECIDE — Where's the bottleneck? (3 always-visible + 2 collapsed rows)
    // ════════════════════════════════════════════════════════════════
    decide=[
      // Fan-out Amplification — spanmetrics: broadcast rate / move event rate
      layout.panel(
        panels.timeseriesPanel(
          title='Fan-out Amplification',
          targets=[targets.target(
            'sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) / sum(rate(%s{service_name="%s", span_name="bus:move_executed"}[1m]))' % [targets.spanmetricsMetric.calls, svc, targets.spans.wsBroadcast, targets.spanmetricsMetric.calls, svc],
            'broadcasts/move',
            'A',
          )],
          unit='short',
          color=colors.fixedColor(colors.ws),
        ) + modifiers.withSloThreshold(thresholds.ccFanOut),
        w=8, h=8,
        description='Normal: ~4x (one per player). Watch for: > 10x (SLO yellow) suggests broadcast storms. Check next: WebSocket collapsed row for per-message detail.',
      ),

      // DB Latency Share — spanmetrics for both numerator and denominator
      layout.panel(
        panels.timeseriesPanel(
          title='DB Latency Share',
          targets=[targets.target(
            'histogram_quantile(0.95, sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) by (le)) / histogram_quantile(0.95, sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) by (le)) * 100' % [targets.spanmetricsMetric.duration, svc, targets.spans.db, targets.spanmetricsMetric.duration, svc, targets.spans.http],
            'DB % of HTTP p95',
            'A',
          )],
          unit='percent',
          color=colors.fixedColor(colors.db),
        ) + modifiers.withSloThreshold(thresholds.ccDbLatencyShare),
        w=8, h=8,
        description='Normal: 30-50% (DB is a fraction of total latency). Watch for: > 70% (SLO yellow) means DB dominates request time. Check next: Postgres Internals collapsed row.',
      ),

      // Runtime Memory — runtime metric, keep
      layout.panel(
        panels.timeseriesPanel(
          title='Runtime Memory',
          targets=[
            targets.target('go_memory_used_bytes{service_name="%s"}' % svc, 'Used', 'A'),
            targets.target('go_memstats_heap_alloc_bytes{service_name="%s"}' % svc, 'Heap Alloc', 'B'),
          ],
          unit='bytes',
        ),
        w=8, h=8,
        description='Normal: used memory tracks heap alloc with stable overhead. Watch for: used growing without GC reclaiming (heap leak). Check next: Goroutines in Server & HTTP row.',
      ),
    ],

    decideDepth={
      // ── Collapsed: Postgres Internals (~7 panels) ──
      'Postgres Internals': [
        // SHOWCASE: Lock Contention (stacked area with lock mode colors)
        layout.panel(
          panels.timeseriesPanel(
            title='Lock Contention',
            targets=[targets.target(
              'sum(pg_locks_count) by (mode)',
              '{{mode}}',
              'A',
            )],
            unit='short',
          ) + modifiers.withStackedArea(30, 'hue') + modifiers.withSeriesColors(colors.lockModes),
          w=24, h=10,
          description='Normal: AccessShareLock dominates (read locks, green). Watch for: ExclusiveLock or AccessExclusiveLock spikes (red/orange — DDL or heavy writes). Check next: Deadlocks for escalated contention.',
        ),

        // Deadlocks
        layout.panel(
          panels.timeseriesPanel(
            title='Deadlocks',
            targets=[targets.target(
              'rate(pg_stat_database_deadlocks_total{datname=~"postgres"}[1m])',
              'deadlocks/s',
              'A',
            )],
            unit='ops',
            color=colors.fixedColor(colors.errors),
          ),
          w=12, h=8,
          description='Normal: 0 deadlocks. Watch for: any non-zero rate indicates transaction ordering issues. Check next: Lock Contention for lock type breakdown.',
        ),

        // Dead Tuples
        layout.panel(
          panels.timeseriesPanel(
            title='Dead Tuples',
            targets=[targets.target(
              'sum(pg_stat_user_tables_n_dead_tup) by (relname)',
              '{{relname}}',
              'A',
            )],
            unit='short',
          ),
          w=12, h=8,
          description='Normal: low dead tuple count (autovacuum keeping up). Watch for: monotonic growth on any table (vacuum falling behind). Check next: Table Scans for sequential scan impact.',
        ),

        // Table Scans (with dashed series for seq scans)
        layout.panel(
          panels.timeseriesPanel(
            title='Table Scans',
            targets=[
              targets.target('rate(pg_stat_user_tables_seq_scan_total[1m])', 'seq scans/s', 'A'),
              targets.target('rate(pg_stat_user_tables_idx_scan_total[1m])', 'idx scans/s', 'B'),
            ],
            unit='ops',
          ) + modifiers.withDashedSeries(['seq scans/s']),
          w=12, h=8,
          description='Normal: idx scans >> seq scans. Watch for: seq scan rate (dashed) approaching idx scan rate (missing index). Check next: Live Tuples for table growth.',
        ),

        // Live Tuples
        layout.panel(
          panels.timeseriesPanel(
            title='Live Tuples',
            targets=[targets.target(
              'sum(pg_stat_user_tables_n_live_tup) by (relname)',
              '{{relname}}',
              'A',
            )],
            unit='short',
          ),
          w=12, h=8,
          description='Normal: stable row counts proportional to game count. Watch for: rapid growth on single table (data accumulation). Check next: WAL for write amplification.',
        ),

        // WAL
        layout.panel(
          panels.timeseriesPanel(
            title='WAL Size',
            targets=[targets.target(
              'pg_wal_size_bytes',
              'WAL size',
              'A',
            )],
            unit='bytes',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: stable WAL size within configured retention. Watch for: WAL size growing unbounded (archiving or replication lag). Check next: Per-Table Cache Hit for buffer efficiency.',
        ),

        // Per-Table Cache Hit
        layout.panel(
          panels.timeseriesPanel(
            title='Per-Table Cache Hit',
            targets=[targets.target(
              'rate(pg_statio_user_tables_heap_blocks_hit_total[1m]) / clamp_min(rate(pg_statio_user_tables_heap_blocks_hit_total[1m]) + rate(pg_statio_user_tables_heap_blocks_read_total[1m]), 0.001) * 100',
              '{{relname}}',
              'A',
            )],
            unit='percent',
          ),
          w=24, h=8,
          description='Normal: > 99% per table. Watch for: individual tables dropping below 95% (hot table exceeding cache). Check next: Live Tuples for growth correlation.',
        ),
      ],

      // ── Collapsed: Query Performance (~1 panel) ──
      'Query Performance': [
        // Query Latency Heatmap — spanmetrics for DB transactions
        layout.panel(
          panels.heatmapPanel(
            title='Query Latency Heatmap',
            targets=[targets.heatmapTarget(
              'sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) by (le)' % [targets.spanmetricsMetric.duration, svc, targets.spans.db],
            )],
            unit='ms',
            colorScheme='Spectral',
            colorFill='dark-red',
          ),
          w=24, h=10,
          description='Normal: dense band at low latencies (< 10ms). Watch for: heat spreading to higher buckets over time. Check next: Postgres Internals for lock and vacuum state.',
        ),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // ACT — What's the evidence? (2 always-visible + 1 collapsed row)
    // ════════════════════════════════════════════════════════════════
    act=[
      // Slow Traces table (Tempo datasource)
      layout.panel(
        {
          title: 'Slow Traces (> 250ms)',
          type: 'table',
          datasource: targets.datasources.tempo,
          targets: [{
            refId: 'A',
            queryType: 'nativeSearch',
            serviceName: 'risk-it',
            minDuration: '250ms',
            limit: 20,
          }],
          transformations: [{
            id: 'sortBy',
            options: { fields: {}, sort: [{ field: 'Duration', desc: true }] },
          }],
          fieldConfig: {
            defaults: {
              custom: {
                align: 'auto',
                cellOptions: { type: 'auto' },
              },
              links: [{
                title: 'Investigate Trace',
                url: '/d/system-health/?var-traceId=${__value.raw}',
                targetBlank: false,
              }],
            },
            overrides: [],
          },
          options: {
            showHeader: true,
            footer: { show: false },
          },
        },
        w=24, h=10,
        description='Normal: few or no traces > 250ms. Watch for: recurring slow traces on same endpoint. Check next: click Trace ID to load waterfall below.',
      ),

      // WS Broadcast Logs
      layout.panel(
        panels.logPanel(
          title='WS Broadcast Logs',
          expr='{service_name="%s"} | scope_name=`go-risk-it` |~ "broadcast|ws"' % svc,
        ),
        w=24, h=8,
        description='Normal: broadcast log entries for each move. Watch for: error-level entries or timeouts. Check next: Trace Investigation collapsed row for correlated traces.',
      ),

      // Game Event Logs
      layout.panel(
        panels.logPanel(
          title='Game Event Logs',
          expr='{service_name="%s"} | json | eventType != "" | line_format "{{.eventType}} game={{.gameId}} {{.payload_action_type}} → {{.payload_to_phase}}"' % svc,
        ),
        w=24, h=8,
        description='Normal: event stream showing game lifecycle (moves, phase transitions, completions). Watch for: gaps in event flow or unexpected event types. Check next: Event Bus collapsed row in Orient for handler performance.',
      ),
    ],

    actDepth={
      // ── Collapsed: Trace Investigation ──
      'Trace Investigation': [
        // Trace Waterfall (Tempo)
        layout.panel(
          {
            title: 'Trace Waterfall',
            type: 'traces',
            datasource: targets.datasources.tempo,
            targets: [{
              refId: 'A',
              queryType: 'traceql',
              query: '${traceId}',
            }],
          },
          w=24, h=24,
          description='Normal: complete span tree visible. Watch for: missing spans (instrumentation gaps) or long gaps between spans (async contention). Check next: Correlated Logs below.',
        ),

        // Correlated Logs (Loki)
        layout.panel(
          panels.logPanel(
            title='Correlated Logs',
            expr='{service_name="%s"} | trace_id=`${traceId}`' % svc,
            showLabels=true,
          ),
          w=24, h=14,
          description='Normal: log lines matching the traced request. Watch for: error-level entries or unexpected log patterns. Check next: return to Observe tiles for current system state.',
        ),
      ],
    },
  ),
)
