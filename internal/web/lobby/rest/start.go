package rest

import (
	"errors"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
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
	middleware.HandleErrors(h.handle).ServeHTTP(writer, req)
}

func (h *StartHandlerImpl) handle(writer http.ResponseWriter, req *http.Request) error {
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
