package cards

import (
	cardsapi "github.com/go-risk-it/go-risk-it/internal/game/api/moves/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/rand"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

const DefaultTroopGrant = 2

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
	Draw(deck []snapshot.CardState) (snapshot.CardState, error)
}

type service struct {
	rng rand.RNG
}

var _ Service = (*service)(nil)

func NewService(
	rng rand.RNG,
) (Service, moveservice.Service[Move, *MoveResult]) {
	svc := &service{
		rng: rng,
	}

	return svc, moveservice.NewTracedService[Move, *MoveResult](svc)
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeCARDS
}

// Draw selects a random card from the given deck using the RNG. It is a pure
// computation — no DB queries or writes. The caller is responsible for
// persisting the draw (DrawCard) and recording the DeckDelta.
func (s *service) Draw(deck []snapshot.CardState) (snapshot.CardState, error) {
	if len(deck) == 0 {
		return snapshot.CardState{}, domainerrors.NewValidationError(
			"no cards available in deck",
		)
	}

	return deck[s.rng.IntN(len(deck))], nil
}
