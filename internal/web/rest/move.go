package rest

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

func serveMove[T any](
	log *zap.SugaredLogger,
	writer http.ResponseWriter,
	req *http.Request,
	perform func(ctx context.Context, gameID int64, userID string, move T) error,
) {
	moveRequest, err := decodeRequest[T](writer, req)
	if err != nil {
		return
	}

	gameID, err := extractGameID(req)
	if err != nil {
		return
	}

	userID, err := extractUserID(req)
	if err != nil {
		return
	}

	if err := perform(req.Context(), gameID, userID, moveRequest); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	WriteResponse(writer, log, []byte{}, http.StatusNoContent)
}
