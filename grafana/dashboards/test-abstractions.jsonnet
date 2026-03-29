// Test dashboard — validates all library abstractions.
// NOT a real dashboard — used for regression testing during the DRY refactor.
// Compile: jsonnet -J grafana/lib grafana/dashboards/test-abstractions.jsonnet
local common = import 'common.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local layout = import 'layout.libsonnet';
local modifiers = import 'modifiers.libsonnet';
local panels = import 'panels.libsonnet';
local thresholds = import 'thresholds.libsonnet';

dashboard.new(
  uid='test-abstractions',
  title='Test Abstractions (not a real dashboard)',
  description='Exercises every library abstraction for validation',
  tags=['test'],
  refresh='5s',
  graphTooltip=1,
  annotations={ list: [dashboard.perfTestAnnotation] },
  panels=layout.ooda(

    // OBSERVE — stat panels + SLO threshold composability tests
    observe=[
      // 1. withSloThreshold — standalone
      layout.panel(
        panels.timeseriesPanel(
          title='SLO Threshold Test',
          targets=[{ expr: 'up', legendFormat: 'up', refId: 'A' }],
          unit='s',
        ) + modifiers.withSloThreshold(thresholds.e2eP95),
        w=12, h=8,
        description='Normal: validates SLO threshold overlay. Watch for: missing threshold line. Check next: composability test.',
      ),

      // 2. withSloThreshold + sibling fillOpacity (composability edge case)
      layout.panel(
        panels.timeseriesPanel(
          title='SLO + FillOpacity Composability Test',
          targets=[{ expr: 'up', legendFormat: 'up', refId: 'A' }],
          unit='s',
        ) + modifiers.withSloThreshold(thresholds.httpError) + {
          fieldConfig+: {
            defaults+: {
              custom+: {
                fillOpacity: 15,
              },
            },
          },
        },
        w=12, h=8,
        description='Normal: SLO threshold + custom fillOpacity coexist. Watch for: fillOpacity overridden by threshold. Check next: log panel tests.',
      ),
    ],

    // ORIENT — log panel tests + lifecycle targets
    orient=[
      // 3. logPanel — defaults (showLabels=false, sortOrder=Descending, prettifyLogMessage=true)
      layout.panel(
        panels.logPanel(
          title='Log Panel Defaults Test',
          expr='{service_name="risk-it"} |= "test"',
        ),
        w=12, h=8,
        description='Normal: log panel with default settings. Watch for: prettify or sort order changes. Check next: variant test.',
      ),

      // 4. logPanel — request-lifecycle variant
      layout.panel(
        panels.logPanel(
          title='Log Panel Request-Lifecycle Variant Test',
          expr='{service_name="risk-it"} | trace_id=`test`',
          showLabels=true,
          sortOrder='Ascending',
        ),
        w=12, h=8,
        description='Normal: labels visible, ascending sort. Watch for: missing labels or wrong sort order. Check next: lifecycle targets.',
      ),

      // 5. lifecycleTargets + lifecycleOverrides
      layout.panel(
        panels.timeseriesPanel(
          title='Lifecycle Targets/Overrides Test',
          targets=panels.lifecycleTargets,
          unit='s',
          overrides=panels.lifecycleOverrides,
        ),
        w=24, h=8,
        description='Normal: 5 lifecycle boundaries with correct colors. Watch for: missing series or wrong colors. Check next: collapsed row.',
      ),
    ],

    orientDepth={
      // Collapsed row test — validates layout.addCollapsedRow
      'Collapsed Row Test': [
        layout.panel(
          panels.statPanel(
            title='Collapsed Stat Test',
            targets=[{ expr: 'up', legendFormat: 'up', refId: 'A' }],
            thresholds=thresholds.activeGames,
          ),
          w=12, h=8,
          description='Normal: stat panel inside collapsed row. Watch for: panel not rendering on expand. Check next: nothing.',
        ),

        layout.panel(
          panels.timeseriesPanel(
            title='Collapsed Timeseries Test',
            targets=[{ expr: 'up', legendFormat: 'up', refId: 'A' }],
            unit='short',
          ) + modifiers.withStackedArea(),
          w=12, h=8,
          description='Normal: stacked area inside collapsed row. Watch for: stacking mode not applied. Check next: nothing.',
        ),
      ],
    },
  ),
)
