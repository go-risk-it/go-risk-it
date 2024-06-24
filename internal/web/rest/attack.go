package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type AttackHandler interface {
	Route
}

type AttackHandlerImpl struct {
	log            *zap.SugaredLogger
	moveController controller.MoveController
}

func NewAttackHandler(
	log *zap.SugaredLogger,
	moveController controller.MoveController,
) *AttackHandlerImpl {
	return &AttackHandlerImpl{
		log:            log,
		moveController: moveController,
	}
}

func (h *AttackHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/attacks"
}

func (h *AttackHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	serveMove[request.AttackMove](h.log, writer, req, h.moveController.PerformAttackMove)
}
