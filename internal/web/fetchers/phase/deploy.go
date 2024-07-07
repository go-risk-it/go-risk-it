package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type DeployPhaseFetcher interface {
	Fetcher
}

type DeployPhaseFetcherImpl struct {
	phaseController controller.PhaseController
}

var _ DeployPhaseFetcher = (*DeployPhaseFetcherImpl)(nil)

func NewDeployPhaseFetcher(
	phaseController controller.PhaseController,
) DeployPhaseFetcher {
	return &DeployPhaseFetcherImpl{
		phaseController: phaseController,
	}
}

func (f *DeployPhaseFetcherImpl) FetchState(
	context ctx.GameContext,
	game *state.Game,
	stateChannel chan json.RawMessage,
) {
	FetchState(
		context,
		game,
		message.GameState,
		f.phaseController.GetDeployPhaseState,
		stateChannel,
	)
}
