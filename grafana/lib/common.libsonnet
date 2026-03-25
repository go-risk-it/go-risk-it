// Reusable panel builder functions for go-risk-it Grafana dashboards.
// Encodes project conventions (P1-datasource, P4-panel-style).
{
  // P1-datasource: every panel uses the same Prometheus datasource.
  datasource():: { type: 'prometheus', uid: 'prometheus' },

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

  // Generate histogram_quantile targets for a metric.
  // metric: string (bucket metric name), quantiles: array of [quantile_str, label] pairs
  // quantile_str should be a string like "0.5", "0.95", "0.99" to avoid float precision issues.
  // serviceName: string (OTel service_name label, default 'risk-it'; use 'perftest' for client metrics)
  // Returns array of target objects.
  histogramQuantileTargets(metric, quantiles, serviceName='risk-it')::
    [
      {
        expr: 'histogram_quantile(%s, sum(rate(%s{service_name="%s"}[1m])) by (le))' % [q[0], metric, serviceName],
        legendFormat: q[1],
        refId: std.char(65 + i),  // A, B, C, ...
      }
      for i in std.range(0, std.length(quantiles) - 1)
      for q in [quantiles[i]]
    ],

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
  // Inner band (p95→p50) has fillOpacity 10, outer band (p99→p95) has fillOpacity 5.
  // Uses fillBelowTo overrides so each band fills down to the next percentile line.
  //
  // Usage:
  //   common.percentileBandsPanel(
  //     title='HTTP Latency Bands',
  //     metric='http_server_request_duration_seconds_bucket',
  //     unit='s',
  //   ) + { id: 1, gridPos: { h: 8, w: 12, x: 0, y: 0 } }
  percentileBandsPanel(title, metric, unit, serviceName='risk-it')::
    $.timeseriesPanel(
      title=title,
      targets=$.histogramQuantileTargets(metric, [['0.5', 'p50'], ['0.95', 'p95'], ['0.99', 'p99']], serviceName=serviceName),
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

  // Build a row panel (section header).
  // title: string
  rowPanel(title)::
    {
      title: title,
      type: 'row',
      collapsed: false,
    },
}
