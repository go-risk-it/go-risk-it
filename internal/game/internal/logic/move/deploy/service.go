package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

type Move struct {
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}

type Service interface {
	moveservice.Service[Move, struct{}]
	GetDeployableTroops(ctx ctx.GameContext) (int64, error)
	GetDeployableTroopsWithQuerier(ctx ctx.GameContext, querier db.Querier) (int64, error)
}

type service struct {
	querier db.Querier
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
) (Service, moveservice.Service[Move, struct{}]) {
	svc := &service{
		querier: querier,
	}

	return svc, moveservice.NewTracedService[Move, struct{}](svc)
}

func (s *service) GetDeployableTroops(ctx ctx.GameContext) (int64, error) {
	return s.GetDeployableTroopsWithQuerier(ctx, s.querier)
}

func (s *service) GetDeployableTroopsWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
) (int64, error) {
	deployableTroops, err := querier.GetDeployableTroops(ctx, ctx.GameID())
	if err != nil {
		return 0, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	return deployableTroops, nil
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeDEPLOY
}
