package cards

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/jackc/pgx/v5/pgtype"
)

type Move struct{}

type MoveResult struct{}

type Service interface {
	service.Service[Move, *MoveResult]
	Draw(ctx ctx.GameContext, querier db.Querier) error
}

type ServiceImpl struct {
	phaseService  phase.Service
	playerService player.Service
}

func (s *ServiceImpl) PerformQ(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	panic("implement me")
}

func (s *ServiceImpl) Walk(ctx ctx.GameContext, querier db.Querier) (sqlc.PhaseType, error) {
	panic("implement me")
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	phaseService phase.Service,
	playerService player.Service,
) *ServiceImpl {
	return &ServiceImpl{
		phaseService:  phaseService,
		playerService: playerService,
	}
}

func (s *ServiceImpl) PhaseType() sqlc.PhaseType {
	return sqlc.PhaseTypeCARDS
}

func (s *ServiceImpl) ForcedAdvancementPhase() sqlc.PhaseType {
	return sqlc.PhaseTypeDEPLOY
}

func (s *ServiceImpl) Draw(ctx ctx.GameContext, querier db.Querier) error {
	ctx.Log().Infow("drawing card")

	cards, err := querier.GetAvailableCards(ctx, ctx.GameID())
	if err != nil {
		return fmt.Errorf("failed to get available cards: %w", err)
	}

	if len(cards) == 0 {
		return errors.New("no cards available")
	}

	playerID, err := s.extractPlayerID(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to extract player id: %w", err)
	}

	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	card := cards[0]
	if err := querier.DrawCard(ctx, sqlc.DrawCardParams{
		ID:      card.ID,
		OwnerID: pgtype.Int8{Int64: playerID, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to draw card: %w", err)
	}

	ctx.Log().Infow("card drawn", "card", card.ID)

	return nil
}

func (s *ServiceImpl) extractPlayerID(ctx ctx.GameContext, querier db.Querier) (int64, error) {
	players, err := s.playerService.GetPlayersQ(ctx, querier)
	if err != nil {
		return 0, fmt.Errorf("failed to get players: %w", err)
	}

	for _, player := range players {
		if player.UserID == ctx.UserID() {
			return player.ID, nil
		}
	}

	return 0, errors.New("player not found")
}
