package testonly

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

func NewResetHandler(
	testOnlyController Controller,
) *route.Route {
	return route.Authed("POST /api/v1/reset", func(w http.ResponseWriter, r *http.Request) error {
		if err := testOnlyController.ResetState(r.Context()); err != nil {
			return err
		}

		restutils.WriteResponse(w, []byte{}, http.StatusNoContent)

		return nil
	})
}
