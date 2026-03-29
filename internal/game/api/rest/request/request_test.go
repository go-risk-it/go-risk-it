package request_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployMove_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		move    request.DeployMove
		wantErr string
	}{
		{name: "valid", move: request.DeployMove{RegionID: "r1", DesiredTroops: 3}},
		{
			name:    "empty region",
			move:    request.DeployMove{RegionID: "", DesiredTroops: 3},
			wantErr: "regionId",
		},
		{
			name:    "zero troops",
			move:    request.DeployMove{RegionID: "r1", DesiredTroops: 0},
			wantErr: "desiredTroops",
		},
		{
			name:    "negative troops",
			move:    request.DeployMove{RegionID: "r1", DesiredTroops: -1},
			wantErr: "desiredTroops",
		},
		{
			name:    "troops exceeds maximum",
			move:    request.DeployMove{RegionID: "r1", DesiredTroops: 10001},
			wantErr: "exceeds maximum",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.move.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}

func TestAttackMove_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		move    request.AttackMove
		wantErr string
	}{
		{
			name: "valid",
			move: request.AttackMove{
				SourceRegionID:  "r1",
				TargetRegionID:  "r2",
				AttackingTroops: 2,
			},
		},
		{
			name: "empty source",
			move: request.AttackMove{
				SourceRegionID:  "",
				TargetRegionID:  "r2",
				AttackingTroops: 2,
			},
			wantErr: "sourceRegionId",
		},
		{
			name: "empty target",
			move: request.AttackMove{
				SourceRegionID:  "r1",
				TargetRegionID:  "",
				AttackingTroops: 2,
			},
			wantErr: "targetRegionId",
		},
		{
			name: "same regions",
			move: request.AttackMove{
				SourceRegionID:  "r1",
				TargetRegionID:  "r1",
				AttackingTroops: 2,
			},
			wantErr: "must be different",
		},
		{
			name: "zero troops",
			move: request.AttackMove{
				SourceRegionID:  "r1",
				TargetRegionID:  "r2",
				AttackingTroops: 0,
			},
			wantErr: "attackingTroops",
		},
		{
			name: "troops exceeds maximum",
			move: request.AttackMove{
				SourceRegionID:  "r1",
				TargetRegionID:  "r2",
				AttackingTroops: 10001,
			},
			wantErr: "exceeds maximum",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.move.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}

func TestConquerMove_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		move    request.ConquerMove
		wantErr string
	}{
		{name: "valid", move: request.ConquerMove{Troops: 1}},
		{name: "zero", move: request.ConquerMove{Troops: 0}, wantErr: "troops"},
		{name: "negative", move: request.ConquerMove{Troops: -1}, wantErr: "troops"},
		{
			name:    "exceeds max",
			move:    request.ConquerMove{Troops: 10001},
			wantErr: "exceeds maximum",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.move.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}

func TestReinforceMove_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		move    request.ReinforceMove
		wantErr string
	}{
		{
			name: "valid",
			move: request.ReinforceMove{
				SourceRegionID: "r1",
				TargetRegionID: "r2",
				MovingTroops:   1,
			},
		},
		{
			name: "same regions",
			move: request.ReinforceMove{
				SourceRegionID: "r1",
				TargetRegionID: "r1",
				MovingTroops:   1,
			},
			wantErr: "must be different",
		},
		{
			name: "zero troops",
			move: request.ReinforceMove{
				SourceRegionID: "r1",
				TargetRegionID: "r2",
				MovingTroops:   0,
			},
			wantErr: "movingTroops",
		},
		{
			name: "troops exceeds maximum",
			move: request.ReinforceMove{
				SourceRegionID: "r1",
				TargetRegionID: "r2",
				MovingTroops:   10001,
			},
			wantErr: "exceeds maximum",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.move.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}

func TestCardsMove_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		move    request.CardsMove
		wantErr string
	}{
		{
			name: "valid",
			move: request.CardsMove{
				Combinations: []request.CardCombination{{CardIDs: []int64{1, 2, 3}}},
			},
		},
		{
			name:    "empty combinations",
			move:    request.CardsMove{Combinations: []request.CardCombination{}},
			wantErr: "combinations must not be empty",
		},
		{
			name: "empty card ids",
			move: request.CardsMove{
				Combinations: []request.CardCombination{{CardIDs: []int64{}}},
			},
			wantErr: "cardIds must not be empty",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.move.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}

func TestCreateGame_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     request.CreateGame
		wantErr string
	}{
		{
			name: "valid",
			req:  request.CreateGame{Players: []request.Player{{UserID: "u1", Name: "p1"}}},
		},
		{
			name:    "empty players",
			req:     request.CreateGame{Players: []request.Player{}},
			wantErr: "players must not be empty",
		},
		{
			name:    "empty user id",
			req:     request.CreateGame{Players: []request.Player{{UserID: "", Name: "p1"}}},
			wantErr: "userId must not be empty",
		},
		{
			name:    "empty name",
			req:     request.CreateGame{Players: []request.Player{{UserID: "u1", Name: ""}}},
			wantErr: "name must not be empty",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.req.Validate()
			if testCase.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.wantErr)
			}
		})
	}
}
