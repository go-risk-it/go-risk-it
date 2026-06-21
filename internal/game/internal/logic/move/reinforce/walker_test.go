package reinforce_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/cards"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		turn             int64
		players          []snapshot.PlayerState
		privateSnapshots map[string]*snapshot.PlayerPrivate
		expectedPhase    sqlc.GamePhaseType
	}{
		{
			name: "next player has valid combination returns CARDS",
			turn: 0, // current player is index 0; next is index 1
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p1": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardInfantry},
						{ID: 2, Type: snapshot.CardInfantry},
						{ID: 3, Type: snapshot.CardInfantry},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeCARDS,
		},
		{
			name: "next player has no valid combination returns DEPLOY",
			turn: 0,
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p1": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardInfantry},
						{ID: 2, Type: snapshot.CardCavalry},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeDEPLOY,
		},
		{
			name: "next player has fewer than 3 cards returns DEPLOY",
			turn: 0,
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p1": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardInfantry},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeDEPLOY,
		},
		{
			name: "skips dead player to find next alive player",
			turn: 0, // current is index 0, index 1 is dead, index 2 is next alive
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerDead},
				{UserID: "p2", Index: 2, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p2": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardArtillery},
						{ID: 2, Type: snapshot.CardArtillery},
						{ID: 3, Type: snapshot.CardArtillery},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeCARDS,
		},
		{
			name: "wraps around player list",
			turn: 2, // current is index 2 (last), next alive wraps to index 0
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
				{UserID: "p2", Index: 2, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p0": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardInfantry},
						{ID: 2, Type: snapshot.CardCavalry},
						{ID: 3, Type: snapshot.CardArtillery},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeCARDS,
		},
		{
			name: "jolly combination is valid",
			turn: 0,
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p1": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardJolly},
						{ID: 2, Type: snapshot.CardCavalry},
						{ID: 3, Type: snapshot.CardCavalry},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeCARDS,
		},
		{
			name: "mixed non-matching types with no valid combination returns DEPLOY",
			turn: 0,
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p1": {
					Cards: []snapshot.CardState{
						{ID: 1, Type: snapshot.CardInfantry},
						{ID: 2, Type: snapshot.CardInfantry},
						{ID: 3, Type: snapshot.CardCavalry},
					},
				},
			},
			expectedPhase: sqlc.GamePhaseTypeDEPLOY,
		},
		{
			name: "next player has no cards returns DEPLOY",
			turn: 0,
			players: []snapshot.PlayerState{
				{UserID: "p0", Index: 0, Status: snapshot.PlayerAlive},
				{UserID: "p1", Index: 1, Status: snapshot.PlayerAlive},
			},
			privateSnapshots: map[string]*snapshot.PlayerPrivate{
				"p1": {
					Cards: nil,
				},
			},
			expectedPhase: sqlc.GamePhaseTypeDEPLOY,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Construct a real service (deps unused by Walk).
			svc, _ := reinforce.NewService(
				board.NewService(t),
				cards.NewService(t),
			)

			wctx := moveservice.WalkContext{
				PrevSnapshot: &snapshot.GameSnapshot{
					Game:    snapshot.GameMeta{Turn: testCase.turn},
					Players: testCase.players,
				},
				PrivateSnapshots: testCase.privateSnapshots,
			}

			got, err := svc.Walk(wctx)

			require.NoError(t, err)
			require.Equal(t, testCase.expectedPhase, got)
		})
	}
}
