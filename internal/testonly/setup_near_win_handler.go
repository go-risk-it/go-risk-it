package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

type SetupNearWinRequest struct {
	GameID int64 `json:"gameId"`
}

func NewSetupNearWinHandler(
	testOnlyController Controller,
) *route.Route {
	h := &setupNearWinHandler{
		testOnlyController: testOnlyController,
	}

	return route.New("/api/v1/setup-near-win", true, h)
}

type setupNearWinHandler struct {
	testOnlyController Controller
}

func (h *setupNearWinHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	body, err := restutils.DecodeRequest[SetupNearWinRequest](writer, req)
	if err != nil {
		_ = restutils.WriteError(writer, err)

		return
	}

	err = h.testOnlyController.SetupNearWin(req.Context(), body.GameID)
	if err != nil {
		_ = restutils.WriteError(writer, err)

		return
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
