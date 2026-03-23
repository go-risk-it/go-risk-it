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

type AdvancementHandler interface {
	route.Route
}

type AdvancementHandlerImpl struct {
	advancementController controller.AdvancementController
}

var _ AdvancementHandler = (*AdvancementHandlerImpl)(nil)

func NewAdvancementHandler(
	advancementController controller.AdvancementController,
) *AdvancementHandlerImpl {
	return &AdvancementHandlerImpl{
		advancementController: advancementController,
	}
}

func (h *AdvancementHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/advancements"
}

func (h *AdvancementHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *AdvancementHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	middleware.HandleErrors(h.handle).ServeHTTP(writer, req)
}

func (h *AdvancementHandlerImpl) handle(
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
