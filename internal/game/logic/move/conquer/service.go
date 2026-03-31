package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/card"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
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
	querier        db.Querier
	attackService  attack.Service
	cardService    card.Service
	missionService mission.Service
	phaseService   phase.Service
	regionService  region.Service
}

func NewService(
	querier db.Querier,
	attackService attack.Service,
	cardService card.Service,
	missionService mission.Service,
	phaseService phase.Service,
	regionService region.Service,
) (Service, moveservice.Service[Move, struct{}]) {
	svc := &service{
		querier:        querier,
		attackService:  attackService,
		cardService:    cardService,
		missionService: missionService,
		phaseService:   phaseService,
		regionService:  regionService,
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
