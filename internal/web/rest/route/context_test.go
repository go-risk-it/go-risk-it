package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// --- ExtractID ---

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

// --- ExtractUserContext ---

func newUserContext(t *testing.T) kernelctx.UserContext {
	t.Helper()

	tc := kernelctx.WithSpan(t.Context(), noop.Span{})

	return kernelctx.WithUserID(tc, "user-123")
}

func TestExtractUserContext_Valid(t *testing.T) {
	t.Parallel()

	uc := newUserContext(t)
	req := httptest.NewRequestWithContext(uc, http.MethodGet, "/test", nil)

	result, err := route.ExtractUserContext(req)

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.UserID())
}

func TestExtractUserContext_MissingUserContext(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)

	_, err := route.ExtractUserContext(req)

	require.Error(t, err)

	var domainErr *domainerrors.DomainError

	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domainerrors.CategoryUnauthorized, domainErr.Category())
}

// --- BuildDomainContext ---

type testDomainContext struct {
	userID string
	id     int64
}

func TestBuildDomainContext_Success(t *testing.T) {
	t.Parallel()

	uc := newUserContext(t)
	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/things/42/action", nil)
	req.SetPathValue("id", "42")

	result, err := route.BuildDomainContext(
		req,
		func(userCtx kernelctx.UserContext, id int64) testDomainContext {
			return testDomainContext{userID: userCtx.UserID(), id: id}
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.userID)
	assert.Equal(t, int64(42), result.id)
}

func TestBuildDomainContext_MissingAuth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/things/42/action",
		nil,
	)
	req.SetPathValue("id", "42")

	_, err := route.BuildDomainContext(
		req,
		func(_ kernelctx.UserContext, _ int64) testDomainContext {
			t.Fatal("withID should not be called when auth is missing")

			return testDomainContext{}
		},
	)

	require.Error(t, err)

	var domainErr *domainerrors.DomainError

	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domainerrors.CategoryUnauthorized, domainErr.Category())
}

func TestBuildDomainContext_InvalidID(t *testing.T) {
	t.Parallel()

	uc := newUserContext(t)
	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/things/abc/action", nil)
	req.SetPathValue("id", "abc")

	_, err := route.BuildDomainContext(
		req,
		func(_ kernelctx.UserContext, _ int64) testDomainContext {
			t.Fatal("withID should not be called when ID is invalid")

			return testDomainContext{}
		},
	)

	require.Error(t, err)

	var domainErr *domainerrors.DomainError

	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domainerrors.CategoryValidation, domainErr.Category())
	assert.Contains(t, err.Error(), "invalid path parameter")
}
