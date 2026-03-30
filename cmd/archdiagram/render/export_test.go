package render

// Exported wrappers for testing unexported functions from render_test package.

// ExtractTaxonomyLayersForTest wraps extractTaxonomyLayers for black-box tests.
var ExtractTaxonomyLayersForTest = extractTaxonomyLayers

// CompareLayersForTest wraps compareLayers for black-box tests.
var CompareLayersForTest = compareLayers
