// Game Theater dashboard — single-game deep dive.
// Source of truth: grafana/dashboards/game-theater.jsonnet
// Regenerate: make dashboards
local colors = import 'colors.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local layout = import 'layout.libsonnet';
local links = import 'links.libsonnet';
local modifiers = import 'modifiers.libsonnet';
local panels = import 'panels.libsonnet';
local queries = import 'queries.libsonnet';
local targets = import 'targets.libsonnet';

local svc = targets.serviceName;

// ── Template variables ──

local gameIdVar = {
  name: 'gameId',
  type: 'textbox',
  label: 'Game ID',
  current: { value: '' },
};

local traceIdVar = {
  name: 'traceId',
  type: 'textbox',
  label: 'Trace ID',
  current: { value: '' },
};

// ── Phase color map for State Timeline ──

// ── Phase regex color mappings for State Timeline ──
// Turn-suffixed values (DEPLOY_0, ATTACK_4) use regex to match ^PHASE.* patterns.
// Display text strips the suffix. This creates visual turn boundaries with mergeValues:true.
local phaseRegexMappings = [
  { type: 'regex', options: { pattern: '/^DEPLOY/', result: { text: 'DEPLOY', color: colors.phase.deploy, index: 0 } } },
  { type: 'regex', options: { pattern: '/^ATTACK/', result: { text: 'ATTACK', color: colors.phase.attack, index: 1 } } },
  { type: 'regex', options: { pattern: '/^CARDS/', result: { text: 'CARDS', color: colors.phase.cards, index: 2 } } },
  { type: 'regex', options: { pattern: '/^CONQUER/', result: { text: 'CONQUER', color: colors.phase.conquer, index: 3 } } },
  { type: 'regex', options: { pattern: '/^REINFORCE/', result: { text: 'REINFORCE', color: colors.phase.reinforce, index: 4 } } },
  { type: 'regex', options: { pattern: '/^WAITING/', result: { text: ' ', color: '#111217', index: 5 } } },
];

// ── Reusable Loki query fragments ──
// OTLP log format: body = "game_event", slog attributes = Loki labels.
// Use |= "game_event" for body match, | key=`value` for label filters.

// Base event stream filtered to this game.
local gameEventStream =
  queries.lokiEventStream(svc, '$gameId');

// Phase transition events for this game.
local phaseTransitionStream =
  queries.lokiEventStream(svc, '$gameId', queries.eventTypes.phaseTransitioned);

// Move events for this game.
local moveEventStream =
  queries.lokiEventStream(svc, '$gameId', queries.eventTypes.moveExecuted);

// Attack move events for this game (pipeline filter, not stream selector — Loki 3.6 structured metadata).
local attackMoveStream =
  queries.lokiEventStream(svc, '$gameId', queries.eventTypes.moveExecuted) + ' | payload_action_type=`ATTACK`';

// Headline events for this game.
local headlineStream =
  queries.lokiEventStream(svc, '$gameId', queries.headlineFilter);

// Game completed events for this game.
local gameCompletedStream =
  queries.lokiEventStream(svc, '$gameId', queries.eventTypes.gameCompleted);

// ── Cross-dashboard data links ──

local engineLink = links.toDashboard('Compare in Game Engine', links.dashboardUids.gameEngine);
local healthLink = links.toDashboard('System Health', links.dashboardUids.systemHealth);

// Links applied to Observe stat tiles (fleet-scoped context navigation).
local observeStatLinks = [healthLink, engineLink];

// ── Deviation stat tile thresholds (fleet-calibrated, absolute) ──
// Yellow at fleet p84, red at fleet p97.5. These values are derived from
// game.summary.* histograms collected across completed games. Calibrate
// by running a few dozen games and checking the recording rules:
//   game:summary_moves:p84, game:summary_moves:p975, etc.
// Initial estimates based on 4-player Risk game characteristics.

