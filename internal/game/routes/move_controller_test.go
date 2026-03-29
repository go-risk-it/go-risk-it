package routes //nolint:testpackage // testing unexported mapper functions

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/reinforce"
	"github.com/stretchr/testify/assert"
)

func TestMapDeployMove(t *testing.T) {
	t.Parallel()

	req := request.DeployMove{
		RegionID:      "region-1",
		CurrentTroops: 3,
		DesiredTroops: 5,
	}

	got := mapDeployMove(req)
	assert.Equal(t, deploy.Move{
		RegionID:      "region-1",
		CurrentTroops: 3,
		DesiredTroops: 5,
	}, got)
}

func TestMapAttackMove(t *testing.T) {
	t.Parallel()

	req := request.AttackMove{
		SourceRegionID:  "attacker",
		TargetRegionID:  "defender",
		TroopsInSource:  10,
		TroopsInTarget:  5,
		AttackingTroops: 3,
	}

	got := mapAttackMove(req)
	assert.Equal(t, attack.Move{
		AttackingRegionID: "attacker",
		DefendingRegionID: "defender",
		TroopsInSource:    10,
		TroopsInTarget:    5,
		AttackingTroops:   3,
	}, got)
}

func TestMapConquerMove(t *testing.T) {
	t.Parallel()

	req := request.ConquerMove{Troops: 4}

	got := mapConquerMove(req)
	assert.Equal(t, conquer.Move{Troops: 4}, got)
}

func TestMapReinforceMove(t *testing.T) {
	t.Parallel()

	req := request.ReinforceMove{
		SourceRegionID: "src",
		TargetRegionID: "dst",
		TroopsInSource: 8,
		TroopsInTarget: 2,
		MovingTroops:   3,
	}

	got := mapReinforceMove(req)
	assert.Equal(t, reinforce.Move{
		SourceRegionID: "src",
		TargetRegionID: "dst",
		TroopsInSource: 8,
		TroopsInTarget: 2,
		MovingTroops:   3,
	}, got)
}

func TestMapCardsMove(t *testing.T) {
	t.Parallel()

	req := request.CardsMove{
		Combinations: []request.CardCombination{
			{CardIDs: []int64{1, 2, 3}},
			{CardIDs: []int64{4, 5, 6}},
		},
	}

	got := mapCardsMove(req)
	assert.Equal(t, cards.Move{
		Combinations: []cards.CardCombination{
			{CardIDs: []int64{1, 2, 3}},
			{CardIDs: []int64{4, 5, 6}},
		},
	}, got)
}

func TestMapCardsMove_Empty(t *testing.T) {
	t.Parallel()

	req := request.CardsMove{
		Combinations: []request.CardCombination{},
	}

	got := mapCardsMove(req)
	assert.Empty(t, got.Combinations)
}
