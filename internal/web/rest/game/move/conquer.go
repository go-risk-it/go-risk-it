package move

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type ConquerHandler interface {
	route.Route
}

type ConquerHandlerImpl struct {
	moveController controller.MoveController
}

var _ ConquerHandler = (*ConquerHandlerImpl)(nil)

func NewConquerHandler(moveController controller.MoveController) *ConquerHandlerImpl {
	return &ConquerHandlerImpl{
		moveController: moveController,
	}
}

func (h *ConquerHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/conquers"
}

func (h *ConquerHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *ConquerHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	serveMove[request.ConquerMove](writer, req, h.moveController.PerformConquerMove)
}
