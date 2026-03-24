package rest

import (
	"errors"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

func NewStartHandler(
	startController *controller.StartController,
) *route.Route {
	h := &startHandler{
		startController: startController,
	}

	return route.New("/api/v1/lobbies/{id}/start", true, middleware.HandleErrors(h.handle))
}

type startHandler struct {
	startController *controller.StartController
}

func (h *startHandler) handle(writer http.ResponseWriter, req *http.Request) error {
	lobbyContext, ok := req.Context().(ctx.LobbyContext)
	if !ok {
		return errors.New("invalid lobby context")
	}

	if err := h.startController.StartGame(lobbyContext); err != nil {
		return err
	}

	writer.WriteHeader(http.StatusNoContent)

	return nil
}
