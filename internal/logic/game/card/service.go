package card

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/rand"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateCards(ctx ctx.GameContext, querier db.Querier) error
	GetCardsForPlayer(ctx ctx.GameContext) ([]sqlc.GetCardsForPlayerRow, error)
	GetCardsForPlayerWithQuerier(
		ctx ctx.GameContext,
		querier db.Querier,
	) ([]sqlc.GetCardsForPlayerRow, error)
	TransferCardsOwnership(ctx ctx.GameContext, querier db.Querier, defendingPlayerID int64) error
}

type service struct {
	querier       db.Querier
	regionService region.Service
	rng           rand.RNG
}

func (s *service) TransferCardsOwnership(
	ctx ctx.GameContext,
	querier db.Querier,
	defendingPlayerID int64,
) error {
	slog.InfoContext(ctx, "transferring cards ownership")

	if err := querier.TransferCardsOwnership(ctx, sqlc.TransferCardsOwnershipParams{
		GameID: ctx.GameID(),
		From: pgtype.Int8{
			Int64: defendingPlayerID,
			Valid: true,
		},
		To: ctx.UserID(),
	}); err != nil {
		return fmt.Errorf("unable defender transfer cards ownership: %w", err)
	}

	slog.InfoContext(ctx, "transferred cards ownership")

	return nil
}

var _ Service = (*service)(nil)

func New(
	querier db.Querier,
	regionService region.Service,
	rng rand.RNG,
) Service {
	return &service{
		querier:       querier,
		regionService: regionService,
		rng:           rng,
	}
}

func (s *service) CreateCards(ctx ctx.GameContext, querier db.Querier) error {
	slog.InfoContext(ctx, "creating cards")

	cards, err := s.buildCards(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to build cards: %w", err)
	}

	_, err = querier.InsertCards(ctx, cards)
	if err != nil {
		return fmt.Errorf("unable to insert cards: %w", err)
	}

	slog.InfoContext(ctx, "cards created")

	return nil
}

func (s *service) buildCards(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.InsertCardsParams, error) {
	regions, err := s.regionService.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("unable to get regions: %w", err)
	}

	s.rng.Shuffle(len(regions), func(i, j int) {
		regions[i], regions[j] = regions[j], regions[i]
	})

	cards := make([]sqlc.InsertCardsParams, 0, len(regions))
	cardTypes := []sqlc.GameCardType{
		sqlc.GameCardTypeINFANTRY,
		sqlc.GameCardTypeARTILLERY,
		sqlc.GameCardTypeCAVALRY,
	}

	nCardsPerType := len(regions) / 3
	for i, region := range regions {
		cards = append(cards, sqlc.InsertCardsParams{
			RegionID: pgtype.Int8{Int64: region.ID, Valid: true},
			GameID:   ctx.GameID(),
			CardType: cardTypes[i/nCardsPerType],
		})
	}

	for range 2 {
		cards = append(cards, sqlc.InsertCardsParams{
			RegionID: pgtype.Int8{Int64: 0, Valid: false},
			GameID:   ctx.GameID(),
			CardType: sqlc.GameCardTypeJOLLY,
		})
	}

	return cards, nil
}

func (s *service) GetCardsForPlayer(ctx ctx.GameContext) ([]sqlc.GetCardsForPlayerRow, error) {
	return s.GetCardsForPlayerWithQuerier(ctx, s.querier)
}

func (s *service) GetCardsForPlayerWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetCardsForPlayerRow, error) {
	slog.InfoContext(ctx, "getting cards for player")

	result, err := querier.GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
		ID:     ctx.GameID(),
		UserID: ctx.UserID(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get cards for player: %w", err)
	}

	slog.InfoContext(ctx, "got cards for player", "cards", result)

	return result, nil
}
