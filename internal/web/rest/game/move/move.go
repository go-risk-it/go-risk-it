package move

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		route.AsRoute(NewDeployHandler),
		route.AsRoute(NewAttackHandler),
	),
)

func serveMove[T any](
	writer http.ResponseWriter,
	req *http.Request,
	perform func(ctx ctx.MoveContext, move T) error,
) {
	moveRequest, err := restutils.DecodeRequest[T](writer, req)
	if err != nil {
		return
	}

	moveContext, ok := req.Context().(ctx.MoveContext)
	if !ok {
		http.Error(writer, "invalid move context", http.StatusInternalServerError)

		return
	}

	if err := perform(moveContext, moveRequest); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
