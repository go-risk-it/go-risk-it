package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

func NewResetHandler(
	testOnlyController Controller,
) *route.Route {
	h := &resetHandler{
		testOnlyController: testOnlyController,
	}

	return route.New("/api/v1/reset", true, h)
}

type resetHandler struct {
	testOnlyController Controller
}

func (h *resetHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	err := h.testOnlyController.ResetState(req.Context())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
