package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type cardFetcher struct {
	cardController *controller.CardController
}

type CardFetcherResult struct {
	fx.Out

	Fetcher Fetcher `group:"private_fetchers"`
}

func NewCardFetcher(
	cardController *controller.CardController,
) CardFetcherResult {
	return CardFetcherResult{
		Fetcher: &cardFetcher{
			cardController: cardController,
		},
	}
}

func (c *cardFetcher) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	sharedfetcher.FetchState(ctx, message.CardState, c.cardController.GetCardState, stateChannel)
}
