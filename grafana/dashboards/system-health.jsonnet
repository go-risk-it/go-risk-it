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
  annotations={ list: [dashboard.perfTestAnnotation] },
  panels=layout.ooda(

    // ════════════════════════════════════════════════════════════════
    // OBSERVE — Am I OK? (7 panels: 4x stat bg + 3x stat)
    // ════════════════════════════════════════════════════════════════
    observe=[
      // Row 1: 4x w=6 stat panels with background color
      layout.panel(
        panels.statPanel(
          title='HTTP Latency p95',
          targets=[{
            expr: 'histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
            legendFormat: 'p95',
            refId: 'A',
          }],
          thresholds=thresholds.e2eP95,
          unit='s',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 500ms p95 (green). Watch for: sustained yellow (> 500ms) or red (> 1s). Check next: Lifecycle Timing in Orient for breakdown.',
      ),

      layout.panel(
        panels.statPanel(
          title='DB Transaction p95',
          targets=[{
            expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le))',
            legendFormat: 'p95',
            refId: 'A',
          }],
          thresholds=thresholds.dbTxnP95,
          unit='s',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 50ms p95 (green). Watch for: yellow (> 50ms) signals pool pressure, red (> 100ms) signals contention. Check next: Database collapsed row in Orient.',
      ),

      layout.panel(
        panels.statPanel(
          title='HTTP Error Rate %',
          targets=[{
            expr: 'sum(rate(http_server_requests_total{service_name="risk-it",http_status_code=~"[45].."}[1m])) / sum(rate(http_server_requests_total{service_name="risk-it"}[1m])) * 100',
            legendFormat: 'error %',
            refId: 'A',
          }],
          thresholds=thresholds.httpError,
          unit='percent',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 1% errors (green). Watch for: yellow (> 1%) may be transient; red (> 5%) indicates systematic failures. Check next: HTTP Error Rate timeseries in Server & HTTP.',
      ),

      layout.panel(
        panels.statPanel(
          title='Pool Utilization',
          targets=[{
            expr: 'db_pool_active{service_name="risk-it"} / db_pool_total{service_name="risk-it"} * 100',
            legendFormat: 'utilization',
            refId: 'A',
          }],
          thresholds=thresholds.poolUtil,
          unit='percent',
          colorMode='background',
        ) + modifiers.withLinks(observeLinks),
        w=6, h=6,
        description='Normal: < 60% utilization (green). Watch for: yellow (> 60%) under moderate load, red (> 80%) approaching saturation. Check next: Pool Usage timeseries in Database collapsed row.',
      ),

      // Row 2: 3x w=8 stat panels
      layout.panel(
        panels.statPanel(
          title='Canceled Acquires',
          targets=[{
            expr: 'rate(db_pool_canceled_acquires_total{service_name="risk-it"}[1m])',
            legendFormat: 'canceled/s',
            refId: 'A',
          }],
          thresholds=thresholds.canceledAcquires,
        ) + modifiers.withLinks(observeLinks),
        w=8, h=6,
        description='Normal: 0 canceled acquires (green). Watch for: any non-zero rate (red) means requests timing out waiting for DB connections. Check next: Pool Utilization and Acquire Time in Database row.',
      ),

      layout.panel(
        panels.statPanel(
          title='Cache Hit Rate',
          targets=[{
            expr: 'pg_stat_database_blks_hit_total{datname=~"postgres"} / (pg_stat_database_blks_hit_total{datname=~"postgres"} + pg_stat_database_blks_read_total{datname=~"postgres"}) * 100',
            legendFormat: 'hit %',
            refId: 'A',
          }],
          thresholds=thresholds.cacheHit,
          unit='percent',
        ) + modifiers.withLinks(observeLinks),
        w=8, h=6,
        description='Normal: > 99% cache hit rate (green). Watch for: yellow (< 99%) indicates working set exceeds shared_buffers; red (< 90%) severe miss rate. Check next: Postgres Internals collapsed row.',
      ),

      layout.panel(
        panels.statPanel(
          title='WS Connections',
          targets=[{
            expr: 'ws_connections_active{service_name="risk-it"}',
            legendFormat: 'active',
            refId: 'A',
          }],
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
      // Lifecycle Timing hero panel (w=24, h=10)
      layout.panel(
        panels.timeseriesPanel(
          title='Lifecycle Timing (p95)',
          targets=targets.lifecycleTargets,
          unit='s',
          overrides=targets.lifecycleOverrides,
        ),
        w=24, h=10,
        description='Normal: HTTP Total < 500ms, DB < 50ms, Game Logic < 200ms, WS < 100ms. Watch for: any boundary diverging from baseline. Check next: collapsed rows below for per-boundary detail.',
      ),

      // Percentile Bands panel (w=24, h=8)
      layout.panel(
        panels.percentileBandsPanel(
          title='HTTP Latency Percentile Bands',
          metric='http_server_request_duration_seconds_bucket',
          unit='s',
        ) + modifiers.withPercentileColors('http'),
        w=24, h=8,
        description='Normal: tight bands (p50-p95 close together). Watch for: p99 band widening (tail latency divergence). Check next: Server & HTTP collapsed row for per-route breakdown.',
      ),
    ],

    orientDepth={
      // ── Collapsed: Database (~9 panels) ──
      Database: [
        // Txn Duration Bands
        layout.panel(
          panels.percentileBandsPanel(
            title='DB Transaction Latency Bands',
            metric='db_transaction_duration_seconds_bucket',
            unit='s',
            exemplars=true,
          ) + modifiers.withPercentileColors('db') + modifiers.withSloThreshold(thresholds.dbTxnP95),
          w=12, h=8,
          description='Normal: p95 < 50ms, p99 < 100ms. Watch for: p99 diverging from p95 (long-tail transactions). Check next: Pool Usage for saturation.',
        ),

        // Pool Usage
        layout.panel(
          panels.timeseriesPanel(
            title='Pool Usage',
            targets=[
              { expr: 'db_pool_active{service_name="risk-it"}', legendFormat: 'Active', refId: 'A' },
              { expr: 'db_pool_idle{service_name="risk-it"}', legendFormat: 'Idle', refId: 'B' },
              { expr: 'db_pool_total{service_name="risk-it"}', legendFormat: 'Total', refId: 'C' },
            ],
            unit='short',
          ),
          w=12, h=8,
          description='Normal: active well below total, healthy idle buffer. Watch for: active approaching total (saturation). Check next: Acquire Time for queue pressure.',
        ),

        // Acquire Time
        layout.panel(
          panels.timeseriesPanel(
            title='Acquire Time P50/P95/P99',
            targets=targets.histogramQuantileTargets(
              'db_pool_acquire_duration_seconds_bucket',
              [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
            ),
            unit='s',
          ) + modifiers.withPercentileColors('db'),
          w=12, h=8,
          description='Normal: < 1ms acquire time. Watch for: p95 > 10ms (pool exhaustion starting). Check next: Canceled Acquires in Observe.',
        ),

        // Pool Saturation
        layout.panel(
          panels.timeseriesPanel(
            title='Pool Saturation',
            targets=[{
              expr: 'db_pool_active{service_name="risk-it"} / db_pool_total{service_name="risk-it"} * 100',
              legendFormat: 'utilization %',
              refId: 'A',
            }],
            unit='percent',
          ) + modifiers.withSloThreshold(thresholds.poolUtil),
          w=12, h=8,
          description='Normal: < 60% with SLO line visible. Watch for: crossing yellow threshold under normal load. Check next: Acquire Rate for demand side.',
        ),

        // Acquire Rate
        layout.panel(
          panels.timeseriesPanel(
            title='Acquire Rate',
            targets=[{
              expr: 'rate(db_pool_acquire_count_total{service_name="risk-it"}[1m])',
              legendFormat: 'acquires/s',
              refId: 'A',
            }],
            unit='ops',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: proportional to request rate. Watch for: spikes not correlated with traffic. Check next: Rollbacks for failed transactions.',
        ),

        // Rollbacks
        layout.panel(
          panels.timeseriesPanel(
            title='Rollbacks',
            targets=[{
              expr: 'rate(pg_stat_database_xact_rollback_total{datname=~"postgres"}[1m])',
              legendFormat: 'rollbacks/s',
              refId: 'A',
            }],
            unit='ops',
            color=colors.fixedColor(colors.errors),
          ),
          w=12, h=8,
          description='Normal: near zero rollback rate. Watch for: sustained rollbacks indicating transaction conflicts. Check next: PG Connections for connection-level issues.',
        ),

        // PG Connections
        layout.panel(
          panels.timeseriesPanel(
            title='PG Connections',
            targets=[
              { expr: 'pg_stat_database_numbackends{datname=~"postgres"}', legendFormat: 'backends', refId: 'A' },
            ],
            unit='short',
            color=colors.fixedColor(colors.db),
          ),
          w=8, h=8,
          description='Normal: stable connection count matching pool size. Watch for: gradual increase (connection leak). Check next: PG Txns/sec for throughput.',
        ),

        // PG Txns/sec
        layout.panel(
          panels.timeseriesPanel(
            title='PG Txns/sec',
            targets=[
              { expr: 'rate(pg_stat_database_xact_commit_total{datname=~"postgres"}[1m])', legendFormat: 'commits/s', refId: 'A' },
              { expr: 'rate(pg_stat_database_xact_rollback_total{datname=~"postgres"}[1m])', legendFormat: 'rollbacks/s', refId: 'B' },
            ],
            unit='ops',
          ),
          w=8, h=8,
          description='Normal: commits >> rollbacks. Watch for: rollback rate approaching commit rate. Check next: PG Rows for row-level throughput.',
        ),

        // PG Rows
        layout.panel(
          panels.timeseriesPanel(
            title='PG Rows',
            targets=[
              { expr: 'rate(pg_stat_database_tup_returned_total{datname=~"postgres"}[1m])', legendFormat: 'returned/s', refId: 'A' },
              { expr: 'rate(pg_stat_database_tup_fetched_total{datname=~"postgres"}[1m])', legendFormat: 'fetched/s', refId: 'B' },
              { expr: 'rate(pg_stat_database_tup_inserted_total{datname=~"postgres"}[1m])', legendFormat: 'inserted/s', refId: 'C' },
              { expr: 'rate(pg_stat_database_tup_updated_total{datname=~"postgres"}[1m])', legendFormat: 'updated/s', refId: 'D' },
            ],
            unit='ops',
          ),
          w=8, h=8,
          description='Normal: returned/fetched dominate (read-heavy workload). Watch for: updated/inserted spikes not correlated with game activity. Check next: Postgres Internals in Decide.',
        ),
      ],

      // ── Collapsed: Server & HTTP (~10 panels) ──
      'Server & HTTP': [
        // HTTP Request Rate
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Request Rate',
            targets=[{
              expr: 'sum(rate(http_server_requests_total{service_name="risk-it"}[1m]))',
              legendFormat: 'req/s',
              refId: 'A',
            }],
            unit='reqps',
            color=colors.fixedColor(colors.http),
          ),
          w=12, h=8,
          description='Normal: stable request rate matching expected load. Watch for: sudden drops (upstream issue) or spikes (load surge). Check next: HTTP Error Rate % for failure ratio.',
        ),

        // HTTP Error Rate % timeseries
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Error Rate %',
            targets=[{
              expr: 'sum(rate(http_server_requests_total{service_name="risk-it",http_status_code=~"[45].."}[1m])) / sum(rate(http_server_requests_total{service_name="risk-it"}[1m])) * 100',
              legendFormat: 'error %',
              refId: 'A',
            }],
            unit='percent',
            color=colors.fixedColor(colors.errors),
          ) + modifiers.withSloThreshold(thresholds.httpError),
          w=12, h=8,
          description='Normal: < 1% with SLO line visible. Watch for: error rate crossing SLO threshold. Check next: HTTP Latency P50/P95/P99 for latency correlation.',
        ),

        // HTTP Latency P50/P95/P99
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Latency P50/P95/P99',
            targets=targets.histogramQuantileTargetsWithExemplars(
              'http_server_request_duration_seconds_bucket',
              [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']],
            ),
            unit='s',
          ) + modifiers.withPercentileColors('http'),
          w=24, h=8,
          description='Normal: p50 < 100ms, p95 < 500ms, p99 < 1s. Watch for: p99 diverging sharply from p95 (long tail). Check next: HTTP Rate by Route for per-endpoint breakdown.',
        ),

        // HTTP Rate by Route
        layout.panel(
          panels.timeseriesPanel(
            title='HTTP Rate by Route',
            targets=[{
              expr: 'sum(rate(http_server_requests_total{service_name="risk-it"}[1m])) by (http_route)',
              legendFormat: '{{http_route}}',
              refId: 'A',
            }],
            unit='reqps',
          ),
          w=24, h=8,
          description='Normal: move endpoints dominate. Watch for: unusual routes or disproportionate rates. Check next: Game Engine dashboard for game-specific routes.',
        ),

        // Process CPU
        layout.panel(
          panels.timeseriesPanel(
            title='Process CPU',
            targets=[{
              expr: 'rate(process_cpu_seconds_total{service_name="risk-it"}[1m])',
              legendFormat: 'CPU cores',
              refId: 'A',
            }],
            unit='short',
            color=colors.fixedColor(colors.gameLogic),
          ),
          w=8, h=8,
          description='Normal: < 1 core under normal load. Watch for: sustained increase correlated with game count. Check next: Scheduler Latency for GC pressure.',
        ),

        // Scheduler Latency
        layout.panel(
          panels.timeseriesPanel(
            title='Scheduler Latency',
            targets=[{
              expr: 'histogram_quantile(0.95, sum(rate(runtime_go_sched_latency_bucket{service_name="risk-it"}[1m])) by (le))',
              legendFormat: 'p95',
              refId: 'A',
            }],
            unit='s',
          ),
          w=8, h=8,
          description='Normal: < 1ms scheduler latency. Watch for: spikes during GC pauses or high goroutine count. Check next: Goroutines for concurrency.',
        ),

        // Goroutines
        layout.panel(
          panels.timeseriesPanel(
            title='Goroutines',
            targets=[{
              expr: 'runtime_go_goroutines{service_name="risk-it"}',
              legendFormat: 'goroutines',
              refId: 'A',
            }],
            unit='short',
            color=colors.fixedColor(colors.gameLogic),
          ),
          w=8, h=8,
          description='Normal: stable count proportional to active games. Watch for: monotonic increase (goroutine leak). Check next: Heap Memory for memory pressure.',
        ),

        // Heap Memory
        layout.panel(
          panels.timeseriesPanel(
            title='Heap Memory',
            targets=[
              { expr: 'runtime_go_mem_heap_alloc{service_name="risk-it"}', legendFormat: 'Alloc', refId: 'A' },
              { expr: 'runtime_go_mem_heap_sys{service_name="risk-it"}', legendFormat: 'Sys', refId: 'B' },
            ],
            unit='bytes',
          ),
          w=8, h=8,
          description='Normal: alloc well below sys, sawtooth GC pattern. Watch for: alloc approaching sys (heap pressure). Check next: GC Goal for tuning.',
        ),

      ],

      // ── Collapsed: WebSocket (~4 panels) ──
      WebSocket: [
        // Broadcast Latency Bands
        layout.panel(
          panels.percentileBandsPanel(
            title='WS Broadcast Latency Bands',
            metric='ws_broadcast_duration_seconds_bucket',
            unit='s',
            exemplars=true,
          ) + modifiers.withPercentileColors('ws'),
          w=12, h=8,
          description='Normal: p95 < 100ms. Watch for: p99 > 500ms (slow consumers blocking broadcast). Check next: Messages Rate for throughput.',
        ),

        // Messages Rate
        layout.panel(
          panels.timeseriesPanel(
            title='Messages Rate',
            targets=[{
              expr: 'sum(rate(ws_messages_total{service_name="risk-it"}[1m])) by (direction)',
              legendFormat: '{{direction}}',
              refId: 'A',
            }],
            unit='ops',
          ),
          w=12, h=8,
          description='Normal: outbound >> inbound (server pushes state). Watch for: inbound spikes (clients retrying). Check next: Fan-Out for amplification ratio.',
        ),

        // WS Errors
        layout.panel(
          panels.timeseriesPanel(
            title='WS Errors',
            targets=[{
              expr: 'sum(rate(ws_errors_total{service_name="risk-it"}[1m]))',
              legendFormat: 'errors/s',
              refId: 'A',
            }],
            unit='ops',
            color=colors.fixedColor(colors.errors),
          ),
          w=12, h=8,
          description='Normal: 0 errors. Watch for: any sustained error rate (broken connections, write failures). Check next: WS Connections in Observe for connection count.',
        ),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // DECIDE — Where's the bottleneck? (3 always-visible + 2 collapsed rows)
    // ════════════════════════════════════════════════════════════════
    decide=[
      // Fan-out Amplification
      layout.panel(
        panels.timeseriesPanel(
          title='Fan-out Amplification',
          targets=[{
            expr: 'sum(rate(ws_messages_total{service_name="risk-it",direction="outbound"}[1m])) / sum(rate(event_bus_events_total{service_name="risk-it",event_type="move_executed"}[1m]))',
            legendFormat: 'msgs/move',
            refId: 'A',
          }],
          unit='short',
          color=colors.fixedColor(colors.ws),
        ) + modifiers.withSloThreshold(thresholds.ccFanOut),
        w=8, h=8,
        description='Normal: ~4x (one per player). Watch for: > 10x (SLO yellow) suggests broadcast storms. Check next: WebSocket collapsed row for per-message detail.',
      ),

      // DB Latency Share
      layout.panel(
        panels.timeseriesPanel(
          title='DB Latency Share',
          targets=[{
            expr: 'histogram_quantile(0.95, sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)) / histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)) * 100',
            legendFormat: 'DB % of HTTP p95',
            refId: 'A',
          }],
          unit='percent',
          color=colors.fixedColor(colors.db),
        ) + modifiers.withSloThreshold(thresholds.ccDbLatencyShare),
        w=8, h=8,
        description='Normal: 30-50% (DB is a fraction of total latency). Watch for: > 70% (SLO yellow) means DB dominates request time. Check next: Postgres Internals collapsed row.',
      ),

      // Process Memory
      layout.panel(
        panels.timeseriesPanel(
          title='Process Memory',
          targets=[
            { expr: 'process_resident_memory_bytes{service_name="risk-it"}', legendFormat: 'RSS', refId: 'A' },
            { expr: 'runtime_go_mem_heap_alloc{service_name="risk-it"}', legendFormat: 'Heap Alloc', refId: 'B' },
          ],
          unit='bytes',
        ),
        w=8, h=8,
        description='Normal: RSS tracks heap alloc with stable overhead. Watch for: RSS growing while heap is flat (off-heap leak). Check next: Goroutines in Server & HTTP row.',
      ),
    ],

    decideDepth={
      // ── Collapsed: Postgres Internals (~7 panels) ──
      'Postgres Internals': [
        // SHOWCASE: Lock Contention (stacked area with lock mode colors)
        layout.panel(
          panels.timeseriesPanel(
            title='Lock Contention',
            targets=[{
              expr: 'sum(pg_locks_count) by (mode)',
              legendFormat: '{{mode}}',
              refId: 'A',
            }],
            unit='short',
          ) + modifiers.withStackedArea(30, 'hue') + modifiers.withSeriesColors(colors.lockModes),
          w=24, h=10,
          description='Normal: AccessShareLock dominates (read locks, green). Watch for: ExclusiveLock or AccessExclusiveLock spikes (red/orange — DDL or heavy writes). Check next: Deadlocks for escalated contention.',
        ),

        // Deadlocks
        layout.panel(
          panels.timeseriesPanel(
            title='Deadlocks',
            targets=[{
              expr: 'rate(pg_stat_database_deadlocks_total{datname=~"postgres"}[1m])',
              legendFormat: 'deadlocks/s',
              refId: 'A',
            }],
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
            targets=[{
              expr: 'sum(pg_stat_user_tables_n_dead_tup) by (relname)',
              legendFormat: '{{relname}}',
              refId: 'A',
            }],
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
              { expr: 'rate(pg_stat_user_tables_seq_scan_total[1m])', legendFormat: 'seq scans/s', refId: 'A' },
              { expr: 'rate(pg_stat_user_tables_idx_scan_total[1m])', legendFormat: 'idx scans/s', refId: 'B' },
            ],
            unit='ops',
          ) + modifiers.withDashedSeries(['seq scans/s']),
          w=12, h=8,
          description='Normal: idx scans >> seq scans. Watch for: seq scan rate (dashed) approaching idx scan rate (missing index). Check next: Table Sizes for table growth.',
        ),

        // Table Sizes
        layout.panel(
          panels.timeseriesPanel(
            title='Table Sizes',
            targets=[{
              expr: 'sum(pg_stat_user_tables_size) by (relname)',
              legendFormat: '{{relname}}',
              refId: 'A',
            }],
            unit='bytes',
          ),
          w=12, h=8,
          description='Normal: stable sizes proportional to game count. Watch for: rapid growth on single table (data accumulation). Check next: WAL for write amplification.',
        ),

        // WAL
        layout.panel(
          panels.timeseriesPanel(
            title='WAL',
            targets=[{
              expr: 'rate(pg_stat_wal_wal_bytes_total[1m])',
              legendFormat: 'WAL bytes/s',
              refId: 'A',
            }],
            unit='Bps',
            color=colors.fixedColor(colors.db),
          ),
          w=12, h=8,
          description='Normal: proportional to write throughput. Watch for: WAL spikes not correlated with game activity. Check next: Per-Table Cache Hit for buffer efficiency.',
        ),

        // Per-Table Cache Hit
        layout.panel(
          panels.timeseriesPanel(
            title='Per-Table Cache Hit',
            targets=[{
              expr: 'pg_stat_user_tables_heap_blks_hit_total / (pg_stat_user_tables_heap_blks_hit_total + pg_stat_user_tables_heap_blks_read_total) * 100',
              legendFormat: '{{relname}}',
              refId: 'A',
            }],
            unit='percent',
          ),
          w=24, h=8,
          description='Normal: > 99% per table. Watch for: individual tables dropping below 95% (hot table exceeding cache). Check next: Table Sizes for growth correlation.',
        ),
      ],

      // ── Collapsed: Query Performance (~3 panels) ──
      'Query Performance': [
        // Query Latency Heatmap
        layout.panel(
          panels.heatmapPanel(
            title='Query Latency Heatmap',
            targets=[{
              expr: 'sum(rate(db_transaction_duration_seconds_bucket{service_name="risk-it"}[1m])) by (le)',
              format: 'heatmap',
              legendFormat: '{{le}}',
              refId: 'A',
            }],
            unit='s',
            colorScheme='Spectral',
            colorFill='dark-red',
          ),
          w=24, h=10,
          description='Normal: dense band at low latencies (< 10ms). Watch for: heat spreading to higher buckets over time. Check next: Top Queries by Call Rate for hottest queries.',
        ),

        // Top Queries by Call Rate
        layout.panel(
          panels.timeseriesPanel(
            title='Top Queries by Call Rate',
            targets=[{
              expr: 'topk(10, sum(rate(pg_stat_statements_calls_total[1m])) by (query))',
              legendFormat: '{{query}}',
              refId: 'A',
            }],
            unit='ops',
          ),
          w=12, h=8,
          description='Normal: game state queries dominate. Watch for: unexpected queries in top 10. Check next: Top Queries by Latency for slow query identification.',
        ),

        // Top Queries by Latency
        layout.panel(
          panels.timeseriesPanel(
            title='Top Queries by Latency',
            targets=[{
              expr: 'topk(10, sum(rate(pg_stat_statements_total_exec_time_total[1m])) by (query) / sum(rate(pg_stat_statements_calls_total[1m])) by (query))',
              legendFormat: '{{query}}',
              refId: 'A',
            }],
            unit='s',
          ),
          w=12, h=8,
          description='Normal: all queries < 10ms avg. Watch for: any query > 50ms avg (missing index or lock contention). Check next: Postgres Internals for lock and vacuum state.',
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
          datasource: { type: 'tempo', uid: 'tempo' },
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
          expr='{service_name="risk-it"} |= "broadcast" |= "ws"',
        ),
        w=24, h=8,
        description='Normal: broadcast log entries for each move. Watch for: error-level entries or timeouts. Check next: Trace Investigation collapsed row for correlated traces.',
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
            datasource: { type: 'tempo', uid: 'tempo' },
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
            expr='{service_name="risk-it"} | trace_id=`${traceId}`',
            showLabels=true,
          ),
          w=24, h=14,
          description='Normal: log lines matching the traced request. Watch for: error-level entries or unexpected log patterns. Check next: return to Observe tiles for current system state.',
        ),
      ],
    },
  ),
)
