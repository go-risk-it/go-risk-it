package rest

import (
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
		http.Error(writer, "an internal error occurred", http.StatusInternalServerError)

		return
	}

	if err := h.managementController.JoinLobby(lobbyContext, joinLobbyRequest); err != nil {
		if logErr := restutils.WriteError(writer, err); logErr != nil {
			lobbyContext.Log().Errorw("request failed", "error", logErr)
		}

		return
	}

	writer.WriteHeader(http.StatusOK)
}
