package signals

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MovePerformedHandlerParams struct {
	fx.In

	Log               *zap.SugaredLogger
	Signal            signals.MovePerformedSignal
	MoveLogController controller.MoveLogController
	MoveLogFetcher    fetcher.MoveLogFetcher
	ConnectionManager ws.Manager
}

func HandleMovePerformed(
	params MovePerformedHandlerParams,
) {
	params.Signal.AddListener(func(context context.Context, data signals.MovePerformedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			params.Log.Errorw("context is not game context", "context", context)

			return
		}

		gameContext.Log().Infow("handling move performed. fetching move logs and publishing")

		fetchStateAndPublish(
			gameContext,
			func(gameCtx ctx.GameContext, stateChannel chan json.RawMessage) {
				fetcher.FetchState(
					gameCtx,
					message.MoveHistory,
					func(gameCtx2 ctx.GameContext) (messaging.MoveHistory, error) {
						history, err := params.MoveLogController.ConvertMoveLogs(
							gameCtx2,
							[]sqlc.GameMoveLog{data.MoveLog},
						)
						if err != nil {
							return messaging.MoveHistory{}, fmt.Errorf(
								"failed to convert move logs: %w",
								err,
							)
						}

						return history, nil
					},
					stateChannel,
				)
			},
			params.ConnectionManager.Broadcast,
		)
	})
}
