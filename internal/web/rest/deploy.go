package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type DeployHandler interface {
	Pattern() string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type DeployHandlerImpl struct {
	log            *zap.SugaredLogger
	moveController controller.MoveController
}

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
	gameID, err := extractGameID(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	deployMoveRequest, err := decodeRequest[request.DeployMove](writer, req)
	if err != nil {
		return
	}

	err = h.moveController.PerformDeployMove(
		req.Context(),
		int64(gameID),
		deployMoveRequest.UserID,
		deployMoveRequest,
	)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	WriteResponse(writer, h.log, []byte{}, http.StatusNoContent)
}
