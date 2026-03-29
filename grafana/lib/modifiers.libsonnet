// Composable panel modifiers for go-risk-it Grafana dashboards.
// Each modifier returns a merge object — apply with `panel + modifier`.
// Modifiers compose: `panel + withStackedArea() + withSeriesColors({...})`.
local colors = import 'colors.libsonnet';
{
  // SLO threshold overlay for timeseries panels.
  // Adds threshold lines + shaded area.
  // Usage: panel + modifiers.withSloThreshold(thresholds.e2eP95)
  withSloThreshold(threshold):: {
    fieldConfig+: {
      defaults+: {
        thresholds: threshold,
        custom+: {
          thresholdsStyle: { mode: 'line+area' },
        },
      },
    },
  },

  // Stacked area chart modifier.
  // Preserves drawStyle and lineInterpolation from the base panel.
  // Usage: panel + modifiers.withStackedArea()
  withStackedArea(fillOpacity=20, gradientMode='none'):: {
    fieldConfig+: {
      defaults+: {
        custom+: {
          stacking+: { mode: 'normal' },
          fillOpacity: fillOpacity,
          gradientMode: gradientMode,
          lineWidth: 1,
        },
      },
    },
  },

  // Per-series color overrides.
  // colorMap: object mapping series names to hex colors.
  // Usage: panel + modifiers.withSeriesColors({ 'p95': '#FF9830', 'p99': '#E02F44' })
  withSeriesColors(colorMap):: {
    fieldConfig+: {
      overrides+: [
        {
          matcher: { id: 'byName', options: name },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colorMap[name] } },
          ],
        }
        for name in std.objectFields(colorMap)
      ],
    },
  },

  // Dashed line style for specific series (e.g., async or projected).
  // Applies dash pattern [6,4] and removes fill.
  // Usage: panel + modifiers.withDashedSeries(['Event Handler', 'Projected'])
  withDashedSeries(seriesNames):: {
    fieldConfig+: {
      overrides+: [
        {
          matcher: { id: 'byName', options: name },
          properties: [
            { id: 'custom.lineStyle', value: { fill: 'dash', dash: [6, 4] } },
            { id: 'custom.fillOpacity', value: 0 },
          ],
        }
        for name in seriesNames
      ],
    },
  },

  // Data links for cross-dashboard navigation or external URLs.
  // dataLinks: array of link objects (e.g., from links.libsonnet).
  // Usage: panel + modifiers.withLinks([links.toDashboard('Lifecycle', links.dashboardUids.lifecycle)])
  withLinks(dataLinks):: {
    fieldConfig+: {
      defaults+: {
        links: dataLinks,
      },
    },
  },

  // Percentile shade colors for within-boundary multi-series panels.
  // Applies light/medium/dark shades to p50/p95/p99 series.
  // Usage: panel + modifiers.withPercentileColors('db')
  withPercentileColors(boundaryColorKey):: {
    fieldConfig+: {
      overrides+: [
        {
          matcher: { id: 'byName', options: 'p50' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.shades[boundaryColorKey].light } },
          ],
        },
        {
          matcher: { id: 'byName', options: 'p95' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.shades[boundaryColorKey].medium } },
          ],
        },
        {
          matcher: { id: 'byName', options: 'p99' },
          properties: [
            { id: 'color', value: { mode: 'fixed', fixedColor: colors.shades[boundaryColorKey].dark } },
          ],
        },
      ],
    },
  },
}
