package route_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
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

// --- Game ---

func TestGame_RequiresAuth(t *testing.T) {
	t.Parallel()

	gameRoute := route.Game(
		"POST /api/v1/games/{id}/advancements",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.GameContext) error {
			return nil
		},
	)

	assert.True(t, gameRoute.RequiresAuth())
}

func TestGame_ExtractsGameContext(t *testing.T) {
	t.Parallel()

	var receivedGameID int64

	gameRoute := route.Game(
		"POST /api/v1/games/{id}/advancements",
		func(_ http.ResponseWriter, _ *http.Request, gc ctx.GameContext) error {
			receivedGameID = gc.GameID()

			return nil
		},
	)

	rec := httptest.NewRecorder()
	uc := userContext(t)
	req := httptest.NewRequestWithContext(
		uc,
		http.MethodPost,
		"/api/v1/games/42/advancements",
		nil,
	)
	req.SetPathValue("id", "42")
	gameRoute.ServeHTTP(rec, req)

	assert.Equal(t, int64(42), receivedGameID)
}

func TestGame_InvalidID_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	traceCtx := ctx.WithSpan(spanCtx, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-123")

	gameRoute := route.Game(
		"POST /api/v1/games/{id}/moves",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.GameContext) error {
			t.Fatal("handler should not be called")

			return nil
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(userCtx, http.MethodPost, "/api/v1/games/abc/moves", nil)
	req.SetPathValue("id", "abc")
	gameRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestGame_MissingUserContext_ReturnsError(t *testing.T) {
	t.Parallel()

	gameRoute := route.Game(
		"POST /api/v1/games/{id}/moves",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.GameContext) error {
			t.Fatal("handler should not be called")

			return nil
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/api/v1/games/42/moves",
		nil,
	)
	req.SetPathValue("id", "42")
	gameRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGame_HandlerError_MapsToHTTPResponse(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	traceCtx := ctx.WithSpan(spanCtx, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-123")

	gameRoute := route.Game(
		"POST /api/v1/games/{id}/advancements",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.GameContext) error {
			return domainerrors.NewConflictError("wrong phase")
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/games/42/advancements",
		nil,
	)
	req.SetPathValue("id", "42")
	gameRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "CONFLICT", resp.Code)
}

// --- Lobby ---

func TestLobby_RequiresAuth(t *testing.T) {
	t.Parallel()

	lobbyRoute := route.Lobby(
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

	lobbyRoute := route.Lobby(
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
	traceCtx := ctx.WithSpan(spanCtx, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-123")

	lobbyRoute := route.Lobby(
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
