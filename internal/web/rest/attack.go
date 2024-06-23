package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type AttackHandler interface {
	Pattern() string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
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
	gameID, err := extractGameID(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	deployMoveRequest, err := decodeRequest[request.AttackMove](writer, req)
	if err != nil {
		return
	}

	err = h.moveController.PerformAttackMove(
		req.Context(),
		int64(gameID),
		"gianfranco",
		deployMoveRequest,
	)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	WriteResponse(writer, h.log, []byte{}, http.StatusNoContent)
}
