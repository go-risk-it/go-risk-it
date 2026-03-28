# LGTM Validation Spike — Findings

## A-1: Exemplar Storage — VALIDATED

**Evidence:**
- `run-prometheus.sh` inside LGTM container (`grafana/otel-lgtm:latest`) contains:
  ```
  --enable-feature=exemplar-storage
  ```
- Grafana Prometheus datasource already has `exemplarTraceIdDestinations` configured:
  ```yaml
  exemplarTraceIdDestinations:
    - name: trace_id
      datasourceUid: tempo
      urlDisplayLabel: "Trace: ${__value.raw}"
  ```
- Existing dashboards already use `histogramQuantileTargetsWithExemplars()` (perf-test, command-center dashboards).

**Conclusion:** Prometheus stores exemplars and the Grafana datasource is pre-configured to link them to Tempo traces. No additional configuration needed for Phase 3.

## A-2: Grafana Data Source UIDs — VALIDATED

**Source:** `/otel-lgtm/grafana/conf/provisioning/datasources/grafana-datasources.yaml` inside the LGTM container.

| Data Source | Type | UID | URL |
|-------------|------|-----|-----|
| Prometheus | prometheus | `prometheus` | http://127.0.0.1:9090 |
| Tempo | tempo | `tempo` | http://127.0.0.1:3200 |
| Loki | loki | `loki` | http://127.0.0.1:3100 |
| Pyroscope | grafana-pyroscope-datasource | `pyroscope` | http://127.0.0.1:4040 |

**For Jsonnet references:**
- `{ type: 'prometheus', uid: 'prometheus' }` — already used in `common.libsonnet`
- `{ type: 'tempo', uid: 'tempo' }` — use for lifecycle dashboard Act III
- `{ type: 'loki', uid: 'loki' }` — use for log panels

## Cross-Datasource Linking (partial A-12 evidence)

The LGTM image pre-configures bidirectional linking between all three data sources:

1. **Prometheus → Tempo**: `exemplarTraceIdDestinations` links exemplar `trace_id` to Tempo
2. **Tempo → Loki**: `tracesToLogsV2` with custom query `'{${__tags}} | trace_id = "${__trace.traceId}"'`
3. **Tempo → Prometheus**: `serviceMap.datasourceUid: "prometheus"` (service graph)
4. **Loki → Tempo**: `derivedFields` matches `trace_id` label and links to Tempo

This pre-wired cross-datasource linking means:
- Clicking an exemplar dot on a Prometheus panel navigates to Tempo trace view
- Tempo trace waterfall shows "Logs" tab linking to Loki
- Loki log entries with `trace_id` label show "Trace" link to Tempo

**However, A-12 (variable interpolation across data sources) still needs runtime validation** — the pre-configured links use Grafana's built-in link templating (`${__value.raw}`, `${__trace.traceId}`), not dashboard-level variable interpolation (`$traceId`). T4 must verify whether a dashboard `$traceId` textbox variable can drive a Loki panel query.

---

## Loki & otelslog Findings

### A-4: OTel Attributes Appear as Loki Structured Metadata — VALIDATED

**Evidence:**

The OTel collector sends logs to Loki via the native `/otlp` endpoint (confirmed in `/otel-lgtm/otelcol-config.yaml`). Loki's OTLP ingestion path automatically maps OTel log record attributes and trace context to **structured metadata**, NOT stream labels.

**Stream labels vs structured metadata:**

| Category | Field names | Storage | Query syntax |
|----------|------------|---------|-------------|
| Stream label | `service_name` | Indexed label | `{service_name="..."}` (stream selector) |
| Structured metadata | `gameID`, `userID`, `detail`, `trace_id`, `span_id`, `severity_text`, `severity_number`, `detected_level`, `scope_name`, `service_version`, `flags`, `observed_timestamp` | Structured metadata | `{...} \| field="value"` (pipe filter) |
| Log body | Message text only (e.g., `"spike validation log"`) | Log line | `{...} \|= "text"` (line filter) |

**Query syntax that WORKS for filtering by custom attributes:**

1. **Pipe filter on structured metadata (recommended):**
   ```
   {service_name="spike-lgtm-validation"} | gameID="42"
   ```
   Returns: 2 results (all log entries with gameID=42). This is the correct syntax for structured metadata.

2. **Multiple structured metadata filters chained:**
   ```
   {service_name="spike-lgtm-validation"} | gameID="42" | severity_text="WARN"
   ```
   Returns: 1 result (only the WARN entry). Multiple pipe filters AND together.

3. **Regex on structured metadata:**
   ```
   {service_name="spike-lgtm-validation"} | trace_id=~"6a83.*"
   ```
   Returns: 2 results. Regex matching works on structured metadata.

4. **Line filter + structured metadata:**
   ```
   {service_name="spike-lgtm-validation"} |= "warning" | gameID="42"
   ```
   Returns: 1 result. Line content and metadata filters can be combined.

**Query syntax that DOES NOT work:**

