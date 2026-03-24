package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type LobbyStateFetcher interface {
	Fetcher
}

type LobbyStateFetcherImpl struct {
	stateController *controller.StateController
}

var _ LobbyStateFetcher = (*LobbyStateFetcherImpl)(nil)

func NewLobbyStateFetcher(stateController *controller.StateController) *LobbyStateFetcherImpl {
	return &LobbyStateFetcherImpl{
		stateController: stateController,
	}
}

func (f *LobbyStateFetcherImpl) FetchState(
	context ctx.LobbyContext,
	stateChannel chan json.RawMessage,
) {
	sharedfetcher.FetchState(
		context,
		message.LobbyState,
		f.stateController.GetLobbyState,
		stateChannel,
	)
}
