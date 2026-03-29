package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/lobby/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildLobbyContext_ValidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/lobbies/99/join", nil)
	req.SetPathValue("id", "99")

	lc, err := routes.BuildLobbyContext(req)

	require.NoError(t, err)
	assert.Equal(t, int64(99), lc.LobbyID())
	assert.Equal(t, "user-123", lc.UserID())
}

func TestBuildLobbyContext_InvalidID(t *testing.T) {
	t.Parallel()

	uc := userContext(t)

	req := httptest.NewRequestWithContext(uc, http.MethodPost, "/api/v1/lobbies/xyz/join", nil)
	req.SetPathValue("id", "xyz")

	_, err := routes.BuildLobbyContext(req)

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

	_, err := routes.BuildLobbyContext(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user context not found")
}
