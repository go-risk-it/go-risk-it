package rest

import (
	"errors"
	"log"
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

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(writer, "invalid game id", http.StatusBadRequest)

		return
	}

	var deployMove request2.DeployMove

	err = decodeJSONBody(writer, req, &deployMove)
	if err != nil {
		var mr *malformedRequestError
		if errors.As(err, &mr) {
			http.Error(writer, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(
				writer,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}

		return
	}

	err = m.moveController.PerformDeployMove(req.Context(), int64(gameID), deployMove)
	if err != nil {
		http.Error(writer, "unable to perform deploy move", http.StatusInternalServerError)

		return
	}
}
