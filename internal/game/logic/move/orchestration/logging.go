package orchestration

import (
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
)

type LoggingService interface {
	GetMoveLogs(ctx ctx.GameContext, limit int64) ([]sqlc.GetMoveLogsRow, error)
	LogMove(ctx ctx.GameContext, querier db.Querier, move, result any) (sqlc.GameMoveLog, error)
}

type loggingServiceImpl struct {
	querier db.Querier
}

var _ LoggingService = (*loggingServiceImpl)(nil)

func NewLoggingService(
	querier db.Querier,
) LoggingService {
	return &loggingServiceImpl{
		querier: querier,
	}
}

func (s *loggingServiceImpl) GetMoveLogs(
	ctx ctx.GameContext,
	limit int64,
) ([]sqlc.GetMoveLogsRow, error) {
	moveLogs, err := s.querier.GetMoveLogs(ctx, sqlc.GetMoveLogsParams{
		GameID:  ctx.GameID(),
		MaxLogs: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get move logs: %w", err)
	}

	return moveLogs, nil
}

func (s *loggingServiceImpl) LogMove(
	ctx ctx.GameContext,
	querier db.Querier,
	move, result any,
) (sqlc.GameMoveLog, error) {
	ctx, done := tracing.StartGameSpan(ctx, "game.move.log")
	defer done(nil)

	moveJSON, err := json.Marshal(move)
	if err != nil {
		return sqlc.GameMoveLog{}, fmt.Errorf("failed to marshal move: %w", err)
	}

	var resultJSON []byte
	if result != nil {
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return sqlc.GameMoveLog{}, fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	moveLog, err := querier.CreateMoveLog(ctx, sqlc.CreateMoveLogParams{
		GameID:   ctx.GameID(),
		UserID:   ctx.UserID(),
		MoveData: moveJSON,
		Result:   resultJSON,
	})
	if err != nil {
		return sqlc.GameMoveLog{}, fmt.Errorf("failed to insert move log: %w", err)
	}

	return moveLog, nil
}
