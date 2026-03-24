package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/lobby/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type CreationHandlerImpl struct {
	creationController *controller.CreationController
}

var _ route.Route = (*CreationHandlerImpl)(nil)

func NewCreationHandler(
	creationController *controller.CreationController,
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
	middleware.HandleErrors(h.handle).ServeHTTP(writer, req)
}

func (h *CreationHandlerImpl) handle(writer http.ResponseWriter, req *http.Request) error {
	createLobbyRequest, err := restutils.DecodeRequest[request.CreateLobby](writer, req)
	if err != nil {
		return err
	}

	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		return errors.New("invalid user context")
	}

	lobbyID, err := h.creationController.CreateLobby(userContext, createLobbyRequest)
	if err != nil {
		return err
	}

	resp, err := json.Marshal(response.CreateLobby{LobbyID: lobbyID})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	restutils.WriteResponse(writer, resp, http.StatusCreated)

	return nil
}