local gameDurationThresholds = {
  mode: 'absolute',
  steps: [
    { color: 'yellow', value: null },
    { color: 'green', value: 1 },
  ],
};

local moveCountThresholds = {
  mode: 'absolute',
  steps: [
    { color: 'green', value: null },
    { color: 'yellow', value: 150 },
    { color: 'red', value: 400 },
  ],
};

local attackRateThresholds = {
  mode: 'absolute',
  steps: [
    { color: 'green', value: null },
    { color: 'yellow', value: 75 },
    { color: 'red', value: 200 },
  ],
};

local headlineCountThresholds = {
  mode: 'absolute',
  steps: [
    { color: 'green', value: null },
    { color: 'yellow', value: 8 },
    { color: 'red', value: 20 },
  ],
};

local moveLatencyThresholds = {
  mode: 'absolute',
  steps: [
    { color: 'green', value: null },
    { color: 'yellow', value: 200 },
    { color: 'red', value: 500 },
  ],
};

local turnCountThresholds = {
  mode: 'absolute',
  steps: [
    { color: 'green', value: null },
    { color: 'yellow', value: 40 },
    { color: 'red', value: 100 },
  ],
};

// ── Panel definitions (each <= 30 lines) ──

// Row 0: Game Duration — presence of game_completed event (1=finished, 0=in progress).
local gameDuration =
  panels.statPanel(
    title='Game Completed',
    targets=[targets.lokiTarget(
      queries.lokiCountOverTime(gameCompletedStream),
      'completed',
    )],
    thresholds=gameDurationThresholds,
    colorMode='background',
  ) + { datasource: targets.datasources.loki } + modifiers.withLinks(observeStatLinks);

// Row 0: Move Count — total moves for this game.
local moveCount =
  panels.statPanel(
    title='Move Count',
    targets=[targets.lokiTarget(
      queries.lokiCountOverTime(moveEventStream),
      'moves',
    )],
    thresholds=moveCountThresholds,
    colorMode='background',
  ) + { datasource: targets.datasources.loki } + modifiers.withLinks(observeStatLinks);

// Row 0: Attack Count — attack moves for this game.
local attackCount =
  panels.statPanel(
    title='Attack Count',
    targets=[targets.lokiTarget(
      queries.lokiCountOverTime(attackMoveStream),
      'attacks',
    )],
    thresholds=attackRateThresholds,
    colorMode='background',
  ) + { datasource: targets.datasources.loki } + modifiers.withLinks(observeStatLinks);

// Row 0: Headline Count — notable game moments (eliminations, continent events).
local headlineCount =
  panels.statPanel(
    title='Headline Count',
    targets=[targets.lokiTarget(
      queries.lokiCountOverTime(headlineStream),
      'headlines',
    )],
    thresholds=headlineCountThresholds,
    colorMode='background',
  ) + { datasource: targets.datasources.loki } + modifiers.withLinks(observeStatLinks);

// Row 0: Avg Move Latency — fleet-wide orchestrate_move latency (Prometheus).
local avgMoveLatency =
  panels.statPanel(
    title='Avg Move Latency',
    targets=[targets.target(
      'avg(rate(traces_span_metrics_duration_milliseconds_sum{service_name="%s", span_name="game.orchestrate_move"}[$__rate_interval])) / avg(rate(traces_span_metrics_duration_milliseconds_count{service_name="%s", span_name="game.orchestrate_move"}[$__rate_interval]))' % [svc, svc],
      'avg ms',
    )],
    thresholds=moveLatencyThresholds,
    unit='ms',
    colorMode='background',
  ) + modifiers.withLinks(observeStatLinks);

// Row 0: Turn Count — phase transitions for this game.
local turnCount =
  panels.statPanel(
    title='Turn Count',
    targets=[targets.lokiTarget(
      queries.lokiCountOverTime(phaseTransitionStream),
      'phase transitions',
    )],
    thresholds=turnCountThresholds,
    colorMode='background',
  ) + { datasource: targets.datasources.loki } + modifiers.withLinks(observeStatLinks);

