package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/zap"
)

func NewResetHandler(
	log *zap.SugaredLogger,
	testOnlyController Controller,
) *route.Route {
	h := &resetHandler{
		log:                log,
		testOnlyController: testOnlyController,
	}

	return route.New("/api/v1/reset", true, h)
}

type resetHandler struct {
	log                *zap.SugaredLogger
	testOnlyController Controller
}

func (h *resetHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	err := h.testOnlyController.ResetState(ctx.WithLog(req.Context(), h.log))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
