package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws/message"
	"go.uber.org/fx"
)

type CardFetcher interface {
	Fetcher
}

type CardFetcherImpl struct {
	cardController controller.CardController
}

var _ CardFetcher = (*CardFetcherImpl)(nil)

type CardFetcherResult struct {
	fx.Out

	CardFetcher CardFetcher
	Fetcher     Fetcher `group:"private_fetchers"`
}

func NewCardFetcher(
	cardController controller.CardController,
) CardFetcherResult {
	res := &CardFetcherImpl{
		cardController: cardController,
	}

	return CardFetcherResult{
		CardFetcher: res,
		Fetcher:     res,
	}
}

func (c CardFetcherImpl) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	FetchState(ctx, message.CardState, c.cardController.GetCardState, stateChannel)
}
