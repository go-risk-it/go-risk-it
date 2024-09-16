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
		route.AsRoute(NewConquerHandler),
	),
)

func serveMove[T any](
	writer http.ResponseWriter,
	req *http.Request,
	perform func(ctx ctx.GameContext, move T) error,
) {
	moveRequest, err := restutils.DecodeRequest[T](writer, req)
	if err != nil {
		return
	}

	gameContext, ok := req.Context().(ctx.GameContext)
	if !ok {
		http.Error(writer, "invalid move context", http.StatusInternalServerError)

		return
	}

	if err := perform(gameContext, moveRequest); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}
