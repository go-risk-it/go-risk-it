package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/assert"
)

func TestExtractWSToken_ValidSubprotocol(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ws", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "risk-it.websocket.auth.token, my-jwt-token")

	route.ExtractWSToken(req)

	assert.Equal(t, "Bearer my-jwt-token", req.Header.Get("Authorization"))
}

func TestExtractWSToken_NoSubprotocol(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ws", nil)

	route.ExtractWSToken(req)

	assert.Empty(t, req.Header.Get("Authorization"))
}

func TestExtractWSToken_InvalidSubprotocol(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ws", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "some-other-protocol")

	route.ExtractWSToken(req)

	assert.Empty(t, req.Header.Get("Authorization"))
}
