package restutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

// ErrorResponse is the standard JSON error envelope for API responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Detail  string `json:"detail,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}

func WriteResponse(writer http.ResponseWriter, body []byte, status int) {
	writer.WriteHeader(status)

	_, err := writer.Write(body)
	if err != nil {
		// No context available here — accept the write failure silently.
		// The span in the calling middleware will capture the error.
		_ = err
	}
}

// WriteError maps domain errors to appropriate HTTP status codes and writes a JSON error response.
// For 4xx errors, the error message is sent to the client (these are user-safe domain errors).
// For 5xx errors, a generic message is sent to the client and the original error is returned
// so the caller can log it with request context.
func WriteError(writer http.ResponseWriter, err error) error {
	return WriteErrorWithTrace(writer, err, "")
}

// WriteErrorWithTrace is like WriteError but includes a trace ID in the response envelope.
func WriteErrorWithTrace(writer http.ResponseWriter, err error, traceID string) error {
	status, code, clientMsg := mapErrorToResponse(err)
	writeJSONErrorWithTrace(writer, status, code, clientMsg, traceID)

	if status >= http.StatusInternalServerError {
		return err
	}

	return nil
}

func writeJSONErrorWithTrace(
	writer http.ResponseWriter,
	status int,
	code string,
	msg string,
	traceID string,
) {
	resp := ErrorResponse{
		Error:   msg,
		Code:    code,
		TraceID: traceID,
	}

	body, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		http.Error(writer, msg, status)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	WriteResponse(writer, body, status)
}

var categoryToHTTP = map[domainerrors.ErrorCategory]int{
	domainerrors.CategoryValidation:   http.StatusBadRequest,
	domainerrors.CategoryUnauthorized: http.StatusUnauthorized,
	domainerrors.CategoryForbidden:    http.StatusForbidden,
	domainerrors.CategoryNotFound:     http.StatusNotFound,
	domainerrors.CategoryConflict:     http.StatusConflict,
}

func mapErrorToResponse(err error) (int, string, string) {
	var categorizable domainerrors.Categorizable
	if errors.As(err, &categorizable) {
		cat := categorizable.Category()
		status, ok := categoryToHTTP[cat]
		if ok {
			return status, cat.String(), err.Error()
		}
	}

	return http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred"
}

// Validatable is an opt-in interface for request types that need boundary validation.
type Validatable interface {
	Validate() error
}

// DecodeRequest decodes the JSON request body into T and runs Validate() if T implements
// Validatable. It returns domain errors (ValidationError for malformed/invalid requests)
// that the error middleware translates into HTTP responses.
func DecodeRequest[T any](writer http.ResponseWriter, req *http.Request) (T, error) {
	var result T

	if err := decodeJSONBody(writer, req, &result); err != nil {
		return result, err
	}

	if v, ok := any(result).(Validatable); ok {
		if err := v.Validate(); err != nil {
			return result, domainerrors.NewValidationError(err.Error())
		}
	}

	return result, nil
}

func decodeJSONBody[T any](writer http.ResponseWriter, req *http.Request, dst T) error {
	ct := req.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return domainerrors.NewValidationError("Content-Type header is not application/json")
		}
	}

	req.Body = http.MaxBytesReader(writer, req.Body, 1048576)

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	err := decode(dec, dst)
	if err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}

	return nil
}

func decode[T any](dec *json.Decoder, dst T) error {
	err := dec.Decode(&dst)
	if err != nil {
		var (
			syntaxError        *json.SyntaxError
			unmarshalTypeError *json.UnmarshalTypeError
		)

		switch {
		case errors.As(err, &syntaxError):
			return domainerrors.NewValidationError(fmt.Sprintf(
				"Request body contains badly-formed JSON (at position %d)",
				syntaxError.Offset,
			))

		case errors.Is(err, io.ErrUnexpectedEOF):
			return domainerrors.NewValidationError("Request body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			return domainerrors.NewValidationError(fmt.Sprintf(
				"Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field,
				unmarshalTypeError.Offset,
			))

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")

			return domainerrors.NewValidationError(
				"Request body contains unknown field " + fieldName,
			)

		case errors.Is(err, io.EOF):
			return domainerrors.NewValidationError("Request body must not be empty")

		case err.Error() == "http: request body too large":
			return domainerrors.NewValidationError("Request body must not be larger than 1MB")

		default:
			return fmt.Errorf("unexpected error: %w", err)
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return domainerrors.NewValidationError(
			"Request body must only contain a single JSON object",
		)
	}

	return nil
}
