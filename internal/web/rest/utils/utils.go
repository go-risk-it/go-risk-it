package restutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
)

// ErrorResponse is the standard JSON error envelope for API responses.
type ErrorResponse struct {
	Error  string `json:"error"`
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func WriteResponse(writer http.ResponseWriter, body []byte, status int) {
	writer.WriteHeader(status)

	_, err := writer.Write(body)
	if err != nil {
		log.Printf("failed to write HTTP response: %v", err)
	}
}

// WriteError maps domain errors to appropriate HTTP status codes and writes a JSON error response.
// For 4xx errors, the error message is sent to the client (these are user-safe domain errors).
// For 5xx errors, a generic message is sent to the client and the original error is returned
// so the caller can log it with request context.
func WriteError(writer http.ResponseWriter, err error) error {
	status, code, clientMsg := mapErrorToResponse(err)
	writeJSONError(writer, status, code, clientMsg)

	if status >= http.StatusInternalServerError {
		return err
	}

	return nil
}

func writeJSONError(writer http.ResponseWriter, status int, code string, msg string) {
	resp := ErrorResponse{
		Error: msg,
		Code:  code,
	}

	body, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		http.Error(writer, msg, status)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	WriteResponse(writer, body, status)
}

func mapErrorToResponse(err error) (int, string, string) {
	var validationErr *domainerrors.ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	}

	var conflictErr *domainerrors.ConflictError
	if errors.As(err, &conflictErr) {
		return http.StatusConflict, "CONFLICT", err.Error()
	}

	var forbiddenErr *domainerrors.ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return http.StatusForbidden, "FORBIDDEN", err.Error()
	}

	return http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred"
}

type malformedRequestError struct {
	status int
	msg    string
}

func (mr *malformedRequestError) Error() string {
	return mr.msg
}

// Validatable is an opt-in interface for request types that need boundary validation.
type Validatable interface {
	Validate() error
}

func DecodeRequest[T any](writer http.ResponseWriter, req *http.Request) (T, error) {
	var result T

	err := decodeJSONBody(writer, req, &result)
	if err != nil {
		var mr *malformedRequestError
		if errors.As(err, &mr) {
			writeJSONError(writer, mr.status, "MALFORMED_REQUEST", mr.msg)
		} else {
			writeJSONError(
				writer,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				http.StatusText(http.StatusInternalServerError),
			)
		}

		return result, err
	}

	if v, ok := any(result).(Validatable); ok {
		if err := v.Validate(); err != nil {
			writeJSONError(writer, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

			return result, err
		}
	}

	return result, nil
}

func decodeJSONBody[T any](writer http.ResponseWriter, req *http.Request, dst T) error {
	ct := req.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-MissionType header is not application/json"

			return &malformedRequestError{status: http.StatusUnsupportedMediaType, msg: msg}
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
			msg := fmt.Sprintf(
				"Request body contains badly-formed JSON (at position %d)",
				syntaxError.Offset,
			)

			return &malformedRequestError{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"

			return &malformedRequestError{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf(
				"Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field,
				unmarshalTypeError.Offset,
			)

			return &malformedRequestError{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := "Request body contains unknown field " + fieldName

			return &malformedRequestError{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"

			return &malformedRequestError{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"

			return &malformedRequestError{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return fmt.Errorf("unexpected error: %w", err)
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"

		return &malformedRequestError{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
