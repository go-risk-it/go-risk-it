package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type malformedRequestError struct {
	status int
	msg    string
}

func (mr *malformedRequestError) Error() string {
	return mr.msg
}

func decodeRequest[T any](writer http.ResponseWriter, req *http.Request) (T, error) {
	var result T

	err := decodeJSONBody(writer, req, &result)
	if err != nil {
		var mr *malformedRequestError
		if errors.As(err, &mr) {
			http.Error(writer, mr.msg, mr.status)
		} else {
			http.Error(
				writer,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}

		return result, err
	}

	return result, nil
}

func decodeJSONBody[T any](writer http.ResponseWriter, req *http.Request, dst T) error {
	ct := req.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"

			return &malformedRequestError{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	req.Body = http.MaxBytesReader(writer, req.Body, 1048576)

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	err := decode(dec, dst)
	if err != nil {
		return fmt.Errorf("failed to decode request body: %writer", err)
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
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)

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

func WriteResponse(writer http.ResponseWriter, log *zap.SugaredLogger, body []byte, status int) {
	writer.WriteHeader(status)

	log.Debugw("writing response", "status", status, "body", string(body))

	_, err := writer.Write(body)
	if err != nil {
		log.Errorw("unable to write response", "error", err)
	}
}

func extractGameID(req *http.Request) (int64, error) {
	if gameID, ok := req.Context().Value(riskcontext.GameIDKey).(int64); ok {
		return gameID, nil
	}

	return -1, fmt.Errorf("invalid game id")
}

func extractUserID(req *http.Request) (string, error) {
	if userID, ok := req.Context().Value(riskcontext.UserIDKey).(string); ok {
		return userID, nil
	}

	return "", fmt.Errorf("invalid user id")
}

var Module = fx.Options(
	fx.Provide(
		AsRoute(NewDeployHandler),
		AsRoute(NewAttackHandler),
		AsRoute(NewGameHandler),
		AsRoute(NewWebSocketHandler),
	),
)
