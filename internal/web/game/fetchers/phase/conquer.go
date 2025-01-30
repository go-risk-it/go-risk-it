package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws/message"
)

type ConquerPhaseFetcher interface {
	Fetcher
}

type ConquerPhaseFetcherImpl struct {
	phaseController controller.PhaseController
}

var _ ConquerPhaseFetcher = (*ConquerPhaseFetcherImpl)(nil)

func NewConquerPhaseFetcher(phaseController controller.PhaseController) ConquerPhaseFetcher {
	return &ConquerPhaseFetcherImpl{
		phaseController: phaseController,
	}
}

func (f *ConquerPhaseFetcherImpl) FetchState(
	context ctx.GameContext,
	game *state.Game,
	stateChannel chan json.RawMessage,
) {
	fetcher.FetchState(
		context,
		message.GameState,
		getFetcherFunc(game, f.phaseController.GetConquerPhaseState),
		stateChannel,
	)
}
