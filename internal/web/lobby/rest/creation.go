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

func NewCreationHandler(
	creationController *controller.CreationController,
) *route.Route {
	h := &creationHandler{
		creationController: creationController,
	}

	return route.New("/api/v1/lobbies", true, middleware.HandleErrors(h.handle))
}

type creationHandler struct {
	creationController *controller.CreationController
}

func (h *creationHandler) handle(writer http.ResponseWriter, req *http.Request) error {
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
