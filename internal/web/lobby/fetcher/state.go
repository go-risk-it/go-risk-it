package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type LobbyStateFetcher interface {
	Fetcher
}

type LobbyStateFetcherImpl struct {
	stateController controller.StateController
}

var _ LobbyStateFetcher = (*LobbyStateFetcherImpl)(nil)

func NewLobbyStateFetcher(stateController controller.StateController) LobbyStateFetcher {
	return &LobbyStateFetcherImpl{
		stateController: stateController,
	}
}

func (f *LobbyStateFetcherImpl) FetchState(
	context ctx.LobbyContext,
	stateChannel chan json.RawMessage,
) {
	FetchState(
		context,
		message.LobbyState,
		f.stateController.GetLobbyState,
		stateChannel,
	)
}
