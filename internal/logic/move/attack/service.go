package attack

import (
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/performer"
	"github.com/go-risk-it/go-risk-it/internal/logic/orchestration/validation"
	"github.com/go-risk-it/go-risk-it/internal/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"github.com/go-risk-it/go-risk-it/internal/signals"
	"go.uber.org/zap"
)

type Move struct {
	SourceRegionID  string
	TargetRegionID  string
	TroopsInSource  int64
	TroopsInTarget  int64
	AttackingTroops int64
}

type Service interface {
	performer.Service[Move]
}

type ServiceImpl struct {
	log                     *zap.SugaredLogger
	querier                 db.Querier
	gameService             game.Service
	playerService           game.Service
	regionService           region.Service
	validationService       validation.Service
	boardStateChangedSignal signals.BoardStateChangedSignal
	gameStateChangedSignal  signals.GameStateChangedSignal
}

var _ Service = &ServiceImpl{}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService game.Service,
	playerService game.Service,
	regionService region.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                 que,
		log:                     log,
		gameService:             gameService,
		playerService:           playerService,
		regionService:           regionService,
		validationService:       validationService,
		boardStateChangedSignal: boardStateChangedSignal,
		gameStateChangedSignal:  gameStateChangedSignal,
	}
}

func (s *ServiceImpl) MustAdvanceQ(
	ctx riskcontext.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
) bool {
	return false
}

func (s *ServiceImpl) PerformQ(
	ctx riskcontext.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
	move Move,
) error {
	s.log.Infow(
		"performing attack move",
		"move", move,
	)

	return nil
}
