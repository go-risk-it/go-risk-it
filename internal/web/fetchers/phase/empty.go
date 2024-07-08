package phase

import (
	"encoding/json"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type EmptyPhaseFetcher interface {
	Fetcher
}

type EmptyPhaseFetcherImpl struct{}

var _ EmptyPhaseFetcher = (*EmptyPhaseFetcherImpl)(nil)

func NewEmptyPhaseFetcher() EmptyPhaseFetcher {
	return &EmptyPhaseFetcherImpl{}
}

func (f *EmptyPhaseFetcherImpl) FetchState(
	context ctx.GameContext,
	game *state.Game,
	stateChannel chan json.RawMessage,
) {
	FetchState(
		context,
		game,
		message.GameState,
		func(ctx ctx.GameContext, game *state.Game) (
			messaging.GameState[messaging.EmptyState], error,
		) {
			return messaging.GameState[messaging.EmptyState]{
				ID:   game.ID,
				Turn: game.Turn,
				Phase: messaging.Phase[messaging.EmptyState]{
					Type:  messaging.PhaseType(strings.ToLower(string(game.Phase))),
					State: messaging.EmptyState{},
				},
			}, nil
		},
		stateChannel,
	)
}
