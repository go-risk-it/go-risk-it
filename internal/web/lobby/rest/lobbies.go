package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type LobbiesHandler interface {
	route.Route
}

type LobbiesHandlerImpl struct {
	managementController controller.ManagementController
}

var _ LobbiesHandler = (*LobbiesHandlerImpl)(nil)

func NewLobbiesHandler(
	managementController controller.ManagementController,
) *LobbiesHandlerImpl {
	return &LobbiesHandlerImpl{
		managementController: managementController,
	}
}

func (h *LobbiesHandlerImpl) Pattern() string {
	return "/api/v1/lobbies/summary"
}

func (h *LobbiesHandlerImpl) RequiresAuth() bool {
	return false
}

func (h *LobbiesHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	traceContext, ok := req.Context().(ctx.TraceContext)
	if !ok {
		http.Error(
			writer,
			fmt.Sprintf("invalid user context in route: %v", h.Pattern()),
			http.StatusInternalServerError,
		)

		return
	}

	lobbies, err := h.managementController.GetAvailableLobbies(traceContext)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	lobbiesResponse, err := json.Marshal(lobbies)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, lobbiesResponse, http.StatusOK)
}
