package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/fx"
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

func WriteResponse(writer http.ResponseWriter, body []byte, status int) error {
	writer.WriteHeader(status)

	_, err := writer.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

func extractGameID(req *http.Request) (int, error) {
	gameIDStr := req.PathValue("id")

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid game id: %w", err)
	}

	return gameID, nil
}

var Module = fx.Options(
	fx.Provide(
		AsRoute(NewDeployHandler),
		AsRoute(NewGameHandler),
		AsRoute(NewWebSocketHandler),
		fx.Annotate(
			NewServeMux,
			fx.ParamTags(`group:"routes"`),
		),
	),
)
