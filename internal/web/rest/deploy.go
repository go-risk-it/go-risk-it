package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type DeployHandler interface {
	Route
}

type DeployHandlerImpl struct {
	log            *zap.SugaredLogger
	moveController controller.MoveController
}

var _ DeployHandler = (*DeployHandlerImpl)(nil)

func NewDeployHandler(
	log *zap.SugaredLogger,
	moveController controller.MoveController,
) *DeployHandlerImpl {
	return &DeployHandlerImpl{
		log:            log,
		moveController: moveController,
	}
}

func (h *DeployHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/deployments"
}

func (h *DeployHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	serveMove[request.DeployMove](h.log, writer, req, h.moveController.PerformDeployMove)
}