// Shared transformation: extract Loki structured metadata fields from labels.
// OTLP logs promote slog attributes to labels — extractFields with source="labels"
// makes them available as named columns for table/timeline panels.
local extractFieldsTransform = {
  id: 'extractFields',
  options: { source: 'labels', format: 'auto' },
};

// Orient: Phase Tapestry — State Timeline showing per-player phase progression.
// Uses move_executed + phase_transitioned + turn_ended events with turn-suffixed
// values (DEPLOY_0, ATTACK_4) so mergeValues:true creates visual turn boundaries.
// phase_transitioned events make auto-advanced phases visible (e.g. CARDS skipped
// when player has no cards). Uses from_phase → action_type to show the departing
// phase, which merges correctly with MoveExecuted's action_type.
// Regex value mappings strip the turn suffix: ^DEPLOY.* → blue, display "DEPLOY".
local phaseTapestry =
  panels.stateTimelinePanel(
    title='Phase Tapestry',
    targets=[targets.lokiTarget(
      queries.lokiStateTimeline(svc, '$gameId', 'move_executed|phase_transitioned|turn_ended', 'payload_user_id', 'payload_action_type'),
      '{{payload_user_id}}',
    )],
  ) + {
    datasource: targets.datasources.loki,
    transformations: [
      extractFieldsTransform,
      {
        id: 'organize',
        options: {
          excludeByName: { tsNs: true, id: true, labels: true, labelTypes: true, trace_id: true },
          renameByName: { Line: '' },
        },
      },
      { id: 'partitionByValues', options: { fields: ['payload_user_id'], keepFields: false } },
    ],
    // Override fieldConfig completely to avoid the stateTimelinePanel's color:{mode:'fixed'}
    fieldConfig: {
      defaults: {
        custom: { fillOpacity: 80, lineWidth: 0 },
        links: [engineLink],
        mappings: phaseRegexMappings,
      },
      overrides: [],
    },
    options+: { mergeValues: true, rowHeight: 0.85 },
  };

// Headline Moments transformations: extract labels, organize columns, sort by time.
local headlineMomentsTransforms = [
  extractFieldsTransform,
  {
    id: 'organize',
    options: {
      excludeByName: {
        id: true, tsNs: true, labels: true, labelTypes: true, Line: true,
        detected_level: true, flags: true, observed_timestamp: true,
        scope_name: true, service_name: true, service_version: true,
        severity_number: true, severity_text: true,
        telemetry_sdk_language: true, telemetry_sdk_name: true, telemetry_sdk_version: true,
        span_id: true, game_id: true, gameId: true,
        eventTimestamp: true, payload_timestamp: true, payload_game_id: true,
      },
      renameByName: {
        Time: 'Time',
        eventType: 'Event Type',
        payload_user_id: 'Player',
        payload_event_type: 'Detail',
        trace_id: 'Trace ID',
      },
    },
  },
  { id: 'sortBy', options: { fields: {}, sort: [{ field: 'Time', desc: false }] } },
];

// Headline event type color overrides for table cells.
local headlineEventTypeOverride = {
  matcher: { id: 'byName', options: 'Event Type' },
  properties: [{
    id: 'mappings',
    value: [
      { type: 'value', options: {
        player_eliminated: { text: 'Player Eliminated', color: colors.headline.elimination },
        continent_captured: { text: 'Continent Captured', color: colors.headline.continent },
        continent_lost: { text: 'Continent Lost', color: colors.headline.continent },
        game_completed: { text: 'Game Completed', color: colors.headline.victory },
      } },
    ],
  }, {
    id: 'custom.cellOptions',
    value: { type: 'color-text' },
  }],
};

