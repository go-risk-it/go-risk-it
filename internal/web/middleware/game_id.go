package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type GameMiddleware interface {
	Middleware
}

type GameMiddlewareImpl struct {
	log *zap.SugaredLogger
}

var _ GameMiddleware = (*GameMiddlewareImpl)(nil)

func NewGameMiddleware(log *zap.SugaredLogger) GameMiddleware {
	return &GameMiddlewareImpl{log: log}
}

func (g *GameMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	if !strings.HasPrefix(routeToWrap.Pattern(), "/api/v1/games/{id}/") {
		return routeToWrap
	}

	return route.NewRoute(
		routeToWrap.Pattern(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			gameID, err := extractGameID(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)

				return
			}

			routeToWrap.ServeHTTP(
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
