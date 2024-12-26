package logging

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Service interface {
	LogMoveQ(ctx ctx.GameContext, querier db.Querier, move, result any) error
}

type ServiceImpl struct{}

var _ Service = (*ServiceImpl)(nil)

func New() *ServiceImpl {
	return &ServiceImpl{}
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

	_, err = querier.CreateMoveLog(ctx, sqlc.CreateMoveLogParams{
		GameID:   ctx.GameID(),
		UserID:   ctx.UserID(),
		MoveData: moveJSON,
		Result:   resultJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to insert move log: %w", err)
	}

	return nil
}
