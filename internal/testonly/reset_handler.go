package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"go.uber.org/zap"
)

type ResetHandler interface {
	Pattern() string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type ResetHandlerImpl struct {
	log                *zap.SugaredLogger
	testOnlyController Controller
}

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
	return "/api/1/reset"
}

func (h *ResetHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	err := h.testOnlyController.ResetState(req.Context())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	rest.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
