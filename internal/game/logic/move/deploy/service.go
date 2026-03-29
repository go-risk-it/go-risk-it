package deploy

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
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
	querier       db.Querier
	phaseService  phase.Service
	regionService region.Service
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
	phaseService phase.Service,
	regionService region.Service,
) (Service, moveservice.Service[Move, struct{}]) {
	svc := &service{
		querier:       querier,
		phaseService:  phaseService,
		regionService: regionService,
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
	slog.DebugContext(ctx, "getting deployable troops")

	deployableTroops, err := querier.GetDeployableTroops(ctx, ctx.GameID())
	if err != nil {
		return 0, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	slog.DebugContext(ctx, "got deployable troops", "troops", deployableTroops)

	return deployableTroops, nil
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeDEPLOY
}
