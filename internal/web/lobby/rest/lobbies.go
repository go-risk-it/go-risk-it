package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

func NewLobbiesHandler(
	managementController *controller.ManagementController,
) *route.Route {
	h := &lobbiesHandler{
		managementController: managementController,
	}

	return route.New("/api/v1/lobbies/summary", true, middleware.HandleErrors(h.handle))
}

type lobbiesHandler struct {
	managementController *controller.ManagementController
}

func (h *lobbiesHandler) handle(writer http.ResponseWriter, req *http.Request) error {
	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		return errors.New("invalid user context")
	}

	lobbies, err := h.managementController.GetUserLobbies(userContext)
	if err != nil {
		return err
	}

	lobbiesResponse, err := json.Marshal(lobbies)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	restutils.WriteResponse(writer, lobbiesResponse, http.StatusOK)

	return nil
}
