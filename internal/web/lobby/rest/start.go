package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type StartHandler interface {
	route.Route
}

type StartHandlerImpl struct {
	startController controller.StartController
}

var _ StartHandler = (*StartHandlerImpl)(nil)

func NewStartHandler(
	startController controller.StartController,
) *StartHandlerImpl {
	return &StartHandlerImpl{
		startController: startController,
	}
}

func (h *StartHandlerImpl) Pattern() string {
	return "/api/v1/lobbies/{id}/start"
}

func (h *StartHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *StartHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	lobbyContext, ok := req.Context().(ctx.LobbyContext)
	if !ok {
		http.Error(writer, "an internal error occurred", http.StatusInternalServerError)

		return
	}

	if err := h.startController.StartGame(lobbyContext); err != nil {
		if logErr := restutils.WriteError(writer, err); logErr != nil {
			lobbyContext.Log().Errorw("request failed", "error", logErr)
		}

		return
	}

	writer.WriteHeader(http.StatusOK)
}
