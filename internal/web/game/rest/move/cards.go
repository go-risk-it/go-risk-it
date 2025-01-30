package move

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type CardsHandler interface {
	route.Route
}

type CardsHandlerImpl struct {
	moveController controller.MoveController
}

var _ CardsHandler = (*CardsHandlerImpl)(nil)

func NewCardsHandler(moveController controller.MoveController) *CardsHandlerImpl {
	return &CardsHandlerImpl{
		moveController: moveController,
	}
}

func (h *CardsHandlerImpl) Pattern() string {
	return "/api/v1/games/{id}/moves/cards"
}

func (h *CardsHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *CardsHandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	serveMove[request.CardsMove](writer, req, h.moveController.PerformCardsMove)
}
