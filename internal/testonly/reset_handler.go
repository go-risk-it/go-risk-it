package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/zap"
)

type ResetHandler interface {
	route.Route
}

type ResetHandlerImpl struct {
	log                *zap.SugaredLogger
	testOnlyController Controller
}

var _ ResetHandler = (*ResetHandlerImpl)(nil)

func NewResetHandler(
	log *zap.SugaredLogger,
	testOnlyController Controller,
) *ResetHandlerImpl {
	return &ResetHandlerImpl{
		log:                log,
		testOnlyController: testOnlyController,
	}
}

func (h *ResetHandlerImpl) Pattern() string {
	return "/api/v1/reset"
}

func (h *ResetHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *ResetHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	err := h.testOnlyController.ResetState(ctx.WithLog(req.Context(), h.log))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
