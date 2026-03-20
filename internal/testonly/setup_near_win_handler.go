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

type SetupNearWinHandler interface {
	route.Route
}

type SetupNearWinHandlerImpl struct {
	log                *zap.SugaredLogger
	testOnlyController Controller
}

var _ SetupNearWinHandler = (*SetupNearWinHandlerImpl)(nil)

func NewSetupNearWinHandler(
	log *zap.SugaredLogger,
	testOnlyController Controller,
) *SetupNearWinHandlerImpl {
	return &SetupNearWinHandlerImpl{
		log:                log,
		testOnlyController: testOnlyController,
	}
}

func (h *SetupNearWinHandlerImpl) Pattern() string {
	return "/api/v1/setup-near-win"
}

func (h *SetupNearWinHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *SetupNearWinHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	body, err := restutils.DecodeRequest[SetupNearWinRequest](writer, req)
	if err != nil {
		return
	}

	err = h.testOnlyController.SetupNearWin(ctx.WithLog(req.Context(), h.log), body.GameID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
