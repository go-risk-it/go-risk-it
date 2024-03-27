package rest

import (
	"net/http"
	"strconv"

	"github.com/tomfran/go-risk-it/internal/api/game/rest/request"
	"github.com/tomfran/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type Handler interface {
	OnMoveDeploy(w http.ResponseWriter, r *http.Request)
}

type HandlerImpl struct {
	log            *zap.SugaredLogger
	moveController controller.MoveController
}

func NewHandler(
	log *zap.SugaredLogger,
	moveController controller.MoveController,
) *HandlerImpl {
	return &HandlerImpl{
		log:            log,
		moveController: moveController,
	}
}

func (m *HandlerImpl) OnMoveDeploy(writer http.ResponseWriter, req *http.Request) {
	gameIDStr := req.PathValue("id")

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(writer, "invalid game id", http.StatusBadRequest)

		return
	}

	deployMove, err := decodeRequest[request.DeployMove](writer, req)
	if err != nil {
		return
	}

	err = m.moveController.PerformDeployMove(req.Context(), int64(gameID), deployMove)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	m.log.Infow("writing http response")

	writer.WriteHeader(http.StatusNoContent)

	_, err = writer.Write([]byte{})
	if err != nil {
		m.log.Errorw("failed to write response", "err", err)

		return
	}

	m.log.Infow("successfully wrote http response")
}
