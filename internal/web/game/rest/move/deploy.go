package move

import (
	"log/slog"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

func NewDeployHandler(moveController *controller.MoveController) *route.Route {
	return route.New(
		"/api/v1/games/{id}/moves/deployments",
		true,
		http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			slog.InfoContext(req.Context(), "deploy move received (slog spike)")
			handleMove[request.DeployMove](writer, req, moveController.PerformDeployMove)
		}),
	)
}
