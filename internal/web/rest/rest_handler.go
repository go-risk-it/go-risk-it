package rest

import (
	"fmt"
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
	gameID, err := extractGameID(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

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

	m.log.Infow("writing successful http response with no content")

	err = writeNoContentResponse(writer, http.StatusNoContent)
	if err != nil {
		m.log.Errorw("unable to write response", "error", err)

		return
	}

	m.log.Infow("successfully wrote http response with no content")
}

func writeNoContentResponse(writer http.ResponseWriter, status int) error {
	writer.WriteHeader(status)

	_, err := writer.Write([]byte{})
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

func extractGameID(req *http.Request) (int, error) {
	gameIDStr := req.PathValue("id")

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid game id: %w", err)
	}

	return gameID, nil
}
