package restutils_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON_Success(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	payload := struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}{
		Name:  "test",
		Count: 42,
	}

	err := restutils.WriteJSON(recorder, http.StatusOK, payload)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var result map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
	assert.Equal(t, "test", result["name"])
	assert.InDelta(t, 42, result["count"], 0)
}

func TestWriteJSON_CustomStatus(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	payload := struct {
		ID string `json:"id"`
	}{
		ID: "abc-123",
	}

	err := restutils.WriteJSON(recorder, http.StatusCreated, payload)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var result map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
	assert.Equal(t, "abc-123", result["id"])
}

func TestWriteJSON_MarshalFailure(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	err := restutils.WriteJSON(recorder, http.StatusOK, make(chan int))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal response")
	// Response must be untouched on marshal failure
	assert.Empty(t, recorder.Header().Get("Content-Type"))
	assert.Empty(t, recorder.Body.String())
}

func TestWriteJSON_EmptyStruct(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	err := restutils.WriteJSON(recorder, http.StatusOK, struct{}{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "{}", recorder.Body.String())
}
