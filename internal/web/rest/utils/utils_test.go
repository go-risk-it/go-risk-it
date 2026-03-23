package restutils_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteError_ValidationError_Returns400AndNil(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewValidationError("invalid input")

	logErr := restutils.WriteError(recorder, err)

	require.NoError(t, logErr)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid input")
	assert.Contains(t, recorder.Body.String(), "VALIDATION_ERROR")
}

func TestWriteError_ConflictError_Returns409AndNil(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewConflictError("wrong phase")

	logErr := restutils.WriteError(recorder, err)

	require.NoError(t, logErr)
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "wrong phase")
	assert.Contains(t, recorder.Body.String(), "CONFLICT")
}

func TestWriteError_ForbiddenError_Returns403AndNil(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewForbiddenError("not your turn")

	logErr := restutils.WriteError(recorder, err)

	require.NoError(t, logErr)
	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "not your turn")
	assert.Contains(t, recorder.Body.String(), "FORBIDDEN")
}

func TestWriteError_InternalError_Returns500WithGenericMessageAndOriginalError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := errors.New("database connection failed: host=db.internal port=5432")

	logErr := restutils.WriteError(recorder, err)

	require.Error(t, logErr)
	assert.Equal(t, err, logErr)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	// Client must NOT see internal details
	assert.NotContains(t, recorder.Body.String(), "database")
	assert.NotContains(t, recorder.Body.String(), "db.internal")
	assert.Contains(t, recorder.Body.String(), "an internal error occurred")
	assert.Contains(t, recorder.Body.String(), "INTERNAL_ERROR")
}

func TestWriteError_NotFoundError_Returns404AndNil(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewNotFoundError("game not found")

	logErr := restutils.WriteError(recorder, err)

	require.NoError(t, logErr)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "game not found")
	assert.Contains(t, recorder.Body.String(), "NOT_FOUND")
}

func TestWriteError_UnauthorizedError_Returns401AndNil(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewUnauthorizedError("invalid token")

	logErr := restutils.WriteError(recorder, err)

	require.NoError(t, logErr)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid token")
	assert.Contains(t, recorder.Body.String(), "UNAUTHORIZED")
}

func TestWriteErrorWithTrace_IncludesTraceID(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewValidationError("bad input")

	logErr := restutils.WriteErrorWithTrace(recorder, err, "abc123def456")

	require.NoError(t, logErr)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "abc123def456", resp.TraceID)
	assert.Equal(t, "bad input", resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
}

func TestWriteErrorWithTrace_EmptyTraceID_OmitsField(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := domainerrors.NewConflictError("wrong phase")

	logErr := restutils.WriteErrorWithTrace(recorder, err, "")

	require.NoError(t, logErr)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	_, hasTraceID := resp["traceId"]
	assert.False(t, hasTraceID, "traceId should be omitted when empty")
}

func TestWriteErrorWithTrace_InternalError_NoMessageLeak(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := errors.New("secret DB password: p4ssw0rd")

	logErr := restutils.WriteErrorWithTrace(recorder, err, "trace-123")

	require.Error(t, logErr)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "trace-123", resp.TraceID)
	assert.Equal(t, "an internal error occurred", resp.Error)
	assert.NotContains(t, recorder.Body.String(), "p4ssw0rd")
}
