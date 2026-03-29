package route_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Public ---

func TestPublic_NoAuth(t *testing.T) {
	t.Parallel()

	publicRoute := route.Public("GET /status", func(_ http.ResponseWriter, _ *http.Request) error {
		return nil
	})

	assert.Equal(t, "GET /status", publicRoute.Pattern())
	assert.False(t, publicRoute.RequiresAuth())
}

func TestPublic_IsNotWebSocket(t *testing.T) {
	t.Parallel()

	publicRoute := route.Public("GET /status", func(_ http.ResponseWriter, _ *http.Request) error {
		return nil
	})

	assert.False(t, publicRoute.IsWebSocket())
}

func TestPublic_SuccessPassesThrough(t *testing.T) {
	t.Parallel()

	publicRoute := route.Public("GET /status", func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)

		return nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/status", nil)
	publicRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPublic_ErrorMapsToHTTPResponse(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	publicRoute := route.Public("GET /fail", func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(spanCtx, http.MethodGet, "/fail", nil)
	publicRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// --- Authed ---

func TestAuthed_RequiresAuth(t *testing.T) {
	t.Parallel()

	authedRoute := route.Authed(
		"POST /api/v1/games",
		func(_ http.ResponseWriter, _ *http.Request) error {
			return nil
		},
	)

	assert.Equal(t, "POST /api/v1/games", authedRoute.Pattern())
	assert.True(t, authedRoute.RequiresAuth())
}

// --- Wrap ---

func TestRoute_Wrap_PreservesMetadata(t *testing.T) {
	t.Parallel()

	original := route.Authed(
		"POST /api/v1/games/{id}/deploy",
		func(_ http.ResponseWriter, _ *http.Request) error {
			return nil
		},
	)

	wrapped := original.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		// no-op wrapper
	}))

	assert.Equal(t, original.Pattern(), wrapped.Pattern())
	assert.Equal(t, original.RequiresAuth(), wrapped.RequiresAuth())
	assert.Equal(t, original.IsWebSocket(), wrapped.IsWebSocket())
}

func TestRoute_Wrap_UsesNewHandler(t *testing.T) {
	t.Parallel()

	original := route.Public(
		"GET /status",
		func(_ http.ResponseWriter, _ *http.Request) error {
			return nil
		},
	)

	called := false
	wrapped := original.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/status", nil)
	wrapped.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRoute_Wrap_PreservesWebSocketFlag(t *testing.T) {
	t.Parallel()

	wsRoute := route.NewWS(
		"GET /api/v1/games/{id}/ws",
		true,
		http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
	)

	wrapped := wsRoute.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	assert.True(t, wrapped.IsWebSocket())
	assert.True(t, wrapped.RequiresAuth())
	assert.Equal(t, wsRoute.Pattern(), wrapped.Pattern())
}

// --- Domain ---

const testContextValue = "ctx"

func TestDomain_RequiresAuth(t *testing.T) {
	t.Parallel()

	domainRoute := route.Domain(
		"POST /api/v1/things/{id}/action",
		func(_ *http.Request) (string, error) {
			return testContextValue, nil
		},
		func(_ http.ResponseWriter, _ *http.Request, _ string) error {
			return nil
		},
	)

	assert.True(t, domainRoute.RequiresAuth())
	assert.False(t, domainRoute.IsWebSocket())
}

func TestDomain_ExtractsContext(t *testing.T) {
	t.Parallel()

	type testContext struct {
		value string
	}

	var received testContext

	domainRoute := route.Domain(
		"POST /api/v1/things/{id}/action",
		func(_ *http.Request) (testContext, error) {
			return testContext{value: "hello"}, nil
		},
		func(_ http.ResponseWriter, _ *http.Request, tc testContext) error {
			received = tc

			return nil
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/things/1/action",
		nil,
	)
	domainRoute.ServeHTTP(rec, req)

	assert.Equal(t, "hello", received.value)
}

func TestDomain_BuildCtxError_ReturnsHTTPError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	domainRoute := route.Domain(
		"POST /api/v1/things/{id}/action",
		func(_ *http.Request) (string, error) {
			return "", domainerrors.NewValidationError("bad id")
		},
		func(_ http.ResponseWriter, _ *http.Request, _ string) error {
			t.Fatal("handler should not be called")

			return nil
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		spanCtx,
		http.MethodPost,
		"/api/v1/things/abc/action",
		nil,
	)
	domainRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestDomain_HandlerError_MapsToHTTPResponse(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	domainRoute := route.Domain(
		"POST /api/v1/things/{id}/action",
		func(_ *http.Request) (string, error) {
			return testContextValue, nil
		},
		func(_ http.ResponseWriter, _ *http.Request, _ string) error {
			return domainerrors.NewConflictError("wrong state")
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(spanCtx, http.MethodPost, "/api/v1/things/1/action", nil)
	domainRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "CONFLICT", resp.Code)
}

// --- DomainWS ---

func TestDomainWS_CreatesWSRoute(t *testing.T) {
	t.Parallel()

	wsRoute := route.DomainWS(
		"GET /api/v1/things/{id}/ws",
		func(_ *http.Request) (string, error) {
			return testContextValue, nil
		},
		func(_ http.ResponseWriter, _ *http.Request, _ string) error {
			return nil
		},
	)

	assert.True(t, wsRoute.IsWebSocket())
	assert.True(t, wsRoute.RequiresAuth())
	assert.Equal(t, "GET /api/v1/things/{id}/ws", wsRoute.Pattern())
}
