package annotations_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotator_Annotate_SendsCorrectPayload(t *testing.T) {
	t.Parallel()

	var received map[string]interface{}

	server := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "/api/annotations", request.URL.Path)
			assert.Equal(t, "POST", request.Method)

			require.NoError(
				t,
				json.NewDecoder(request.Body).Decode(&received),
			)

			responseWriter.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	annotator := annotations.NewAnnotator(server.URL)
	annotator.Annotate("test started", "perf-test", "phase")

	assert.Equal(t, "test started", received["text"])

	tags, ok := received["tags"].([]interface{})
	require.True(t, ok, "tags should be an array")
	assert.Contains(t, tags, "perf-test")
	assert.Contains(t, tags, "phase")
	assert.NotZero(t, received["time"])
}

func TestAnnotator_Annotate_NoOpWhenDisabled(t *testing.T) {
	t.Parallel()

	called := false

	server := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			called = true
		}),
	)
	defer server.Close()

	annotator := annotations.NewAnnotator("") // empty = disabled
	annotator.Annotate("should not send", "tag")

	assert.False(t, called)
}