// Decide: Headline Moments — table of notable game events.
local headlineMoments = {
  title: 'Headline Moments',
  type: 'table',
  datasource: targets.datasources.loki,
  targets: [{
    refId: 'A',
    expr: queries.lokiEventStream(
      svc, '$gameId', '%s|%s' % [queries.headlineFilter, queries.eventTypes.gameCompleted]
    ),
  }],
  transformations: headlineMomentsTransforms,
  fieldConfig: {
    defaults: { custom: { align: 'auto', cellOptions: { type: 'auto' } } },
    overrides: [headlineEventTypeOverride],
  },
  options: { showHeader: true, footer: { show: false } },
};

// Event Chronicle transformations: extract labels, organize columns, sort by time.
local chronicleTransforms = [
  extractFieldsTransform,
  {
    id: 'organize',
    options: {
      excludeByName: {
        id: true, tsNs: true, labels: true, labelTypes: true, Line: true,
        detected_level: true, flags: true, observed_timestamp: true,
        scope_name: true, service_name: true, service_version: true,
        severity_number: true, severity_text: true,
        telemetry_sdk_language: true, telemetry_sdk_name: true, telemetry_sdk_version: true,
        span_id: true, game_id: true, gameId: true,
        eventTimestamp: true, payload_timestamp: true, payload_game_id: true,
        payload_game_over: true, payload_move_log_id: true,
      },
      renameByName: {
        Time: 'Time',
        eventType: 'Event Type',
        payload_user_id: 'Player',
        payload_turn: 'Turn',
        payload_action_type: 'Action',
        payload_to_phase: 'Phase',
        payload_from_phase: 'From Phase',
        payload_target_phase: 'Target Phase',
        payload_event_type: 'Event Detail',
        trace_id: 'Trace ID',
      },
    },
  },
  { id: 'sortBy', options: { fields: {}, sort: [{ field: 'Time', desc: false }] } },
];

// AC-9: trace_id column data link populates $traceId variable.
local traceIdLink = {
  matcher: { id: 'byName', options: 'trace_id' },
  properties: [{
    id: 'links',
    value: [{
      title: 'Investigate Trace',
      url: '/d/%s/?var-traceId=${__value.raw}&var-gameId=${gameId}&from=${__from}&to=${__to}' % links.dashboardUids.gameTheater,
      targetBlank: false,
    }],
  }],
};

// Decide: Event Chronicle — Loki table showing all game events.
local eventChronicle = {
  title: 'Event Chronicle',
  type: 'table',
  datasource: targets.datasources.loki,
  targets: [{ refId: 'A', expr: gameEventStream }],
  transformations: chronicleTransforms,
  fieldConfig: {
    defaults: { custom: { align: 'auto', cellOptions: { type: 'auto' } } },
    overrides: [traceIdLink],
  },
  options: { showHeader: true, footer: { show: false } },
};

// Act: Move Forensics — trace waterfall via $traceId.
local moveForensics =
  panels.tracesPanel(
    title='Move Forensics',
    query='${traceId}',
  );

// Act: Correlated Logs — Loki logs filtered by trace ID.
local correlatedLogs =
  panels.logPanel(
    title='Correlated Logs',
    expr='{service_name="%s"} | trace_id=`${traceId}` | json | line_format "{{if .eventType}}{{.eventType}} game={{.gameId}} {{.payload_action_type}} → {{.payload_to_phase}}{{else}}{{.__line__}}{{end}}"' % svc,
    showLabels=true,
  );

// Orient depth: Move Latency Distribution histogram (collapsed sub-row).
local moveLatencyHisto =
  panels.heatmapPanel(
    title='Move Latency Distribution',
    targets=[targets.heatmapTarget(
      'sum(rate(%s{service_name="%s", span_name=~"%s"}[1m])) by (le)' % [
        targets.spanmetricsMetric.duration,
        svc,
        targets.spans.gameLogic,
      ],
    )],
    unit='ms',
    colorScheme='Oranges',
    colorFill='dark-orange',
  );

