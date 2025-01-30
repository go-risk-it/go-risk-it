package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type Handler interface {
	route.Route
}

type HandlerImpl struct {
	creationController controller.CreationController
}

var _ Handler = (*HandlerImpl)(nil)

func NewCreationHandler(creationController controller.CreationController) *HandlerImpl {
	return &HandlerImpl{
		creationController: creationController,
	}
}

func (h *HandlerImpl) Pattern() string {
	return "/api/v1/lobbies"
}

func (h *HandlerImpl) RequiresAuth() bool {
	return true
}

func (h *HandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		http.Error(
			writer,
			fmt.Sprintf("invalid user context in route: %v", h.Pattern()),
			http.StatusInternalServerError,
		)

		return
	}

	lobbyID, err := h.creationController.CreateLobby(userContext)
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
