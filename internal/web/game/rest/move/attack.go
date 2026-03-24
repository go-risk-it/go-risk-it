package move

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

func NewAttackHandler(moveController *controller.MoveController) *route.Route {
	return route.New(
		"/api/v1/games/{id}/moves/attacks",
		true,
		http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			handleMove[request.AttackMove](writer, req, moveController.PerformAttackMove)
		}),
	)
}
