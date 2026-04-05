package player

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

// FindCardCombo finds a valid 3-card combination from the given hand.
// Returns nil if no valid combination exists.
//
//nolint:cyclop // card combination search across 3 strategies
func FindCardCombo(cards []snapshot.CardState) []int64 {
	byType := make(map[snapshot.CardType][]int64)
	var jollyIDs []int64

	for _, c := range cards {
		if c.Type == snapshot.CardJolly {
			jollyIDs = append(jollyIDs, c.ID)
		} else {
			byType[c.Type] = append(byType[c.Type], c.ID)
		}
	}

	// Try 3-of-a-kind.
	for _, ids := range byType {
		if len(ids) >= 3 {
			return ids[:3]
		}
	}

	// Try one-of-each (cavalry + infantry + artillery).
	types := []snapshot.CardType{snapshot.CardCavalry, snapshot.CardInfantry, snapshot.CardArtillery}
	var oneOfEach []int64

	for _, t := range types {
		if ids, ok := byType[t]; ok && len(ids) > 0 {
			oneOfEach = append(oneOfEach, ids[0])
		}
	}

	if len(oneOfEach) == 3 {
		return oneOfEach
	}

	// Try 2-of-a-kind + jolly (the only valid jolly combo).
	if len(jollyIDs) > 0 {
		for _, ids := range byType {
			if len(ids) >= 2 {
				return []int64{ids[0], ids[1], jollyIDs[0]}
			}
		}
	}

	return nil
}

// NewAdvanceAction returns an action that advances past the given phase.
func NewAdvanceAction(phase snapshot.PhaseType) *Action {
	return &Action{
		Type: ActionAdvance,
		Advance: &AdvanceAction{
			CurrentPhase: string(phase),
		},
	}
}
