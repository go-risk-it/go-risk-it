package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGameContext_ValidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/games/42/moves", nil)
	req.SetPathValue("id", "42")

	gc, err := routes.BuildGameContext(req)

	require.NoError(t, err)
	assert.Equal(t, int64(42), gc.GameID())
	assert.Equal(t, "user-123", gc.UserID())
}

func TestBuildGameContext_InvalidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/games/abc/moves", nil)
	req.SetPathValue("id", "abc")

	_, err := routes.BuildGameContext(req)

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

	_, err := routes.BuildGameContext(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user context not found")
}
