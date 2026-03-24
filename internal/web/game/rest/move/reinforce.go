package move

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type ReinforceHandlerImpl struct {
	moveController *controller.MoveController
}

var _ route.Route = (*ReinforceHandlerImpl)(nil)

func NewReinforceHandler(moveController *controller.MoveController) *ReinforceHandlerImpl {
	return &ReinforceHandlerImpl{
		moveController: moveController,
	}
}

func (h *ReinforceHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/reinforcements"
}

func (h *ReinforceHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *ReinforceHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	handleMove[request.ReinforceMove](writer, req, h.moveController.PerformReinforceMove)
}
