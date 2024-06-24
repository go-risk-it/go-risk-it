package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
)

type AttackHandler interface {
	Route
}

type AttackHandlerImpl struct {
	moveController controller.MoveController
}

var _ AttackHandler = (*AttackHandlerImpl)(nil)

func NewAttackHandler(moveController controller.MoveController) *AttackHandlerImpl {
	return &AttackHandlerImpl{
		moveController: moveController,
	}
}

func (h *AttackHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/attacks"
}

func (h *AttackHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	serveMove[request.AttackMove](writer, req, h.moveController.PerformAttackMove)
}
