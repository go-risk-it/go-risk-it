package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type CardsPhaseFetcher interface {
	Fetcher
}

type CardsPhaseFetcherImpl struct {
	phaseController controller.PhaseController
}

var _ CardsPhaseFetcher = (*CardsPhaseFetcherImpl)(nil)

func NewCardsPhaseFetcher(phaseController controller.PhaseController) CardsPhaseFetcher {
	return &CardsPhaseFetcherImpl{
		phaseController: phaseController,
	}
}

func (f *CardsPhaseFetcherImpl) FetchState(
	context ctx.GameContext,
	game *state.Game,
	stateChannel chan json.RawMessage,
) {
	fetcher.FetchState(
		context,
		message.GameState,
		getFetcherFunc(game, f.phaseController.GetCardsPhaseState),
		stateChannel,
	)
}
