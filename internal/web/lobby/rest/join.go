package rest

import (
	"errors"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type JoinHandlerImpl struct {
	managementController *controller.ManagementController
}

var _ route.Route = (*JoinHandlerImpl)(nil)

func NewJoinHandler(
	managementController *controller.ManagementController,
) *JoinHandlerImpl {
	return &JoinHandlerImpl{
		managementController: managementController,
	}
}

func (h *JoinHandlerImpl) Pattern() string {
	return "/api/v1/lobbies/{id}/join"
}

func (h *JoinHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *JoinHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	middleware.HandleErrors(h.handle).ServeHTTP(writer, req)
}

func (h *JoinHandlerImpl) handle(writer http.ResponseWriter, req *http.Request) error {
	joinLobbyRequest, err := restutils.DecodeRequest[request.JoinLobby](writer, req)
	if err != nil {
		return err
	}

	lobbyContext, ok := req.Context().(ctx.LobbyContext)
	if !ok {
		return errors.New("invalid lobby context")
	}

	if err := h.managementController.JoinLobby(lobbyContext, joinLobbyRequest); err != nil {
		return err
	}

	writer.WriteHeader(http.StatusNoContent)

	return nil
}
