package route_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

// --- test types ---

type createRequest struct {
	Name string `json:"name"`
}

type createResponse struct {
	ID int64 `json:"id"`
}

type queryResponse struct {
	Items []string `json:"items"`
}

// --- CreateHandler ---

func TestCreateHandler_Success(t *testing.T) {
	t.Parallel()

	createRoute := route.CreateHandler(
		"POST /api/v1/things",
		func(_ kernelctx.UserContext, req createRequest) (createResponse, error) {
			return createResponse{ID: 99}, nil
		},
	)

	assert.True(t, createRoute.RequiresAuth())
	assert.Equal(t, "POST /api/v1/things", createRoute.Pattern())

	userCtx := newUserContext(t)
	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequestWithContext(userCtx, http.MethodPost, "/api/v1/things", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp createResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, int64(99), resp.ID)
}

func TestCreateHandler_MissingAuth(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	createRoute := route.CreateHandler(
		"POST /api/v1/things",
		func(_ kernelctx.UserContext, _ createRequest) (createResponse, error) {
			t.Fatal("perform should not be called")

			return createResponse{}, nil
		},
	)

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequestWithContext(spanCtx, http.MethodPost, "/api/v1/things", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp.Code)
}

func TestCreateHandler_InvalidBody(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	createRoute := route.CreateHandler(
		"POST /api/v1/things",
		func(_ kernelctx.UserContext, _ createRequest) (createResponse, error) {
			t.Fatal("perform should not be called")

			return createResponse{}, nil
		},
	)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequestWithContext(userCtx, http.MethodPost, "/api/v1/things", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestCreateHandler_HandlerError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	createRoute := route.CreateHandler(
		"POST /api/v1/things",
		func(_ kernelctx.UserContext, _ createRequest) (createResponse, error) {
			return createResponse{}, domainerrors.NewConflictError("already exists")
		},
	)

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequestWithContext(userCtx, http.MethodPost, "/api/v1/things", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "CONFLICT", resp.Code)
}

func TestCreateHandler_InternalError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	createRoute := route.CreateHandler(
		"POST /api/v1/things",
		func(_ kernelctx.UserContext, _ createRequest) (createResponse, error) {
			return createResponse{}, errors.New("db connection lost")
		},
	)

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequestWithContext(userCtx, http.MethodPost, "/api/v1/things", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	assert.NotContains(t, rec.Body.String(), "db connection lost")
}

// --- QueryHandler ---

func TestQueryHandler_Success(t *testing.T) {
	t.Parallel()

	queryRoute := route.QueryHandler(
		"GET /api/v1/things",
		func(_ kernelctx.UserContext) (queryResponse, error) {
			return queryResponse{Items: []string{"a", "b"}}, nil
		},
	)

	assert.True(t, queryRoute.RequiresAuth())
	assert.Equal(t, "GET /api/v1/things", queryRoute.Pattern())

	userCtx := newUserContext(t)
	req := httptest.NewRequestWithContext(userCtx, http.MethodGet, "/api/v1/things", nil)
	rec := httptest.NewRecorder()

	queryRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp queryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, []string{"a", "b"}, resp.Items)
}

func TestQueryHandler_MissingAuth(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	queryRoute := route.QueryHandler(
		"GET /api/v1/things",
		func(_ kernelctx.UserContext) (queryResponse, error) {
			t.Fatal("query should not be called")

			return queryResponse{}, nil
		},
	)

	req := httptest.NewRequestWithContext(spanCtx, http.MethodGet, "/api/v1/things", nil)
	rec := httptest.NewRecorder()

	queryRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp.Code)
}

func TestQueryHandler_HandlerError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	queryRoute := route.QueryHandler(
		"GET /api/v1/things",
		func(_ kernelctx.UserContext) (queryResponse, error) {
			return queryResponse{}, domainerrors.NewNotFoundError("no things")
		},
	)

	req := httptest.NewRequestWithContext(userCtx, http.MethodGet, "/api/v1/things", nil)
	rec := httptest.NewRecorder()

	queryRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

// --- DomainCommand ---

type testHandlerCtx struct {
	userID string
	id     int64
}

type commandRequest struct {
	Value int `json:"value"`
}

