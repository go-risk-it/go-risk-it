package deploy

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}

type Service interface {
	moveservice.Service[Move]
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
) (Service, moveservice.Service[Move]) {
	svc := &service{
		querier:       querier,
		phaseService:  phaseService,
		regionService: regionService,
	}

	return svc, svc
}

func (s *service) GetDeployableTroops(ctx ctx.GameContext) (int64, error) {
	return s.GetDeployableTroopsWithQuerier(ctx, s.querier)
}

func (s *service) GetDeployableTroopsWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
) (int64, error) {
	slog.InfoContext(ctx, "getting deployable troops")

	deployableTroops, err := querier.GetDeployableTroops(ctx, ctx.GameID())
	if err != nil {
		return 0, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	slog.InfoContext(ctx, "got deployable troops", "troops", deployableTroops)

	return deployableTroops, nil
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeDEPLOY
}
