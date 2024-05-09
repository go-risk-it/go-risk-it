package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"go.uber.org/zap"
)

type GameHandler interface {
	Pattern() string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type GameHandlerImpl struct {
	log            *zap.SugaredLogger
	gameController controller.GameController
}

func NewGameHandler(
	log *zap.SugaredLogger,
	gameController controller.GameController,
) *GameHandlerImpl {
	return &GameHandlerImpl{
		log:            log,
		gameController: gameController,
	}
}

func (h *GameHandlerImpl) Pattern() string {
	return "/api/1/game"
}

func (h *GameHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
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

	err = WriteResponse(writer, createGameResponse, http.StatusCreated)
	if err != nil {
		h.log.Errorw("unable to write response", "error", err)

		return
	}
}
