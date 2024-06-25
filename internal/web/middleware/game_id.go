package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
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
	if !strings.HasPrefix(routeToWrap.Pattern(), "/api/v1/games/{id}") {
		return routeToWrap
	}

	return route.NewRoute(
		routeToWrap.Pattern(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			g.log.Debug("applying game middleware")

			gameID, err := extractGameID(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)

				return
			}

			moveContext, err := buildMoveContext(request, gameID)
			if err != nil {
				http.Error(writer, "cannot build move context", http.StatusInternalServerError)

				return
			}

			routeToWrap.ServeHTTP(
				writer,
				moveContext,
			)
		}))
}

func buildMoveContext(request *http.Request, gameID int64) (*http.Request, error) {
	userContext, ok := request.Context().(ctx.UserContext)
	if !ok {
		return nil, fmt.Errorf("user context not found")
	}

	gameContext := ctx.WithGameID(userContext, gameID)

	return request.WithContext(ctx.NewMoveContext(userContext, gameContext)), nil
}

func extractGameID(req *http.Request) (int64, error) {
	gameIDStr := req.PathValue("id")

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		return -1, fmt.Errorf("invalid game id: %w", err)
	}

	return int64(gameID), nil
}
