package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func userContext(t *testing.T) ctx.UserContext {
	t.Helper()

	tc := ctx.WithSpan(t.Context(), noop.Span{})

	return ctx.WithUserID(tc, "user-123")
}

func TestBuildGameContext_ValidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/games/42/moves", nil)
	req.SetPathValue("id", "42")

	gc, err := route.BuildGameContext(req)

	require.NoError(t, err)
	assert.Equal(t, int64(42), gc.GameID())
	assert.Equal(t, "user-123", gc.UserID())
}

func TestBuildGameContext_InvalidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/games/abc/moves", nil)
	req.SetPathValue("id", "abc")

	_, err := route.BuildGameContext(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path parameter")
}

func TestBuildGameContext_MissingUserContext(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/games/42/moves",
		nil,
	)
	req.SetPathValue("id", "42")

	_, err := route.BuildGameContext(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user context not found")
}

func TestBuildLobbyContext_ValidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/lobbies/99/join", nil)
	req.SetPathValue("id", "99")

	lc, err := route.BuildLobbyContext(req)

	require.NoError(t, err)
	assert.Equal(t, int64(99), lc.LobbyID())
	assert.Equal(t, "user-123", lc.UserID())
}

func TestBuildLobbyContext_InvalidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/lobbies/xyz/join", nil)
	req.SetPathValue("id", "xyz")

	_, err := route.BuildLobbyContext(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid path parameter")
}

func TestBuildLobbyContext_MissingUserContext(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/lobbies/99/join",
		nil,
	)
	req.SetPathValue("id", "99")

	_, err := route.BuildLobbyContext(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user context not found")
}
