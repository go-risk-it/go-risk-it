package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

func NewCreationHandler(gameController *controller.GameController) *route.Route {
	h := &creationHandler{
		gameController: gameController,
	}

	return route.New("/api/v1/games", true, middleware.HandleErrors(h.handle))
}

type creationHandler struct {
	gameController *controller.GameController
}

func (h *creationHandler) handle(writer http.ResponseWriter, req *http.Request) error {
	createGameRequest, err := restutils.DecodeRequest[request.CreateGame](writer, req)
	if err != nil {
		return err
	}

	userContext, ok := req.Context().(ctx.UserContext)
	if !ok {
		return errors.New("invalid user context")
	}

	gameID, err := h.gameController.CreateGame(userContext, createGameRequest)
	if err != nil {
		return err
	}

	createGameResponse, err := json.Marshal(response.CreateGame{GameID: gameID})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	restutils.WriteResponse(writer, createGameResponse, http.StatusCreated)

	return nil
}
