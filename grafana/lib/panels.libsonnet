// Panel builder functions for go-risk-it Grafana dashboards.
// Encodes project conventions (P1-datasource, P4-panel-style).
local targets = import 'targets.libsonnet';
{
  // P1-datasource: every panel uses the same Prometheus datasource.
  datasource():: targets.datasources.prometheus,

  // P4-panel-style: standard timeseries visual settings.
  timeseriesDefaults():: {
    drawStyle: 'line',
    lineInterpolation: 'smooth',
    fillOpacity: 10,
    lineWidth: 2,
    pointSize: 5,
    showPoints: 'never',
    spanNulls: false,
    stacking: { group: 'A', mode: 'none' },
  },

  // Legend as table with mean/max calcs (most dashboards use this).
  legendTable():: {
    displayMode: 'table',
    placement: 'bottom',
    calcs: ['mean', 'max'],
  },

  // Legend with sum/last calcs (for rate/counter panels).
  rateLegend():: {
    displayMode: 'table',
    placement: 'bottom',
    calcs: ['sum', 'last'],
  },

  // Multi-series tooltip sorted descending.
  tooltipMulti():: {
    mode: 'multi',
    sort: 'desc',
  },

  // Build a complete timeseries panel.
  // title: string, targets: array, unit: string,
  // overrides: array (optional), color: object (optional)
  timeseriesPanel(title, targets, unit, overrides=[], color=null)::
    {
      title: title,
      type: 'timeseries',
      datasource: $.datasource(),
      targets: targets,
      fieldConfig: {
        defaults: {
          unit: unit,
          custom: $.timeseriesDefaults(),
          [if color != null then 'color']: color,
        },
        overrides: overrides,
      },
      options: {
        legend: $.legendTable(),
        tooltip: $.tooltipMulti(),
      },
    },

  // Build a stat panel with thresholds.
  // title: string, targets: array, thresholds: object, unit: string (optional),
  // colorMode: string (optional, default 'value'; use 'background' for SLO tiles)
  statPanel(title, targets, thresholds, unit=null, colorMode='value')::
    {
      title: title,
      type: 'stat',
      datasource: $.datasource(),
      targets: targets,
      fieldConfig: {
        defaults: {
          thresholds: thresholds,
          [if unit != null then 'unit']: unit,
          // 'background' colorMode requires thresholds-driven color.
          [if colorMode == 'background' then 'color']: { mode: 'thresholds' },
        },
        overrides: [],
      },
      options: {
        colorMode: colorMode,
        graphMode: 'area',
        justifyMode: 'auto',
        textMode: 'auto',
        reduceOptions: {
          calcs: ['lastNotNull'],
          fields: '',
          values: false,
        },
      },
    },

  // Build a gauge panel with thresholds.
  gaugePanel(title, targets, thresholds, unit, min=0, max=100)::
    {
      title: title,
      type: 'gauge',
      datasource: $.datasource(),
      targets: targets,
      fieldConfig: {
        defaults: {
          unit: unit,
          min: min,
          max: max,
          thresholds: thresholds,
        },
        overrides: [],
      },
      options: {
        reduceOptions: { calcs: ['lastNotNull'], fields: '', values: false },
        showThresholdLabels: false,
        showThresholdMarkers: true,
      },
    },

  // Build a heatmap panel.
  // title: string, targets: array, unit: string (default 's'),
  // colorScheme: string (default 'Oranges'), colorFill: string (default 'dark-orange')
  // NOTE: Targets must include format:'heatmap' and legendFormat:'{{le}}' (caller responsibility).
  heatmapPanel(title, targets, unit='s', colorScheme='Oranges', colorFill='dark-orange')::
    {
      title: title,
      type: 'heatmap',
      datasource: $.datasource(),
      targets: targets,
      fieldConfig: {
        defaults: {
          color: { mode: 'scheme', schemeName: colorScheme },
          custom: {
            fillOpacity: 80,
            hideFrom: { legend: false, tooltip: false, viz: false },
            scaleDistribution: { type: 'linear' },
          },
        },
        overrides: [],
      },
      options: {
        calculate: false,
        cellGap: 1,
        color: {
          exponent: 0.5,
          fill: colorFill,
          mode: 'scheme',
          reverse: false,
          scale: 'exponential',
          scheme: colorScheme,
          steps: 64,
        },
        exemplars: { color: 'rgba(255,0,255,0.7)' },
        filterValues: { le: 1e-9 },
        legend: { show: true },
        rowsFrame: { layout: 'auto' },
        tooltip: { mode: 'single', showColorScale: false, yHistogram: false },
        yAxis: { axisPlacement: 'left', reverse: false, unit: unit },
      },
    },

  // Build a percentile bands timeseries panel with filled areas between p50-p95-p99.
  // Inner band (p95->p50) has fillOpacity 10, outer band (p99->p95) has fillOpacity 5.
  // Uses fillBelowTo overrides so each band fills down to the next percentile line.
  // Uses manual metrics (service_name label). For spanmetrics, use spanPercentileBandsPanel.
  percentileBandsPanel(title, metric, unit, serviceName='risk-it', exemplars=false)::
    $.timeseriesPanel(
      title=title,
      targets=if exemplars
        then targets.histogramQuantileTargetsWithExemplars(metric, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], serviceName=serviceName)
        else targets.histogramQuantileTargets(metric, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], serviceName=serviceName),
      unit=unit,
      overrides=[
        {
          matcher: { id: 'byName', options: 'p99' },
          properties: [
            { id: 'custom.fillBelowTo', value: 'p95' },
            { id: 'custom.fillOpacity', value: 5 },
          ],
        },
        {
          matcher: { id: 'byName', options: 'p95' },
          properties: [
            { id: 'custom.fillBelowTo', value: 'p50' },
            { id: 'custom.fillOpacity', value: 10 },
          ],
        },
      ],
    ) + {
      // Override default fillOpacity to 0 so only the band overrides apply.
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 0,
          },
        },
      },
    },

  // Build a percentile bands panel from spanmetrics duration histograms.
  // spanNameFilter: regex matching span_name (use targets.spans.* constants).
  // Same visual treatment as percentileBandsPanel but queries the spanmetrics connector.
  spanPercentileBandsPanel(title, spanNameFilter, unit, exemplars=false)::
    $.timeseriesPanel(
      title=title,
      targets=targets.spanDuration(spanNameFilter, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], exemplars=exemplars),
      unit=unit,
      overrides=[
        {
          matcher: { id: 'byName', options: 'p99' },
          properties: [
            { id: 'custom.fillBelowTo', value: 'p95' },
            { id: 'custom.fillOpacity', value: 5 },
          ],
        },
        {
          matcher: { id: 'byName', options: 'p95' },
          properties: [
            { id: 'custom.fillBelowTo', value: 'p50' },
            { id: 'custom.fillOpacity', value: 10 },
          ],
        },
      ],
    ) + {
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 0,
          },
        },
      },
    },

  // Build a bar gauge panel.
  // title: string, targets: array, thresholds: object (optional), unit: string (optional)
  // Uses palette-classic by default. If thresholds provided, uses thresholds color mode.
  barGaugePanel(title, targets, thresholds=null, unit=null)::
    {
      title: title,
      type: 'bargauge',
      datasource: $.datasource(),
      targets: targets,
      fieldConfig: {
        defaults: {
          color: { mode: if thresholds != null then 'thresholds' else 'palette-classic' },
          [if thresholds != null then 'thresholds']: thresholds,
          [if unit != null then 'unit']: unit,
        },
        overrides: [],
      },
      options: {
        displayMode: 'gradient',
        maxVizHeight: 300,
        minVizHeight: 16,
        minVizWidth: 8,
        namePlacement: 'auto',
        orientation: 'horizontal',
        reduceOptions: {
          calcs: ['lastNotNull'],
          fields: '',
          values: false,
        },
        showUnfilled: true,
        sizing: 'auto',
        valueMode: 'color',
      },
    },

  // Build a log panel with Loki datasource.
  // title: string, expr: string (LogQL query),
  // showLabels: bool (default false), sortOrder: string (default 'Descending'),
  // prettifyLogMessage: bool (default true)
  logPanel(title, expr, showLabels=false, sortOrder='Descending', prettifyLogMessage=true)::
    {
      title: title,
      type: 'logs',
      datasource: targets.datasources.loki,
      targets: [{
        refId: 'A',
        expr: expr,
      }],
      options: {
        showTime: true,
        showLabels: showLabels,
        showCommonLabels: false,
        wrapLogMessage: true,
        prettifyLogMessage: prettifyLogMessage,
        enableLogDetails: true,
        sortOrder: sortOrder,
        dedupStrategy: 'none',
      },
    },

  // Build a traces panel with Tempo datasource.
  // title: string, query: TraceQL string.
  tracesPanel(title, query)::
    {
      title: title,
      type: 'traces',
      datasource: targets.datasources.tempo,
      targets: [{
        refId: 'A',
        queryType: 'traceql',
        query: query,
      }],
    },

  // Build a state timeline panel.
  // title: string, targets: array, colorMap: object (optional, state->hex color mapping).
  // When colorMap is provided, generates value mappings so each state gets a fixed color.
  stateTimelinePanel(title, targets, colorMap={})::
    {
      title: title,
      type: 'state-timeline',
      datasource: $.datasource(),
      targets: targets,
      fieldConfig: {
        defaults: {
          color: { mode: 'fixed' },
          custom: {
            fillOpacity: 70,
            lineWidth: 0,
          },
          [if std.length(colorMap) > 0 then 'mappings']: [{
            type: 'value',
            options: {
              [state]: { text: state, color: colorMap[state] }
              for state in std.objectFields(colorMap)
            },
          }],
        },
        overrides: [],
      },
      options: {
        mergeValues: true,
        showValue: 'auto',
        alignValue: 'center',
        rowHeight: 0.9,
        legend: { displayMode: 'list', placement: 'bottom' },
        tooltip: { mode: 'single' },
      },
    },

  // Build a row panel (section header).
  rowPanel(title)::
    {
      title: title,
      type: 'row',
      collapsed: false,
    },
}