5. **Stream selector with custom attributes:**
   ```
   {service_name="spike-lgtm-validation", gameID="42"}
   ```
   Returns: 0 results. `gameID` is structured metadata, not a stream label. Stream selectors only match indexed labels.

6. **Stream selector with trace_id:**
   ```
   {service_name="spike-lgtm-validation", trace_id="6a83..."}
   ```
   Returns: 0 results. Same reason — `trace_id` is structured metadata.

7. **JSON parser pipeline (technically returns results but wrong):**
   ```
   {service_name="spike-lgtm-validation"} | json | gameID=`42`
   ```
   Returns: 2 results BUT with `__error__: JSONParserErr` because the log body is plain text, not JSON. The filter succeeds only because `gameID` is already available as structured metadata regardless of the parser. Using `| json` here is misleading and adds a parse error to the stream labels.

**Conclusion for dashboard queries:** Always use the pipe filter syntax (`| field="value"`) for OTel attributes. Never put them in stream selectors. The Tempo-to-Loki link pre-configured in LGTM already uses the correct syntax: `'{${__tags}} | trace_id = "${__trace.traceId}"'`.

### A-10: Loki Log Body Format — VALIDATED

**Evidence:**

The log body contains ONLY the slog message string (e.g., `"spike validation log"`, `"spike warning with trace context"`). All slog attributes (`gameID`, `userID`, `detail`) are stored as structured metadata, NOT embedded in the log line.

This means:
- `| json` parser is useless (body is not JSON)
- Line filters (`|=`) only match against the message text
- All filtering by attributes must use structured metadata pipe syntax

**Detected fields (from `/loki/api/v1/detected_fields`):**

All fields have `parsers: null` — confirming they are native structured metadata, not extracted by any log parser:

| Field | Type | Cardinality |
|-------|------|-------------|
| `gameID` | int | 1 |
| `userID` | string | 1 |
| `trace_id` | string | 2 |
| `span_id` | string | 2 |
| `severity_text` | string | 2 |
| `severity_number` | int | 2 |
| `detected_level` | string | 2 |
| `detail` | string | 1 |
| `scope_name` | string | 1 |
| `service_version` | string | 1 |
| `flags` | int | 1 |
| `observed_timestamp` | int | 4 |

### A-3: otelslog Bridge Trace Correlation — VALIDATED

**Evidence:**

The otelslog bridge (`go.opentelemetry.io/contrib/bridges/otelslog@v0.16.0`) works as follows:
1. `Handle(ctx, record)` calls `h.logger.Emit(ctx, h.convertRecord(record))` — passes the context through to the OTel SDK
2. The SDK's `Emit` method automatically extracts `trace_id` and `span_id` from the `context.Context` (if it carries an active span)
3. These are stored in the OTel LogRecord's dedicated trace context fields, NOT as log attributes
4. The OTel collector forwards them to Loki, where they appear as structured metadata fields named `trace_id` and `span_id`

The trace correlation requires passing a traced context to slog methods:
- `logger.InfoContext(parentCtx, "message", ...)` — trace_id IS attached (parentCtx carries an active span)
- `logger.Info("message", ...)` — trace_id is NOT attached (no context = no trace correlation)

**Exact field names in Loki:** `trace_id` (lowercase, underscore) and `span_id` (lowercase, underscore).

### A-8: No Conflict Between otelslog Automatic Trace Correlation and Manual Attributes — VALIDATED

**Evidence:**

The spike program does NOT manually add `trace_id` or `span_id` as slog attributes — it only passes `gameID`, `userID`, and `detail`. The trace context is automatically extracted by the OTel SDK from the context.Context passed to `InfoContext`/`WarnContext`.

In the OTel log data model, trace context (`trace_id`, `span_id`) lives in **dedicated fields** on the LogRecord, separate from the **attributes** map. The otelslog bridge:
1. Converts slog attributes (`gameID`, `userID`) → OTel LogRecord attributes
2. Passes context to `Emit()` → SDK extracts trace context to dedicated fields

These are two different paths that cannot conflict. However, if someone were to manually add `slog.String("trace_id", "...")` as an attribute, it would appear as BOTH a structured metadata field from the dedicated trace context AND as a separate attribute — creating ambiguity. **Recommendation: never manually add `trace_id` or `span_id` as slog attributes when using otelslog with traced contexts.**

**Cross-datasource linking verification:**
- Loki → Tempo: `derivedFields` config uses `matcherType: "label"` on `trace_id` with `datasourceUid: "tempo"`. This works because Loki's derived fields matching includes structured metadata. Clicking a log entry's `trace_id` value opens the corresponding Tempo trace.
- Tempo → Loki: Uses query `'{${__tags}} | trace_id = "${__trace.traceId}"'` — correctly uses pipe filter syntax for structured metadata.

## Tempo & Variable Findings

### A-5: TraceQL in Table Panel — VALIDATED

