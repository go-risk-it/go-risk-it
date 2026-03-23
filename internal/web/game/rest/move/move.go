package move

import (
	"errors"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		route.AsRoute(NewDeployHandler),
		route.AsRoute(NewAttackHandler),
		route.AsRoute(NewConquerHandler),
		route.AsRoute(NewReinforceHandler),
		route.AsRoute(NewCardsHandler),
	),
)

func serveMove[T any](
	writer http.ResponseWriter,
	req *http.Request,
	perform func(ctx ctx.GameContext, move T) error,
) error {
	gameContext, ok := req.Context().(ctx.GameContext)
	if !ok {
		return errors.New("invalid move context")
	}

	moveRequest, err := restutils.DecodeRequest[T](writer, req)
	if err != nil {
		return err
	}

	if err := perform(gameContext, moveRequest); err != nil {
		return err
	}

	restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)

	return nil
}

func handleMove[T any](
	writer http.ResponseWriter,
	req *http.Request,
	perform func(ctx ctx.GameContext, move T) error,
) {
	middleware.HandleErrors(func(w http.ResponseWriter, r *http.Request) error {
		return serveMove[T](w, r, perform)
	}).ServeHTTP(writer, req)
}
