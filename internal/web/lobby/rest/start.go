package rest

import (
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
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
		http.Error(
			writer,
			fmt.Sprintf("invalid user context in route: %v", h.Pattern()),
			http.StatusInternalServerError,
		)

		return
	}

	if err := h.startController.StartGame(lobbyContext); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
