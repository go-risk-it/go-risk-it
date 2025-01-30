package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type Handler interface {
	route.Route
}

type HandlerImpl struct {
	gameController controller.GameController
}

var _ Handler = (*HandlerImpl)(nil)

func NewCreationHandler(gameController controller.GameController) *HandlerImpl {
	return &HandlerImpl{
		gameController: gameController,
	}
}

func (h *HandlerImpl) Pattern() string {
	return "/api/v1/games"
}

func (h *HandlerImpl) RequiresAuth() bool {
	return true
}

func (h *HandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	createGameRequest, err := restutils.DecodeRequest[request.CreateGame](writer, req)
	if err != nil {
		return
	}

	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		http.Error(
			writer,
			fmt.Sprintf("invalid user context in route: %v", h.Pattern()),
			http.StatusInternalServerError,
		)

		return
	}

	gameID, err := h.gameController.CreateGame(userContext, createGameRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	createGameResponse, err := json.Marshal(response.CreateGame{GameID: gameID})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, createGameResponse, http.StatusCreated)
}
