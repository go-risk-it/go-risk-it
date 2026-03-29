// Auto-layout engine for OODA-structured Grafana dashboards.
// Replaces manual id/gridPos assignment with declarative panel placement.
//
// Usage:
//   local layout = import 'layout.libsonnet';
//   layout.ooda(
//     observe=[layout.panel(myStatPanel, 8, 6, 'Normal: < 1%.')],
//     orient=[layout.panel(myTimeseriesPanel, 12, 8)],
//     orientDepth={ 'Per-Route Breakdown': [layout.panel(p1, 24, 8)] },
//   )
{
  // Wrap a panel builder output with layout metadata.
  // w and h are required positional params (grid width and height).
  // Hidden fields (_w::, _h::) are excluded from JSON output automatically.
  panel(p, w, h, description=null)::
    p + {
      _w:: w,
      _h:: h,
      [if description != null then 'description']: description,
    },

  // OODA row titles (matches existing ooda.libsonnet).
  local rowTitles = {
    observe: 'Observe \u2014 Am I OK?',
    orient: 'Orient \u2014 What\u2019s the shape?',
    decide: 'Decide \u2014 Where\u2019s the bottleneck?',
    act: 'Act \u2014 What\u2019s the evidence?',
  },

  // Flow panels left-to-right, wrapping at x=24.
  // startY: first row y for panels, startId: first panel id.
  // Returns { panels: [...], y: nextY, id: nextId }.
  local flowPanels(items, startY, startId) =
    std.foldl(
      function(acc, p)
        local w = p._w;
        local h = p._h;
        // Wrap to next row if panel doesn't fit.
        local x = if acc.x + w > 24 then 0 else acc.x;
        local y = if acc.x + w > 24 then acc.maxY else acc.y;
        acc {
          x: x + w,
          y: y,
          maxY: std.max(acc.maxY, y + h),
          id: acc.id + 1,
          panels+: [p + { id: acc.id, gridPos: { h: h, w: w, x: x, y: y } }],
        },
      items,
      { panels: [], x: 0, y: startY, maxY: startY, id: startId },
    ),

  // Build a collapsed row with nested panels (absolute y-values).
  // Returns { panels: [rowObj], y: outerY + 1, id: nextId }.
  local addCollapsedRow(title, items, outerY, startId) =
    local rowId = startId;
    local innerStart = outerY + 1;
    local result = flowPanels(items, innerStart, rowId + 1);
    {
      panels: [{
        id: rowId,
        type: 'row',
        title: title,
        collapsed: true,
        gridPos: { h: 1, w: 24, x: 0, y: outerY },
        panels: result.panels,
      }],
      // Collapsed rows occupy only 1 row of external space.
      y: outerY + 1,
      id: result.id,
    },

  // Build an expanded OODA section row header.
  local makeRowHeader(title, y, id) = {
    id: id,
    type: 'row',
    title: title,
    collapsed: false,
    gridPos: { h: 1, w: 24, x: 0, y: y },
  },

  // Main entry point. Produces a flat panels array with auto-assigned IDs and gridPos.
  // observe/orient/decide/act: arrays of layout.panel() outputs (always-visible).
  // orientDepth/decideDepth/actDepth: objects mapping title -> panel arrays (collapsed rows).
  // Collapsed row order is alphabetical by title (std.objectFields sorts keys).
  ooda(
    observe=[],
    orient=[],
    orientDepth={},
    decide=[],
    decideDepth={},
    act=[],
    actDepth={},
  )::
    local sections = [
      { title: rowTitles.observe, visible: observe, depth: {} },
      { title: rowTitles.orient, visible: orient, depth: orientDepth },
      { title: rowTitles.decide, visible: decide, depth: decideDepth },
      { title: rowTitles.act, visible: act, depth: actDepth },
    ];
    local result = std.foldl(
      function(acc, section)
        // Row header for this OODA section.
        local rowHeader = makeRowHeader(section.title, acc.y, acc.id);
        local afterRow = acc { y: acc.y + 1, id: acc.id + 1, panels+: [rowHeader] };

        // Flow always-visible panels.
        local visible = flowPanels(section.visible, afterRow.y, afterRow.id);
        local afterVisible = afterRow {
          y: visible.maxY,
          id: visible.id,
          panels+: visible.panels,
        };

        // Add collapsed depth rows (alphabetical order).
        local depthKeys = std.objectFields(section.depth);
        std.foldl(
          function(inner, key)
            local collapsed = addCollapsedRow(
              key, section.depth[key], inner.y, inner.id
            );
            inner {
              y: collapsed.y,
              id: collapsed.id,
              panels+: collapsed.panels,
            },
          depthKeys,
          afterVisible,
        ),
      sections,
      { panels: [], y: 0, id: 1 },
    );
    result.panels,
}
