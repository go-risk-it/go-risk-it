package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/routes"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// --- Lobby ---

func TestLobby_RequiresAuth(t *testing.T) {
	t.Parallel()

	lobbyRoute := routes.Lobby(
		"POST /api/v1/lobbies/{id}/join",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.LobbyContext) error {
			return nil
		},
	)

	assert.True(t, lobbyRoute.RequiresAuth())
}

func TestLobby_ExtractsLobbyContext(t *testing.T) {
	t.Parallel()

	var receivedLobbyID int64

	lobbyRoute := routes.Lobby(
		"POST /api/v1/lobbies/{id}/join",
		func(_ http.ResponseWriter, _ *http.Request, lc ctx.LobbyContext) error {
			receivedLobbyID = lc.LobbyID()

			return nil
		},
	)

	rec := httptest.NewRecorder()
	userCtx := userContext(t)
	req := httptest.NewRequestWithContext(userCtx, http.MethodPost, "/api/v1/lobbies/99/join", nil)
	req.SetPathValue("id", "99")
	lobbyRoute.ServeHTTP(rec, req)

	assert.Equal(t, int64(99), receivedLobbyID)
}

func TestLobby_InvalidID_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	traceCtx := kernelctx.WithSpan(spanCtx, noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "user-123")

	lobbyRoute := routes.Lobby(
		"POST /api/v1/lobbies/{id}/join",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.LobbyContext) error {
			t.Fatal("handler should not be called")

			return nil
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/lobbies/xyz/join",
		nil,
	)
	req.SetPathValue("id", "xyz")
	lobbyRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

// --- LobbyWS ---

func TestLobbyWS_IsWebSocket(t *testing.T) {
	t.Parallel()

	wsRoute := routes.LobbyWS(
		"GET /api/v1/lobbies/{id}/ws",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.LobbyContext) error {
			return nil
		},
	)

	assert.True(t, wsRoute.IsWebSocket())
	assert.True(t, wsRoute.RequiresAuth())
}
