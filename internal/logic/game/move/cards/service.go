package cards

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/rand"
	"github.com/jackc/pgx/v5/pgtype"
)

type CardCombination struct {
	CardIDs []int64
}

type Move struct {
	Combinations []CardCombination
}

type MoveResult struct {
	ExtraDeployableTroops int64
}

type Service interface {
	service.Service[Move, *MoveResult]
	Draw(ctx ctx.GameContext, querier db.Querier) error
}

type ServiceImpl struct {
	phaseService  phase.Service
	playerService player.Service
	regionService region.Service
	rng           rand.RNG
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	phaseService phase.Service,
	playerService player.Service,
	regionService region.Service,
	rng rand.RNG,
) *ServiceImpl {
	return &ServiceImpl{
		phaseService:  phaseService,
		playerService: playerService,
		regionService: regionService,
		rng:           rng,
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

	card := cards[s.rng.IntN(len(cards))]
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
