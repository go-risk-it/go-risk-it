package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type Handler interface {
	OnMoveDeploy(w http.ResponseWriter, r *http.Request)
	OnCreateGame(w http.ResponseWriter, r *http.Request)
}

type HandlerImpl struct {
	log            *zap.SugaredLogger
	moveController controller.MoveController
	gameController controller.GameController
}

func NewHandler(
	log *zap.SugaredLogger,
	moveController controller.MoveController,
	gameController controller.GameController,
) *HandlerImpl {
	return &HandlerImpl{
		log:            log,
		moveController: moveController,
		gameController: gameController,
	}
}

func (h *HandlerImpl) OnMoveDeploy(writer http.ResponseWriter, req *http.Request) {
	gameID, err := extractGameID(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	deployMoveRequest, err := decodeRequest[request.DeployMove](writer, req)
	if err != nil {
		return
	}

	err = h.moveController.PerformDeployMove(req.Context(), int64(gameID), deployMoveRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	err = writeResponse(writer, []byte{}, http.StatusNoContent)
	if err != nil {
		h.log.Errorw("unable to write response", "error", err)

		return
	}
}

func (h *HandlerImpl) OnCreateGame(writer http.ResponseWriter, req *http.Request) {
	createGameRequest, err := decodeRequest[request.CreateGame](writer, req)
	if err != nil {
		return
	}

	gameID, err := h.gameController.CreateGame(req.Context(), createGameRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	createGameResponse, err := json.Marshal(response.CreateGame{GameID: gameID})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	err = writeResponse(writer, createGameResponse, http.StatusCreated)
	if err != nil {
		h.log.Errorw("unable to write response", "error", err)

		return
	}
}
