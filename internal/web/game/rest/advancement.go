package rest

import (
	"errors"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

func NewAdvancementHandler(
	advancementController *controller.AdvancementController,
) *route.Route {
	h := &advancementHandler{
		advancementController: advancementController,
	}

	return route.New(
		"/api/v1/games/{id}/advancements", true, middleware.HandleErrors(h.handle),
	)
}

type advancementHandler struct {
	advancementController *controller.AdvancementController
}

func (h *advancementHandler) handle(
	writer http.ResponseWriter,
	req *http.Request,
) error {
	gameContext, ok := req.Context().(ctx.GameContext)
	if !ok {
		return errors.New("invalid move context")
	}

	advancementRequest, err := restutils.DecodeRequest[request.Advancement](writer, req)
	if err != nil {
		return err
	}

	if err = h.advancementController.Advance(gameContext, advancementRequest); err != nil {
		return err
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)

	return nil
}
