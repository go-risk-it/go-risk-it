package reinforce

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

const (
	// minCardsForCombination is the minimum number of cards a player needs
	// before any valid combination can exist.
	minCardsForCombination = 3

	// Card value constants for combination identification (matching cards/performer.go).
	artilleryValue = 1
	infantryValue  = 10
	cavalryValue   = 100
	jollyValue     = 1000

	// maxJollyPerCombination limits how many jolly cards can appear in a
	// combination. A value at or above this factor times jollyValue is invalid.
	maxJollyPerCombination = 2
)

func (s *service) Walk(wctx moveservice.WalkContext) (sqlc.GamePhaseType, error) {
	nextPlayer, err := findNextAlivePlayer(
		wctx.PrevSnapshot.Players,
		wctx.PrevSnapshot.Game.Turn,
	)
	if err != nil {
		return "", fmt.Errorf("failed to find next player: %w", err)
	}

	priv, ok := wctx.PrivateSnapshots[nextPlayer.UserID]
	if !ok {
		return sqlc.GamePhaseTypeDEPLOY, nil
	}

	if hasValidCombination(priv.Cards) {
		return sqlc.GamePhaseTypeCARDS, nil
	}

	return sqlc.GamePhaseTypeDEPLOY, nil
}

// findNextAlivePlayer finds the next alive player after the current turn
// by walking the player list in index order, wrapping around, and skipping
// dead players.
func findNextAlivePlayer(
	players []snapshot.PlayerState,
	turn int64,
) (snapshot.PlayerState, error) {
	count := int64(len(players))
	if count == 0 {
		return snapshot.PlayerState{}, errors.New("no players")
	}

	// Build an index→player lookup for efficient access.
	byIndex := make(map[int64]snapshot.PlayerState, len(players))
	for _, p := range players {
		byIndex[p.Index] = p
	}

	nextTurn := turn + 1

	for range count {
		idx := nextTurn % count

		p, ok := byIndex[idx]
		if ok && p.Status == snapshot.PlayerAlive {
			return p, nil
		}

		nextTurn++
	}

	return snapshot.PlayerState{}, errors.New("no alive players found")
}

// hasValidCombination checks whether any subset of 3 cards forms a valid
// combination, using the same value-based identification as cards/performer.go.
func hasValidCombination(cards []snapshot.CardState) bool {
	if len(cards) < minCardsForCombination {
		return false
	}

	for i := range len(cards) - 2 {
		for j := i + 1; j < len(cards)-1; j++ {
			for k := j + 1; k < len(cards); k++ {
				if isValidCombination(cards[i].Type, cards[j].Type, cards[k].Type) {
					return true
				}
			}
		}
	}

	return false
}

// isValidCombination checks whether three card types form a valid combination.
func isValidCombination(a, b, c snapshot.CardType) bool {
	value := cardValue(a) + cardValue(b) + cardValue(c)

	if value >= maxJollyPerCombination*jollyValue {
		return false
	}

	validCombinations := map[int64]struct{}{
		3 * artilleryValue: {},
		3 * infantryValue:  {},
		3 * cavalryValue:   {},
		artilleryValue + infantryValue + cavalryValue: {},
		jollyValue + 2*artilleryValue:                 {},
		jollyValue + 2*infantryValue:                  {},
		jollyValue + 2*cavalryValue:                   {},
	}

	_, ok := validCombinations[value]

	return ok
}

// cardValue maps a snapshot CardType to its numeric value for combination
// identification, mirroring cards/performer.go's getCardValue.
func cardValue(ct snapshot.CardType) int64 {
	switch ct {
	case snapshot.CardArtillery:
		return artilleryValue
	case snapshot.CardInfantry:
		return infantryValue
	case snapshot.CardCavalry:
		return cavalryValue
	default:
		return jollyValue
	}
}
