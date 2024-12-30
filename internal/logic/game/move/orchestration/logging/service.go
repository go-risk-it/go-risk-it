package logging

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
)

type Service interface {
	GetMoveLogs(ctx ctx.GameContext, limit int64) ([]sqlc.GetMoveLogsRow, error)
	LogMoveQ(ctx ctx.GameContext, querier db.Querier, move, result any) error
}

type ServiceImpl struct {
	querier             db.Querier
	movePerformedSignal signals.MovePerformedSignal
}

var _ Service = (*ServiceImpl)(nil)

func New(
	querier db.Querier,
	movePerformedSignal signals.MovePerformedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:             querier,
		movePerformedSignal: movePerformedSignal,
	}
}

func (s *ServiceImpl) GetMoveLogs(
	ctx ctx.GameContext,
	limit int64,
) ([]sqlc.GetMoveLogsRow, error) {
	ctx.Log().Infow("getting move logs", "limit", limit)

	moveLogs, err := s.querier.GetMoveLogs(ctx, sqlc.GetMoveLogsParams{
		GameID:  ctx.GameID(),
		MaxLogs: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get move logs: %w", err)
	}

	return moveLogs, nil
}

func (s *ServiceImpl) LogMoveQ(ctx ctx.GameContext, querier db.Querier, move, result any) error {
	moveJSON, err := json.Marshal(move)
	if err != nil {
		return fmt.Errorf("failed to marshal move: %w", err)
	}

	var resultJSON []byte
	if !reflect.ValueOf(result).IsZero() {
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	moveLog, err := querier.CreateMoveLog(ctx, sqlc.CreateMoveLogParams{
		GameID:   ctx.GameID(),
		UserID:   ctx.UserID(),
		MoveData: moveJSON,
		Result:   resultJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to insert move log: %w", err)
	}

	go s.movePerformedSignal.Emit(ctx, signals.MovePerformedData{
		MoveLog: moveLog,
	})

	return nil
}
