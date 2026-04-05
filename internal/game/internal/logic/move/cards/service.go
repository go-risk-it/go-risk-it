package cards

import (
	"errors"
	"fmt"

	cardsapi "github.com/go-risk-it/go-risk-it/internal/game/api/moves/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/rand"
)

const DefaultTroopGrant = 2

// minCardsForCombination is the minimum number of cards a player needs to attempt a combination.
const minCardsForCombination = 3

type CardCombination struct {
	CardIDs []int64 `json:"cardIds"`
}

type Move struct {
	Combinations []CardCombination `json:"combinations"`
}

// MoveResult is a type alias to the canonical definition in game/api/moves/cards.
// This preserves backward compatibility for all existing consumers.
type MoveResult = cardsapi.MoveResult

type Service interface {
	moveservice.Service[Move, *MoveResult]
	Draw(ctx ctx.GameContext, querier db.Querier) (snapshot.CardState, error)
	NextPlayerHasValidCombination(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type service struct {
	boardService  board.Service
	phaseService  phase.Service
	playerService player.Service
	regionService region.Service
	rng           rand.RNG
}

var _ Service = (*service)(nil)

func NewService(
	boardService board.Service,
	phaseService phase.Service,
	playerService player.Service,
	regionService region.Service,
	rng rand.RNG,
) (Service, moveservice.Service[Move, *MoveResult]) {
	svc := &service{
		boardService:  boardService,
		phaseService:  phaseService,
		playerService: playerService,
		regionService: regionService,
		rng:           rng,
	}

	return svc, moveservice.NewTracedService[Move, *MoveResult](svc)
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeCARDS
}

func (s *service) Draw(ctx ctx.GameContext, querier db.Querier) (snapshot.CardState, error) {
	cards, err := querier.GetAvailableCards(ctx, ctx.GameID())
	if err != nil {
		return snapshot.CardState{}, fmt.Errorf("failed to get available cards: %w", err)
	}

	if len(cards) == 0 {
		return snapshot.CardState{}, errors.New("no cards available")
	}

	card := cards[s.rng.IntN(len(cards))]
	if err := querier.DrawCard(ctx, sqlc.DrawCardParams{
		ID:     card.ID,
		UserID: ctx.UserID(),
		GameID: ctx.GameID(),
	}); err != nil {
		return snapshot.CardState{}, fmt.Errorf("failed to draw card: %w", err)
	}

	// Resolve the drawn card's region external reference by querying the
	// player's cards (which join through the region table).
	playerCards, err := querier.GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
		ID:     ctx.GameID(),
		UserID: ctx.UserID(),
	})
	if err != nil {
		return snapshot.CardState{}, fmt.Errorf("failed to get cards for player: %w", err)
	}

	for _, pc := range playerCards {
		if pc.ID == card.ID {
			cardType, err := mapCardType(pc.CardType)
			if err != nil {
				return snapshot.CardState{}, fmt.Errorf("failed to map card type: %w", err)
			}

			regionName := ""
			if pc.Region.Valid {
				regionName = pc.Region.String
			}

			return snapshot.CardState{
				ID:     pc.ID,
				Type:   cardType,
				Region: regionName,
			}, nil
		}
	}

	return snapshot.CardState{}, fmt.Errorf("drawn card %d not found in player's hand", card.ID)
}

func (s *service) NextPlayerHasValidCombination(
	ctx ctx.GameContext,
	querier db.Querier,
) (bool, error) {
	nextPlayer, err := s.playerService.GetNextPlayer(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player: %w", err)
	}

	nextPlayerCards, err := querier.GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
		ID:     ctx.GameID(),
		UserID: nextPlayer.UserID,
	})
	if err != nil {
		return false, fmt.Errorf("unable to get cards for player: %w", err)
	}

	if len(nextPlayerCards) < minCardsForCombination {
		return false, nil
	}

	cardIndex := make(map[int64]sqlc.GetCardsForPlayerRow)
	for _, card := range nextPlayerCards {
		cardIndex[card.ID] = card
	}

	// Try each possible combination of 3 cards
	for i := range len(nextPlayerCards) - 2 {
		for j := i + 1; j < len(nextPlayerCards)-1; j++ {
			for k := j + 1; k < len(nextPlayerCards); k++ {
				combination := CardCombination{
					CardIDs: []int64{
						nextPlayerCards[i].ID,
						nextPlayerCards[j].ID,
						nextPlayerCards[k].ID,
					},
				}

				if _, err := identifyCombination(combination, cardIndex); err == nil {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// mapCardType converts a sqlc.GameCardType to the snapshot.CardType
// used in the public API.
func mapCardType(t sqlc.GameCardType) (snapshot.CardType, error) {
	switch t {
	case sqlc.GameCardTypeINFANTRY:
		return snapshot.CardInfantry, nil
	case sqlc.GameCardTypeCAVALRY:
		return snapshot.CardCavalry, nil
	case sqlc.GameCardTypeARTILLERY:
		return snapshot.CardArtillery, nil
	case sqlc.GameCardTypeJOLLY:
		return snapshot.CardJolly, nil
	default:
		return "", fmt.Errorf("unknown card type: %s", t)
	}
}
