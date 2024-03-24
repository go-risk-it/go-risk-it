package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	request2 "github.com/tomfran/go-risk-it/internal/api/game/rest/request"
	"github.com/tomfran/go-risk-it/internal/web/controller"
)

type Handler interface {
	OnMoveDeploy(w http.ResponseWriter, r *http.Request)
}

type HandlerImpl struct {
	moveController controller.MoveController
}

func NewHandler(moveController controller.MoveController) *HandlerImpl {
	return &HandlerImpl{
		moveController: moveController,
	}
}

func (m *HandlerImpl) OnMoveDeploy(writer http.ResponseWriter, req *http.Request) {
	gameIDStr := req.PathValue("id")

	_, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(writer, "invalid game id", http.StatusBadRequest)

		return
	}

	var deployMove request2.DeployMove

	err = json.NewDecoder(req.Body).Decode(&deployMove)
	if err != nil {
		http.Error(writer, "invalid req body", http.StatusBadRequest)

		return
	}
	// remove gameId from deployMove, use custom struct instead

	err = m.moveController.PerformDeployMove(req.Context(), deployMove)
	if err != nil {
		http.Error(writer, "unable to perform deploy move", http.StatusInternalServerError)

		return
	}
}
