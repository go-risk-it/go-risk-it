package fetcher

import (
	"encoding/json"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

type MoveLogFetcher interface {
	FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage)
}

type moveLogFetcher struct {
	historyConfig     config.HistoryConfig
	moveLogController *controller.MoveLogController
}

var _ MoveLogFetcher = (*moveLogFetcher)(nil)

func NewMoveLogFetcher(
	historyConfig config.HistoryConfig,
	moveLogController *controller.MoveLogController,
) MoveLogFetcher {
	return &moveLogFetcher{
		historyConfig:     historyConfig,
		moveLogController: moveLogController,
	}
}

func (f *moveLogFetcher) FetchState(
	context ctx.GameContext,
	stateChannel chan json.RawMessage,
) {
	slog.InfoContext(context, "history size", "size", f.historyConfig.Size)

	sharedfetcher.FetchState(
		context,
		message.MoveHistory,
		func(context ctx.GameContext) (messaging.MoveHistory, error) {
			return f.moveLogController.GetMoveLogs(context, f.historyConfig.Size)
		},
		stateChannel,
	)
}
