package move

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		route.AsRoute(NewDeployHandler),
		route.AsRoute(NewAttackHandler),
	),
)

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

func serveMove[T any](
	writer http.ResponseWriter,
	req *http.Request,
	perform func(ctx context.Context, gameID int64, userID string, move T) error,
) {
	moveRequest, err := restutils.DecodeRequest[T](writer, req)
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

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
