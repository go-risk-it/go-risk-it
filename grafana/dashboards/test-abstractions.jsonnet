// Test dashboard — validates all library abstractions from dashboard.libsonnet and common.libsonnet.
// NOT a real dashboard — used for regression testing during the DRY refactor.
// Compile: jsonnet -J grafana/lib grafana/dashboards/test-abstractions.jsonnet
local common = import 'common.libsonnet';
local dashboard = import 'dashboard.libsonnet';
local thresholds = import 'thresholds.libsonnet';

dashboard.new(
  uid='test-abstractions',
  title='Test Abstractions (not a real dashboard)',
  description='Exercises every library abstraction for validation',
  tags=['test'],
  refresh='5s',
  graphTooltip=1,
  annotations={ list: [dashboard.perfTestAnnotation] },
  panels=[
    // 1. withSloThreshold — standalone
    common.timeseriesPanel(
      title='SLO Threshold Test',
      targets=[{ expr: 'up', legendFormat: 'up', refId: 'A' }],
      unit='s',
    ) + common.withSloThreshold(thresholds.e2eP95) + {
      id: 1,
      gridPos: { h: 8, w: 12, x: 0, y: 0 },
    },

    // 2. withSloThreshold + sibling fillOpacity (composability edge case)
    common.timeseriesPanel(
      title='SLO + FillOpacity Composability Test',
      targets=[{ expr: 'up', legendFormat: 'up', refId: 'A' }],
      unit='s',
    ) + common.withSloThreshold(thresholds.httpError) + {
      id: 2,
      gridPos: { h: 8, w: 12, x: 12, y: 0 },
      fieldConfig+: {
        defaults+: {
          custom+: {
            fillOpacity: 15,
          },
        },
      },
    },

    // 3. logPanel — defaults (showLabels=false, sortOrder=Descending, prettifyLogMessage=true)
    common.logPanel(
      title='Log Panel Defaults Test',
      expr='{service_name="risk-it"} |= "test"',
    ) + {
      id: 3,
      gridPos: { h: 8, w: 12, x: 0, y: 8 },
    },

    // 4. logPanel — request-lifecycle variant
    common.logPanel(
      title='Log Panel Request-Lifecycle Variant Test',
      expr='{service_name="risk-it"} | trace_id=`test`',
      showLabels=true,
      sortOrder='Ascending',
    ) + {
      id: 4,
      gridPos: { h: 8, w: 12, x: 12, y: 8 },
    },

    // 5. lifecycleTargets + lifecycleOverrides
    common.timeseriesPanel(
      title='Lifecycle Targets/Overrides Test',
      targets=common.lifecycleTargets,
      unit='s',
      overrides=common.lifecycleOverrides,
    ) + {
      id: 5,
      gridPos: { h: 8, w: 24, x: 0, y: 16 },
    },
  ],
)
