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
	return route.Authed(
		"POST /api/v1/setup-near-win",
		func(writer http.ResponseWriter, request *http.Request) error {
			body, err := restutils.DecodeRequest[SetupNearWinRequest](writer, request)
			if err != nil {
				return err
			}

			if err := testOnlyController.SetupNearWin(request.Context(), body.GameID); err != nil {
				return err
			}

			restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)

			return nil
		},
	)
}
