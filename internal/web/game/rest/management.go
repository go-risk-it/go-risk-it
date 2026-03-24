package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

func NewManagementHandler(
	gameController *controller.GameController,
) *route.Route {
	h := &managementHandler{
		gameController: gameController,
	}

	return route.New("/api/v1/games/summary", true, middleware.HandleErrors(h.handle))
}

type managementHandler struct {
	gameController *controller.GameController
}

func (h *managementHandler) handle(writer http.ResponseWriter, req *http.Request) error {
	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		return errors.New("invalid user context")
	}

	games, err := h.gameController.GetUserGames(userContext)
	if err != nil {
		return err
	}

	gamesResponse, err := json.Marshal(games)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	restutils.WriteResponse(writer, gamesResponse, http.StatusOK)

	return nil
}
