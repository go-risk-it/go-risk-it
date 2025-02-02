package rest

import (
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type JoinHandler interface {
	route.Route
}

type JoinHandlerImpl struct {
	managementController controller.ManagementController
}

var _ JoinHandler = (*JoinHandlerImpl)(nil)

func NewJoinHandler(
	managementController controller.ManagementController,
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
	joinLobbyRequest, err := restutils.DecodeRequest[request.JoinLobby](writer, req)
	if err != nil {
		return
	}

	lobbyContext, ok := req.Context().(ctx.LobbyContext)
	if !ok {
		http.Error(
			writer,
			fmt.Sprintf("invalid user context in route: %v", h.Pattern()),
			http.StatusInternalServerError,
		)

		return
	}

	if err := h.managementController.JoinLobby(lobbyContext, joinLobbyRequest); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
