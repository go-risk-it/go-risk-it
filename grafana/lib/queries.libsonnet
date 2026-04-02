// Composable LogQL query helpers for go-risk-it Grafana dashboards.
// All functions return STRINGS (LogQL expressions), not target objects.
// Callers wrap with targets.lokiTarget(expr=...) or pass to panels.logPanel(expr=...).
//
// OTLP log format: The Go backend sends logs via slog → otelslog → OTLP → Loki.
// The log body is the plain-text slog message (e.g. "game_event").
// ALL slog attributes (eventType, gameId, payload_user_id, etc.) are promoted to
// Loki stream labels by the OTLP ingest pipeline. This means:
//   - Line filters (|= "text") match against the body text only
//   - Label filters (| key="value") match against slog attributes
//   - | json is NOT useful (body is not JSON)
//   - Fields are directly accessible as labels — no extraction needed
//
// No imports. Zero dependencies. Callers pass targets.serviceName as the service parameter.
{
  // ── Event type constants ──
  // Matches the event type string constants from internal/game/events/types.go
  // and internal/game/headlines/types.go.
  eventTypes:: {
    moveExecuted: 'move_executed',
    phaseTransitioned: 'phase_transitioned',
    gameCompleted: 'game_completed',
    gameCreated: 'game_created',
    playerConnected: 'player_connected',
    playerEliminated: 'player_eliminated',
    continentCaptured: 'continent_captured',
    continentLost: 'continent_lost',
    turnEnded: 'turn_ended',
  },

  // Pre-built regex alternation of headline event types.
  // Use with lokiEventStream's eventTypeFilter as a regex: |~ "pattern"
  headlineFilter:: '%s|%s|%s' % [
    $.eventTypes.playerEliminated,
    $.eventTypes.continentCaptured,
    $.eventTypes.continentLost,
  ],

  // ── Composable query helpers ──

  // Build a LogQL stream selector string.
  // service: string (e.g. 'risk-it').
  // extraLabels: string (optional) — additional label matchers, e.g. ',eventType="move_executed"'.
  // Returns: '{service_name="<service>"<extraLabels>}'
  lokiStreamSelector(service, extraLabels='')::
    '{service_name="%s"%s}' % [service, extraLabels],

  // Build a LogQL pipeline string for game event streams.
  // Uses |= "game_event" line filter on body, then LABEL FILTERS for gameId/eventType
  // (these are Loki stream labels from OTLP ingest, not body content).
  // service: string (e.g. 'risk-it').
  // gameIdVar: string (optional) — Grafana variable for game ID label filter, e.g. '$gameId'.
  // eventTypeFilter: string (optional) — event type for label filter, e.g. 'move_executed'.
  //   Use |~ "regex" for multiple types (e.g. headlineFilter).
  // extraLabels: string (optional) — additional stream selector labels.
  // Returns: '{...} |= "game_event" [| gameId="<var>"] [| eventType="<type>"]'
  lokiEventStream(service, gameIdVar='', eventTypeFilter='', extraLabels='')::
    local selector = $.lokiStreamSelector(service, extraLabels);
    local gameFilter = if gameIdVar == '' then '' else ' | gameId=`%s`' % gameIdVar;
    local typeFilter =
      if eventTypeFilter == '' then ''
      else if std.length(std.findSubstr('|', eventTypeFilter)) > 0
      then ' | eventType=~`%s`' % eventTypeFilter
      else ' | eventType=`%s`' % eventTypeFilter;
    '%s |= "game_event"%s%s' % [selector, gameFilter, typeFilter],

  // Build a LogQL pipeline for State Timeline visualization.
  // Produces one stream per partitionField value, with valueField_turnNumber as the line content.
  // The turn suffix ensures mergeValues:true creates visual boundaries between turns
  // (CONQUER_0 and DEPLOY_4 are different values → boundary drawn).
  // Grafana regex value mappings strip the suffix: ^DEPLOY.* → blue, display "DEPLOY".
  // Uses | keep to reduce label cardinality, | line_format to set the suffixed value,
  // and | drop to remove non-partition labels from stream identity.
  lokiStateTimeline(service, gameIdVar, eventTypeFilter, partitionField, valueField)::
    '%s | keep %s, %s, payload_turn | line_format "{{ if eq .%s \\"WAITING\\" }}WAITING{{ else }}{{.%s}}_{{.payload_turn}}{{ end }}" | drop %s, payload_turn' % [
      $.lokiEventStream(service, gameIdVar, eventTypeFilter),
      partitionField,
      valueField,
      valueField,
      valueField,
      valueField,
    ],

  // Build a LogQL metric query with count_over_time.
  // OTLP logs create many streams (one per unique label set), so wraps with sum()
  // to aggregate across streams.
  // pipeline: string — a LogQL pipeline (e.g. from lokiEventStream).
  // range: string (default '$__range') — range duration.
  // Returns: 'sum(count_over_time(<pipeline>[<range>]))'
  lokiCountOverTime(pipeline, range='$__range')::
    'sum(count_over_time(%s[%s]))' % [pipeline, range],

  // Build a LogQL line_format stage to reshape output for table/timeline panels.
  // baseQuery: string — existing LogQL expression to extend.
  // fields: object — maps output field names to label references.
  // Returns: '<baseQuery> | line_format "{{.field1}} {{.field2}}"'
  // NOTE: For table panels, Grafana can access labels directly via the extractFields
  // transformation with source="labels". Use this only when you need custom formatting.
  lokiKeepLabels(baseQuery, labels)::
    '%s | keep %s' % [baseQuery, std.join(', ', labels)],
}