**Environment:** Grafana 12.4.1, Tempo 2.10.3 (inside `grafana/otel-lgtm:latest`)

**Evidence: Traces exist in Tempo**

The spike program emitted traces with service name `spike-lgtm-validation`. Tempo's tag values API confirms the service exists:
```
GET /api/search/tag/service.name/values
→ tagValues: ["risk-it", "spike-lgtm-validation"]
```

Two traces were found with `rootTraceName: "spike.validate"`, each containing a parent span (`spike.validate`) and child span (`spike.validate.child`) with correct parent/child relationship (verified by fetching full trace by ID).

**Evidence: Table panel query works**

The Grafana `ds/query` API with `queryType: "traceqlSearch"` returns a properly shaped table response:
```
Schema fields: [traceID, startTime, traceService, traceName, traceDuration, nested]
Trace count: 2
Service names: ["spike-lgtm-validation", "spike-lgtm-validation"]
Span names: ["spike.validate", "spike.validate"]
```

Both `queryType: "traceqlSearch"` and `queryType: "traceql"` work. The Grafana datasource plugin handles the server-side filtering correctly — even though Tempo's raw `api/search` endpoint with local backend returns all traces regardless of TraceQL predicates, the Grafana query layer filters post-hoc and returns only matching results.

**Caveat — `traceDuration: null`:** Tempo's local backend (used in LGTM) does not compute trace duration for the search index. The `traceDuration` column in Table panels will show `null`. This is a display-only issue — the data is still correctly filtered and linked. Duration is visible when opening the trace detail view.

**For Jsonnet Table panels:**
```jsonnet
// Use queryType: 'traceqlSearch' for Table panels
{
  datasource: { type: 'tempo', uid: 'tempo' },
  queryType: 'traceqlSearch',
  query: '{resource.service.name="risk-it"}',
  limit: 20,
  tableType: 'traces',
}
```

### A-12: Cross-Datasource Variable Interpolation — VALIDATED (with caveats)

**Evidence: Template variable definition**

The test dashboard defines a textbox variable:
```json
{
  "name": "traceId",
  "type": "textbox",
  "label": "Trace ID"
}
```
This is correctly structured per Grafana's template variable schema. The variable will appear in the dashboard toolbar as a text input field.

**Evidence: Loki query with variable**

The Loki panel query is:
```
{service_name="spike-lgtm-validation"} | trace_id=`$traceId`
```

When simulated with an actual trace ID (`661ab987844a90ce131fb83ea83c7f7a`), the Grafana `ds/query` API returns both log messages for that trace:
```
frame_count: 1
log_messages: ["spike warning with trace context", "spike validation log"]
```

**Critical finding — `trace_id` is structured metadata, not an indexed label:**

In the LGTM stack, the OTel Collector's Loki exporter promotes `trace_id` to appear in stream output (visible in `stream` JSON), but it is **not** an indexed Loki label. Only `service_name` is a true indexed label.

This means:
- `{trace_id="xxx"}` as a **label selector** — DOES NOT WORK (returns 0 results)
- `| trace_id=`xxx`` as a **pipeline filter** — WORKS (filters stream metadata post-fetch)

For the lifecycle dashboard, always use the pipeline filter syntax:
```
{service_name="risk-it"} | trace_id=`$traceId`
```

**Variable interpolation is client-side only:**

Grafana's `$traceId` interpolation happens in the browser before the query is sent to the backend. This cannot be fully tested via API — the API test above manually substituted the value. Full validation requires opening the dashboard at `http://localhost:3000/d/t4-tempo-validation/` and entering a trace ID in the textbox.

**Test dashboard provisioned at:** `http://localhost:3000/d/t4-tempo-validation/`

**For Jsonnet Loki panels with trace variable:**
```jsonnet
// Use pipeline filter, NOT label selector for trace_id
{
  datasource: { type: 'loki', uid: 'loki' },
  expr: '{service_name="risk-it"} | trace_id=`$traceId`',
}
```

### Pre-Existing Cross-Datasource Links (supplement to A-12)

The LGTM stack's pre-configured datasource linking (Prometheus exemplars → Tempo, Tempo → Loki, Loki → Tempo) uses Grafana's **built-in link templating** (`${__value.raw}`, `${__trace.traceId}`, `${__tags}`), which is distinct from **dashboard-level variable interpolation** (`$traceId`). Both mechanisms work and complement each other:

| Mechanism | How it works | Use case |
|-----------|-------------|----------|
| Exemplar links (`${__value.raw}`) | Click exemplar dot → opens Tempo trace | Prometheus histogram → Tempo |
| Trace-to-logs (`${__trace.traceId}`) | Tempo trace view → "Logs" tab | Tempo waterfall → Loki |
| Derived fields (`trace_id`) | Loki log entry → "Trace" link | Loki → Tempo |
| Dashboard variable (`$traceId`) | Textbox input → drives panel queries | Cross-panel filtering within a dashboard |

All four mechanisms are available in the LGTM stack without additional configuration.
