package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type ManagementHandler interface {
	route.Route
}

type ManagementHandlerImpl struct {
	gameController controller.GameController
}

var _ ManagementHandler = (*ManagementHandlerImpl)(nil)

func NewManagementHandler(
	gameController controller.GameController,
) *ManagementHandlerImpl {
	return &ManagementHandlerImpl{
		gameController: gameController,
	}
}

func (h *ManagementHandlerImpl) Pattern() string {
	return "/api/v1/games/summary"
}

func (h *ManagementHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *ManagementHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		http.Error(
			writer,
			fmt.Sprintf("invalid user context in route: %v", h.Pattern()),
			http.StatusInternalServerError,
		)

		return
	}

	games, err := h.gameController.GetUserGames(userContext)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	gamesResponse, err := json.Marshal(games)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, gamesResponse, http.StatusOK)
}
