package move

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type ReinforceHandler interface {
	route.Route
}

type ReinforceHandlerImpl struct {
	moveController controller.MoveController
}

var _ ReinforceHandler = (*ReinforceHandlerImpl)(nil)

func NewReinforceHandler(moveController controller.MoveController) *ReinforceHandlerImpl {
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
	serveMove[request.ReinforceMove](writer, req, h.moveController.PerformReinforceMove)
}
