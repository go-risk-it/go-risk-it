package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/zap"
)

type SetupNearWinRequest struct {
	GameID int64 `json:"gameId"`
}

func NewSetupNearWinHandler(
	log *zap.SugaredLogger,
	testOnlyController Controller,
) *route.Route {
	h := &setupNearWinHandler{
		log:                log,
		testOnlyController: testOnlyController,
	}

	return route.New("/api/v1/setup-near-win", true, h)
}

type setupNearWinHandler struct {
	log                *zap.SugaredLogger
	testOnlyController Controller
}

func (h *setupNearWinHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	body, err := restutils.DecodeRequest[SetupNearWinRequest](writer, req)
	if err != nil {
		_ = restutils.WriteError(writer, err)

		return
	}

	err = h.testOnlyController.SetupNearWin(ctx.WithLog(req.Context(), h.log), body.GameID)
	if err != nil {
		_ = restutils.WriteError(writer, err)

		return
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
