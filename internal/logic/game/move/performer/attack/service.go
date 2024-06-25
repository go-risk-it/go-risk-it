package attack

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	service2 "github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
)

type Move struct {
	SourceRegionID  string
	TargetRegionID  string
	TroopsInSource  int64
	TroopsInTarget  int64
	AttackingTroops int64
}

type Service interface {
	service.Service[Move]
}

type ServiceImpl struct {
	querier                 db.Querier
	gameService             service2.Service
	playerService           service2.Service
	regionService           region.Service
	boardStateChangedSignal signals.BoardStateChangedSignal
	gameStateChangedSignal  signals.GameStateChangedSignal
}

var _ Service = &ServiceImpl{}

func NewService(
	que db.Querier,
	gameService service2.Service,
	playerService service2.Service,
	regionService region.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                 que,
		gameService:             gameService,
		playerService:           playerService,
		regionService:           regionService,
		boardStateChangedSignal: boardStateChangedSignal,
		gameStateChangedSignal:  gameStateChangedSignal,
	}
}

func (s *ServiceImpl) MustAdvanceQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
) bool {
	return false
}

func (s *ServiceImpl) PerformQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
	move Move,
) error {
	ctx.Log().Infow("performing attack move", "move", move)

	return nil
}
