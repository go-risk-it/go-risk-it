package restutils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteJSON marshals payload as JSON and writes it to the response with the given status code.
// It sets Content-Type to application/json before writing.
// Returns an error if JSON marshaling fails; write failures are silently ignored
// (consistent with WriteResponse — the calling middleware's span captures write errors).
func WriteJSON[T any](writer http.ResponseWriter, status int, payload T) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	writer.Header().Set("Content-Type", "application/json")
	WriteResponse(writer, body, status)

	return nil
}
