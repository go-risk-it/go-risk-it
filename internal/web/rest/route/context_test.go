package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractID_ValidID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test/42", nil)
	req.SetPathValue("id", "42")

	id, err := route.ExtractID(req)

	require.NoError(t, err)
	assert.Equal(t, int64(42), id)
}

func TestExtractID_InvalidID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test/abc", nil)
	req.SetPathValue("id", "abc")

	_, err := route.ExtractID(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}
