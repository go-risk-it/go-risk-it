package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type CreationHandler interface {
	route.Route
}

type CreationHandlerImpl struct {
	creationController controller.CreationController
}

var _ CreationHandler = (*CreationHandlerImpl)(nil)

func NewCreationHandler(
	creationController controller.CreationController,
) *CreationHandlerImpl {
	return &CreationHandlerImpl{
		creationController: creationController,
	}
}

func (h *CreationHandlerImpl) Pattern() string {
	return "/api/v1/lobbies"
}

func (h *CreationHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *CreationHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	createLobbyRequest, err := restutils.DecodeRequest[request.CreateLobby](writer, req)
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

	lobbyID, err := h.creationController.CreateLobby(userContext, createLobbyRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	response, err := json.Marshal(response.CreateLobby{LobbyID: lobbyID})
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, response, http.StatusCreated)
}