func buildTestHandlerCtx(userCtx kernelctx.UserContext, id int64) testHandlerCtx {
	return testHandlerCtx{userID: userCtx.UserID(), id: id}
}

func TestDomainCommand_Success(t *testing.T) {
	t.Parallel()

	var captured testHandlerCtx

	var capturedReq commandRequest

	commandRoute := route.DomainCommand(
		"POST /api/v1/things/{id}/action",
		buildTestHandlerCtx,
		func(ctx testHandlerCtx, req commandRequest) error {
			captured = ctx
			capturedReq = req

			return nil
		},
	)

	assert.True(t, commandRoute.RequiresAuth())

	userCtx := newUserContext(t)
	body := strings.NewReader(`{"value":42}`)
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/7/action",
		body,
	)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	commandRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, "user-123", captured.userID)
	assert.Equal(t, int64(7), captured.id)
	assert.Equal(t, 42, capturedReq.Value)
}

func TestDomainCommand_InvalidID(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	commandRoute := route.DomainCommand(
		"POST /api/v1/things/{id}/action",
		buildTestHandlerCtx,
		func(_ testHandlerCtx, _ commandRequest) error {
			t.Fatal("perform should not be called")

			return nil
		},
	)

	body := strings.NewReader(`{"value":42}`)
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/abc/action",
		body,
	)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	commandRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestDomainCommand_InvalidBody(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	commandRoute := route.DomainCommand(
		"POST /api/v1/things/{id}/action",
		buildTestHandlerCtx,
		func(_ testHandlerCtx, _ commandRequest) error {
			t.Fatal("perform should not be called")

			return nil
		},
	)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/7/action",
		body,
	)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	commandRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestDomainCommand_HandlerError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	commandRoute := route.DomainCommand(
		"POST /api/v1/things/{id}/action",
		buildTestHandlerCtx,
		func(_ testHandlerCtx, _ commandRequest) error {
			return domainerrors.NewForbiddenError("not your turn")
		},
	)

	body := strings.NewReader(`{"value":42}`)
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/7/action",
		body,
	)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	commandRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "FORBIDDEN", resp.Code)
}

// --- DomainVoid ---

func TestDomainVoid_Success(t *testing.T) {
	t.Parallel()

	var captured testHandlerCtx

	voidRoute := route.DomainVoid(
		"POST /api/v1/things/{id}/advance",
		buildTestHandlerCtx,
		func(ctx testHandlerCtx) error {
			captured = ctx

			return nil
		},
	)

	assert.True(t, voidRoute.RequiresAuth())

	userCtx := newUserContext(t)
	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/5/advance",
		nil,
	)
	req.SetPathValue("id", "5")
	rec := httptest.NewRecorder()

	voidRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, "user-123", captured.userID)
	assert.Equal(t, int64(5), captured.id)
}

func TestDomainVoid_InvalidID(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	voidRoute := route.DomainVoid(
		"POST /api/v1/things/{id}/advance",
		buildTestHandlerCtx,
		func(_ testHandlerCtx) error {
			t.Fatal("perform should not be called")

			return nil
		},
	)

	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/abc/advance",
		nil,
	)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	voidRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestDomainVoid_HandlerError(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)
	userCtx := newUserContextWithTrace(t, spanCtx)

	voidRoute := route.DomainVoid(
		"POST /api/v1/things/{id}/advance",
		buildTestHandlerCtx,
		func(_ testHandlerCtx) error {
			return domainerrors.NewConflictError("wrong phase")
		},
	)

	req := httptest.NewRequestWithContext(
		userCtx,
		http.MethodPost,
		"/api/v1/things/5/advance",
		nil,
	)
	req.SetPathValue("id", "5")
	rec := httptest.NewRecorder()

	voidRoute.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "CONFLICT", resp.Code)
}

// --- helpers ---

// newUserContextWithTrace creates a UserContext from an existing traced context,
// so both OTel tracing and UserContext are available for error mapping tests.
func newUserContextWithTrace(t *testing.T, tracedCtx context.Context) kernelctx.UserContext {
	t.Helper()

	tc := kernelctx.WithSpan(tracedCtx, trace.SpanFromContext(tracedCtx))

	return kernelctx.WithUserID(tc, "user-123")
}
