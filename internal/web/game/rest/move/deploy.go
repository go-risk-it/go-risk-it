package move

import (
	"log/slog"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type DeployHandlerImpl struct {
	moveController *controller.MoveController
}

var _ route.Route = (*DeployHandlerImpl)(nil)

func NewDeployHandler(moveController *controller.MoveController) *DeployHandlerImpl {
	return &DeployHandlerImpl{
		moveController: moveController,
	}
}

func (h *DeployHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/deployments"
}

func (h *DeployHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	slog.InfoContext(req.Context(), "deploy move received (slog spike)")
	handleMove[request.DeployMove](writer, req, h.moveController.PerformDeployMove)
}

func (h *DeployHandlerImpl) RequiresAuth() bool {
	return true
}
