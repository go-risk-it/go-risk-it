package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

type Move struct {
	Troops int64 `json:"troops"`
}

type Service interface {
	moveservice.Service[Move, struct{}]
	GetPhaseState(ctx ctx.GameContext) (sqlc.GetConquerPhaseStateRow, error)
	GetPhaseStateWithQuerier(
		ctx ctx.GameContext,
		querier db.Querier,
	) (sqlc.GetConquerPhaseStateRow, error)
}

type service struct {
	querier       db.Querier
	attackService attack.Service
}

func NewService(
	querier db.Querier,
	attackService attack.Service,
) (Service, moveservice.Service[Move, struct{}]) {
	svc := &service{
		querier:       querier,
		attackService: attackService,
	}

	return svc, moveservice.NewTracedService[Move, struct{}](svc)
}

var _ Service = (*service)(nil)

func (s *service) GetPhaseState(ctx ctx.GameContext) (sqlc.GetConquerPhaseStateRow, error) {
	return s.GetPhaseStateWithQuerier(ctx, s.querier)
}

func (s *service) GetPhaseStateWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
) (sqlc.GetConquerPhaseStateRow, error) {
	conquerPhase, err := querier.GetConquerPhaseState(ctx, ctx.GameID())
	if err != nil {
		return sqlc.GetConquerPhaseStateRow{}, fmt.Errorf(
			"failed to get conquer phase state: %w",
			err,
		)
	}

	return conquerPhase, nil
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeCONQUER
}
