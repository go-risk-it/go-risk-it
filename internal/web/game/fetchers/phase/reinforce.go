package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type ReinforcePhaseFetcher interface {
	Fetcher
}

type ReinforcePhaseFetcherImpl struct {
	phaseController controller.PhaseController
}

var _ ReinforcePhaseFetcher = (*ReinforcePhaseFetcherImpl)(nil)

func NewReinforcePhaseFetcher(phaseController controller.PhaseController) ReinforcePhaseFetcher {
	return &ReinforcePhaseFetcherImpl{
		phaseController: phaseController,
	}
}

func (f *ReinforcePhaseFetcherImpl) FetchState(
	context ctx.GameContext,
	game *state.Game,
	stateChannel chan json.RawMessage,
) {
	fetcher.FetchState(
		context,
		message.GameState,
		getFetcherFunc(game, f.phaseController.GetReinforcePhaseState),
		stateChannel,
	)
}
