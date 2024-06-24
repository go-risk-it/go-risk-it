package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"go.uber.org/zap"
)

type GameMiddleware interface {
	Wrap(route rest.Route) rest.Route
}

type GameMiddlewareImpl struct {
	log *zap.SugaredLogger
}

func NewGameMiddleware(log *zap.SugaredLogger) GameMiddleware {
	return &GameMiddlewareImpl{log: log}
}

func (g *GameMiddlewareImpl) Wrap(route rest.Route) rest.Route {
	if !strings.HasPrefix(route.Pattern(), "/api/v1/games/{id}/") {
		return route
	}

	return rest.NewRoute(
		route.Pattern(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			gameID, err := extractGameID(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)

				return
			}

			route.ServeHTTP(
				writer,
				request.WithContext(
					context.WithValue(request.Context(), riskcontext.GameIDKey, int64(gameID)),
				),
			)
		}))
}

func extractGameID(req *http.Request) (int, error) {
	gameIDStr := req.PathValue("id")

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		return -1, fmt.Errorf("invalid game id: %w", err)
	}

	return gameID, nil
}
