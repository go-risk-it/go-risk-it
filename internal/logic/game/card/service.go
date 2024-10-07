package card

import (
	"fmt"
	"math/rand"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateCards(ctx ctx.GameContext, querier db.Querier) error
}

type ServiceImpl struct {
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func New(regionService region.Service) *ServiceImpl {
	return &ServiceImpl{regionService: regionService}
}

func (s *ServiceImpl) CreateCards(ctx ctx.GameContext, querier db.Querier) error {
	ctx.Log().Infow("creating cards")

	cards, err := s.buildCards(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to build cards: %w", err)
	}

	_, err = querier.InsertCards(ctx, cards)
	if err != nil {
		return fmt.Errorf("unable to insert cards: %w", err)
	}

	ctx.Log().Infow("cards created")

	return nil
}

func (s *ServiceImpl) buildCards(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.InsertCardsParams, error) {
	regions, err := s.regionService.GetRegionsQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("unable to get regions: %w", err)
	}

	rand.Shuffle(len(regions), func(i, j int) {
		regions[i], regions[j] = regions[j], regions[i]
	})

	cards := make([]sqlc.InsertCardsParams, 0, len(regions))
	cardTypes := []sqlc.CardType{
		sqlc.CardTypeINFANTRY,
		sqlc.CardTypeARTILLERY,
		sqlc.CardTypeCAVALRY,
	}

	nCardsPerType := len(regions) / 3
	for i, region := range regions {
		cards = append(cards, sqlc.InsertCardsParams{
			RegionID: pgtype.Int8{Int64: region.ID, Valid: true},
			GameID:   ctx.GameID(),
			CardType: cardTypes[i/nCardsPerType],
		})
	}

	cards = append(cards, sqlc.InsertCardsParams{
		RegionID: pgtype.Int8{Int64: 0, Valid: false},
		GameID:   ctx.GameID(),
		CardType: sqlc.CardTypeJOLLY,
	})
	cards = append(cards, sqlc.InsertCardsParams{
		RegionID: pgtype.Int8{Int64: 0, Valid: false},
		GameID:   ctx.GameID(),
		CardType: sqlc.CardTypeJOLLY,
	})

	return cards, nil
}
