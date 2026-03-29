//go:build invariant

package invariant

import (
	"math/rand/v2"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/deploy"
)

// Generator produces valid moves for each game phase.
type Generator struct {
	harness *Harness
	rng     *rand.Rand
}

// NewGenerator creates a Generator with the given random seed.
func NewGenerator(harness *Harness, seed uint64) *Generator {
	return &Generator{
		harness: harness,
		rng:     rand.New(rand.NewPCG(seed, seed)),
	}
}

// CardsMove finds and returns a valid card trade.
// CARDS phase is only entered when the player has a valid combo.
func (gen *Generator) CardsMove(
	tb testing.TB,
	gCtx ctx.GameContext,
	_ *GameSnapshot,
) cards.Move {
	tb.Helper()

	playerCards, err := gen.harness.Querier.GetCardsForPlayer(
		gCtx,
		sqlc.GetCardsForPlayerParams{
			ID:     gCtx.GameID(),
			UserID: gCtx.UserID(),
		},
	)
	if err != nil {
		tb.Fatalf("failed to get cards: %v", err)
	}

	combo := findValidCombination(tb, playerCards)

	return cards.Move{
		Combinations: []cards.CardCombination{
			{CardIDs: combo},
		},
	}
}

// DeployMove deploys all available troops to a random owned region.
func (gen *Generator) DeployMove(
	tb testing.TB,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
) deploy.Move {
	tb.Helper()

	deployable, err := gen.harness.DeployService.GetDeployableTroops(
		gCtx,
	)
	if err != nil {
		tb.Fatalf("failed to get deployable troops: %v", err)
	}

	owned := snap.RegionsOwnedBy(gCtx.UserID())
	if len(owned) == 0 {
		tb.Fatalf("player %s owns no regions", gCtx.UserID())
	}

	target := owned[gen.rng.IntN(len(owned))]

	return deploy.Move{
		RegionID:      target.ExternalReference,
		CurrentTroops: target.Troops,
		DesiredTroops: target.Troops + deployable,
	}
}

// AttackMove finds a valid attack or returns false if none exists.
func (gen *Generator) AttackMove(
	tb testing.TB,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
) (attack.Move, bool) {
	tb.Helper()

	owned := snap.RegionsOwnedBy(gCtx.UserID())

	gen.rng.Shuffle(len(owned), func(i, j int) {
		owned[i], owned[j] = owned[j], owned[i]
	})

	for _, src := range owned {
		if src.Troops < 2 {
			continue
		}

		move, found := gen.findEnemyNeighbour(
			tb, gCtx, snap, src,
		)
		if found {
			return move, true
		}
	}

	return attack.Move{}, false
}

func (gen *Generator) findEnemyNeighbour(
	tb testing.TB,
	gCtx ctx.GameContext,
	snap *GameSnapshot,
	src sqlc.GetRegionsByGameRow,
) (attack.Move, bool) {
	tb.Helper()

	for _, region := range snap.Regions {
		if region.UserID == gCtx.UserID() {
			continue
		}

		neighbours, err := gen.harness.BoardService.AreNeighbours(
			gCtx, src.ExternalReference, region.ExternalReference,
		)
		if err != nil {
			tb.Fatalf("failed to check neighbours: %v", err)
		}

		if !neighbours {
			continue
		}

		attackingTroops := min(3, src.Troops-1)

		return attack.Move{
			AttackingRegionID: src.ExternalReference,
			DefendingRegionID: region.ExternalReference,
			TroopsInSource:    src.Troops,
			TroopsInTarget:    region.Troops,
			AttackingTroops:   attackingTroops,
		}, true
	}

	return attack.Move{}, false
}

// ConquerMove reads the conquer state and moves minimum troops.
func (gen *Generator) ConquerMove(
	tb testing.TB,
	gCtx ctx.GameContext,
) conquer.Move {
	tb.Helper()

	phaseState, err := gen.harness.ConquerService.GetPhaseState(gCtx)
	if err != nil {
		tb.Fatalf("failed to get conquer phase state: %v", err)
	}

	return conquer.Move{
		Troops: phaseState.MinimumTroops,
	}
}

// cardTypeValue maps card types to their numeric combination values.
var cardTypeValue = map[sqlc.GameCardType]int64{
	sqlc.GameCardTypeARTILLERY: 1,
	sqlc.GameCardTypeINFANTRY:  10,
	sqlc.GameCardTypeCAVALRY:   100,
	sqlc.GameCardTypeJOLLY:     1000,
}

// validCombinationValues are sums that form valid 3-card combos.
var validCombinationValues = map[int64]bool{
	3:    true, // 3 artillery
	30:   true, // 3 infantry
	300:  true, // 3 cavalry
	111:  true, // one of each
	1002: true, // jolly + 2 artillery
	1020: true, // jolly + 2 infantry
	1200: true, // jolly + 2 cavalry
}

func findValidCombination(
	tb testing.TB,
	playerCards []sqlc.GetCardsForPlayerRow,
) []int64 {
	tb.Helper()

	numCards := len(playerCards)

	for first := range numCards - 2 {
		for second := first + 1; second < numCards-1; second++ {
			for third := second + 1; third < numCards; third++ {
				val := cardTypeValue[playerCards[first].CardType] +
					cardTypeValue[playerCards[second].CardType] +
					cardTypeValue[playerCards[third].CardType]
				if validCombinationValues[val] {
					return []int64{
						playerCards[first].ID,
						playerCards[second].ID,
						playerCards[third].ID,
					}
				}
			}
		}
	}

	tb.Fatalf(
		"no valid card combination found among %d cards",
		numCards,
	)

	return nil
}
