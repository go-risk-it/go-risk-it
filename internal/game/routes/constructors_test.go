package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/routes"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// --- Game ---

func TestGame_RequiresAuth(t *testing.T) {
	t.Parallel()

	gameRoute := routes.Game(
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

	gameRoute := routes.Game(
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

	gameRoute := routes.Game(
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

	gameRoute := routes.Game(
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

	gameRoute := routes.Game(
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

// --- GameWS ---

func TestGameWS_IsWebSocket(t *testing.T) {
	t.Parallel()

	wsRoute := routes.GameWS(
		"GET /api/v1/games/{id}/ws",
		func(_ http.ResponseWriter, _ *http.Request, _ ctx.GameContext) error {
			return nil
		},
	)

	assert.True(t, wsRoute.IsWebSocket())
	assert.True(t, wsRoute.RequiresAuth())
}
