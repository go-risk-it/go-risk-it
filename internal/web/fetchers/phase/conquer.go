package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
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
	FetchState(
		context,
		game,
		message.GameState,
		f.phaseController.GetConquerPhaseState,
		stateChannel,
	)
}
