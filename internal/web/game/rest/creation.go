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

type Handler interface {
	route.Route
}

type HandlerImpl struct {
	gameController controller.GameController
}

var _ Handler = (*HandlerImpl)(nil)

func NewCreationHandler(gameController controller.GameController) *HandlerImpl {
	return &HandlerImpl{
		gameController: gameController,
	}
}

func (h *HandlerImpl) Pattern() string {
	return "/api/v1/games"
}

func (h *HandlerImpl) RequiresAuth() bool {
	return true
}

func (h *HandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	middleware.HandleErrors(h.handle).ServeHTTP(writer, req)
}

func (h *HandlerImpl) handle(writer http.ResponseWriter, req *http.Request) error {
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