// ── Fleet distribution panels (collapsed Observe sub-row) ──
// Show p50/p75/p95/p99 of game.summary.* histograms across all completed games.
// These panels provide fleet context for interpreting the per-game stat tiles above.
//
// LIMITATION (AC7): Grafana 11 cannot compute z-scores across Loki and Prometheus
// queries in a single stat panel. Per-game counts come from Loki (game event logs),
// while fleet distributions come from Prometheus (OTel histograms). The stat tiles
// show raw per-game values; the fleet distribution panels below show where "normal"
// falls. Operators compare visually. Z-score stat tiles become feasible when either:
//   (a) Grafana adds cross-datasource math transforms, or
//   (b) per-game summary data is also written to Prometheus (e.g., via recording rules
//       on game_completed events with game_id labels — would require exemplar-style
//       per-game histogram recording which the current aggregated histograms don't support).

local fleetQuantiles = [['0.5', 'p50'], ['0.75', 'p75'], ['0.95', 'p95'], ['0.99', 'p99']];

local fleetMovesPanel =
  panels.timeseriesPanel(
    title='Fleet Move Distribution',
    targets=targets.histogramQuantileTargetsWithExemplars(
      'game_summary_moves_bucket', fleetQuantiles
    ),
    unit='short',
  );

local fleetAttacksPanel =
  panels.timeseriesPanel(
    title='Fleet Attack Distribution',
    targets=targets.histogramQuantileTargetsWithExemplars(
      'game_summary_attacks_bucket', fleetQuantiles
    ),
    unit='short',
  );

local fleetTurnsPanel =
  panels.timeseriesPanel(
    title='Fleet Turn Distribution',
    targets=targets.histogramQuantileTargetsWithExemplars(
      'game_summary_turns_bucket', fleetQuantiles
    ),
    unit='short',
  );

local fleetHeadlinesPanel =
  panels.timeseriesPanel(
    title='Fleet Headline Distribution',
    targets=targets.histogramQuantileTargetsWithExemplars(
      'game_summary_headlines_bucket', fleetQuantiles
    ),
    unit='short',
  );

// ── Dashboard assembly ──

