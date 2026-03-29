// Dashboard envelope factory and shared annotation constants.
// Eliminates per-dashboard boilerplate for the standard envelope fields.
{
  // Annotation constant for perf test phase markers.
  // Used by perf-test, perf-test-command-center, and request-lifecycle dashboards.
  perfTestAnnotation:: {
    builtIn: 1,
    datasource: { type: 'grafana', uid: '-- Grafana --' },
    enable: true,
    hide: false,
    iconColor: 'rgba(0, 211, 255, 1)',
    name: 'Perf Test Phases',
    type: 'dashboard',
    target: { matchAny: true, tags: ['perf-test'], type: 'tags' },
  },

  // Dashboard envelope factory.
  // uid, title, panels are required. All other fields have sensible defaults.
  // description and graphTooltip are omitted from output when null.
  new(
    uid,
    title,
    panels,
    description=null,
    tags=[],
    refresh='10s',
    templating={ list: [] },
    annotations={ list: [] },
    graphTooltip=null,
  ):: {
    uid: uid,
    title: title,
    [if description != null then 'description']: description,
    schemaVersion: 39,
    version: 1,
    timezone: 'browser',
    editable: true,
    [if graphTooltip != null then 'graphTooltip']: graphTooltip,
    tags: tags,
    time: { from: 'now-15m', to: 'now' },
    refresh: refresh,
    templating: templating,
    annotations: annotations,
    panels: panels,
  },
}
