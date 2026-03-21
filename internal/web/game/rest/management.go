package rest

import (
	"encoding/json"
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
		http.Error(writer, "an internal error occurred", http.StatusInternalServerError)

		return
	}

	games, err := h.gameController.GetUserGames(userContext)
	if err != nil {
		if logErr := restutils.WriteError(writer, err); logErr != nil {
			userContext.Log().Errorw("request failed", "error", logErr)
		}

		return
	}

	gamesResponse, err := json.Marshal(games)
	if err != nil {
		http.Error(writer, "an internal error occurred", http.StatusInternalServerError)
		userContext.Log().Errorw("failed to marshal response", "error", err)

		return
	}

	restutils.WriteResponse(writer, gamesResponse, http.StatusOK)
}
