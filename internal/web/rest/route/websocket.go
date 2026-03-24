package route

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// ExtractWSToken extracts the JWT from the Sec-WebSocket-Protocol header and sets it as a
// Bearer token in the Authorization header. This is needed because the browser WebSocket API
// does not support custom headers — the token is smuggled via subprotocol.
func ExtractWSToken(request *http.Request) {
	subprotocol := request.Header.Get("Sec-WebSocket-Protocol")
	if subprotocol == "" {
		return
	}

	token, err := parseWSSubprotocol(subprotocol)
	if err != nil {
		slog.ErrorContext(
			request.Context(),
			"unable to extract token from subprotocol",
			"error", err,
		)

		return
	}

	request.Header.Set("Authorization", "Bearer "+token)
}

func parseWSSubprotocol(subprotocol string) (string, error) {
	if !strings.HasPrefix(subprotocol, "risk-it.websocket.auth.token, ") {
		return "", errors.New("invalid subprotocol")
	}

	return strings.TrimPrefix(subprotocol, "risk-it.websocket.auth.token, "), nil
}
