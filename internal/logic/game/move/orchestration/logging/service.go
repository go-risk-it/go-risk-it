package logging

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/tracing"
)

type Service interface {
	GetMoveLogs(ctx ctx.GameContext, limit int64) ([]sqlc.GetMoveLogsRow, error)
	LogMove(ctx ctx.GameContext, querier db.Querier, move, result any) (sqlc.GameMoveLog, error)
}

type service struct {
	querier db.Querier
}

var _ Service = (*service)(nil)

func New(
	querier db.Querier,
) Service {
	return &service{
		querier: querier,
	}
}

func (s *service) GetMoveLogs(
	ctx ctx.GameContext,
	limit int64,
) ([]sqlc.GetMoveLogsRow, error) {
	slog.DebugContext(ctx, "getting move logs", "limit", limit)

	moveLogs, err := s.querier.GetMoveLogs(ctx, sqlc.GetMoveLogsParams{
		GameID:  ctx.GameID(),
		MaxLogs: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get move logs: %w", err)
	}

	return moveLogs, nil
}

func (s *service) LogMove(
	ctx ctx.GameContext,
	querier db.Querier,
	move, result any,
) (sqlc.GameMoveLog, error) {
	ctx, span := tracing.StartGameSpan(ctx, "game.move.log")
	defer span.End()

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