dashboard.new(
  uid='game-theater',
  title='Game Theater',
  description='Single-game deep dive: phase tapestry, event chronicle, and trace forensics',
  tags=['risk-it', 'game-theater'],
  refresh='',
  time={ from: 'now-1h', to: 'now' },
  templating={ list: [gameIdVar, traceIdVar] },
  panels=layout.ooda(

    // ════════════════════════════════════════════════════════════════
    // OBSERVE — Am I OK? (6 deviation stat tiles, w=4 each)
    // Game-level counters with background color thresholds.
    // Fleet-calibrated: yellow at ~p84, red at ~p97.5.
    // ════════════════════════════════════════════════════════════════
    observe=[
      layout.panel(
        gameDuration,
        w=4, h=6,
        description='Normal: 1 (green) means game completed. Watch for: 0 (yellow) on a game that should have finished. Check next: Event Chronicle for last event.',
      ),
      layout.panel(
        moveCount,
        w=4, h=6,
        description='Normal: 20-150 total moves (green). Watch for: > 150 (yellow) above fleet p84; > 400 (red) above fleet p97.5. Check next: Fleet Move Distribution for percentiles.',
      ),
      layout.panel(
        attackCount,
        w=4, h=6,
        description='Normal: attack moves proportional to total moves. Watch for: > 75 (yellow) above fleet p84. Check next: Fleet Attack Distribution for percentiles.',
      ),
      layout.panel(
        headlineCount,
        w=4, h=6,
        description='Normal: a few eliminations and continent events per game. Watch for: > 8 (yellow) above fleet p84; > 20 (red) above fleet p97.5. Check next: Fleet Headline Distribution.',
      ),
      layout.panel(
        avgMoveLatency,
        w=4, h=6,
        description='Normal: < 200ms fleet average (green). Watch for: > 200ms (yellow) degraded performance; > 500ms (red) severe. Check next: Move Forensics trace for slow move detail.',
      ),
      layout.panel(
        turnCount,
        w=4, h=6,
        description='Normal: phase transitions proportional to player count and game length. Watch for: > 40 (yellow) above fleet p84; > 100 (red) above fleet p97.5. Check next: Fleet Turn Distribution.',
      ),
    ],

    observeDepth={
      'Fleet Game Summary Distribution': [
        layout.panel(
          fleetMovesPanel,
          w=12, h=8,
          description='Normal: p50 moves 30-80, p95 under 200. Watch for: percentile curves trending upward (games getting longer). Check next: compare with this game\'s Move Count stat tile above.',
        ),
        layout.panel(
          fleetAttacksPanel,
          w=12, h=8,
          description='Normal: p50 attacks 15-40, p95 under 100. Watch for: attack percentiles growing faster than move percentiles (more combat-heavy meta). Check next: compare with Attack Count tile.',
        ),
        layout.panel(
          fleetTurnsPanel,
          w=12, h=8,
          description='Normal: p50 turns 10-30, p95 under 80. Watch for: rising turn counts suggest games taking more rounds. Check next: compare with Turn Count tile above.',
        ),
        layout.panel(
          fleetHeadlinesPanel,
          w=12, h=8,
          description='Normal: p50 headlines 2-5, p95 under 15. Watch for: unusually flat distribution (all games similar) or extreme outliers. Check next: compare with Headline Count tile.',
        ),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // ORIENT — What's the shape? (Phase Tapestry + collapsed histogram)
    // ════════════════════════════════════════════════════════════════
    orient=[
      layout.panel(
        phaseTapestry,
        w=24, h=10,
        description='Normal: colored markers per player at each phase transition (DEPLOY=blue, ATTACK=red, CARDS=gold, CONQUER=green, REINFORCE=purple). Gaps = idle (not their turn). Dense markers = active turn. Watch for: long gaps (eliminated player) or stuck phases. Check next: Event Chronicle for detailed event sequence.',
      ),
    ],

    orientDepth={
      'Move Latency Distribution': [
        layout.panel(
          moveLatencyHisto,
          w=24, h=8,
          description='Normal: dense band in low-millisecond range (fast moves). Watch for: heat spreading to higher buckets over time. Check next: Move Forensics for individual slow traces.',
        ),
      ],
    },

    // ════════════════════════════════════════════════════════════════
    // DECIDE — Where's the bottleneck? (Headline Moments + Event Chronicle)
    // ════════════════════════════════════════════════════════════════
    decide=[
      layout.panel(
        headlineMoments,
        w=24, h=8,
        description='Normal: player eliminations and continent captures during gameplay. Watch for: rapid consecutive eliminations (one player dominating) or no headlines (passive game). Check next: Event Chronicle for full event log.',
      ),
      layout.panel(
        eventChronicle,
        w=24, h=12,
        description='Normal: chronological event log for this game. Watch for: gaps in event sequence or repeated errors. Check next: click a trace ID to load Move Forensics below.',
      ),
    ],

    // ════════════════════════════════════════════════════════════════
    // ACT — What's the evidence? (Move Forensics + Correlated Logs)
    // ════════════════════════════════════════════════════════════════
    act=[
      layout.panel(
        moveForensics + modifiers.withLinks([healthLink]),
        w=24, h=20,
        description='Normal: complete span tree for the selected trace. Watch for: missing spans (instrumentation gaps) or long gaps between spans (contention). Check next: Correlated Logs below.',
      ),
      layout.panel(
        correlatedLogs,
        w=24, h=10,
        description='Normal: log lines matching the traced request. Watch for: error-level entries or unexpected patterns. Check next: return to Phase Tapestry for game-level context.',
      ),
    ],
  ),
)
