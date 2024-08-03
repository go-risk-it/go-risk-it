package cards

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

type Move struct{}

type Service interface {
	service.Service[Move]
}

type ServiceImpl struct {
	phaseService phase.Service
}

func (s *ServiceImpl) PerformQ(ctx ctx.MoveContext, querier db.Querier, move Move) error {
	panic("implement me")
}

func (s *ServiceImpl) Walk(ctx ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error) {
	panic("implement me")
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	phaseService phase.Service,
) *ServiceImpl {
	return &ServiceImpl{
		phaseService: phaseService,
	}
}
