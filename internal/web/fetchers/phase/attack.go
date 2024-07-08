package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type AttackPhaseFetcher interface {
	Fetcher
}

type AttackPhaseFetcherImpl struct {
	phaseController controller.PhaseController
}

var _ AttackPhaseFetcher = (*AttackPhaseFetcherImpl)(nil)

func NewAttackPhaseFetcher(phaseController controller.PhaseController) AttackPhaseFetcher {
	return &AttackPhaseFetcherImpl{
		phaseController: phaseController,
	}
}

func (f *AttackPhaseFetcherImpl) FetchState(
	context ctx.GameContext,
	game *state.Game,
	stateChannel chan json.RawMessage,
) {
	FetchState(
		context,
		game,
		message.GameState,
		f.phaseController.GetAttackPhaseState,
		stateChannel,
	)
}
