package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type MoveLogFetcher interface {
	Fetcher
}
type MoveLogFetcherImpl struct {
	historyConfig     config.HistoryConfig
	moveLogController controller.MoveLogController
}

var _ MoveLogFetcher = (*MoveLogFetcherImpl)(nil)

func NewMoveLogFetcher(
	historyConfig config.HistoryConfig,
	moveLogController controller.MoveLogController,
) MoveLogFetcher {
	return &MoveLogFetcherImpl{
		historyConfig:     historyConfig,
		moveLogController: moveLogController,
	}
}

func (f *MoveLogFetcherImpl) FetchState(
	context ctx.GameContext,
	stateChannel chan json.RawMessage,
) {
	context.Log().Infow("history size:", "size", f.historyConfig.Size)

	FetchState(
		context,
		message.MoveHistory,
		func(context ctx.GameContext) (messaging.MoveHistory, error) {
			return f.moveLogController.GetMoveLogs(context, f.historyConfig.Size)
		},
		stateChannel,
	)
}
