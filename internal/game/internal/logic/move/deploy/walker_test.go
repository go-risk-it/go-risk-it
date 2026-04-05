package deploy_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/deploy"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/region"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		updatedPhase  snapshot.PhaseState
		expectedPhase sqlc.GamePhaseType
		expectErr     bool
	}{
		{
			name:          "deployable troops > 0 returns DEPLOY",
			updatedPhase:  snapshot.DeployPhaseState{DeployableTroops: 5},
			expectedPhase: sqlc.GamePhaseTypeDEPLOY,
		},
		{
			name:          "deployable troops == 0 returns ATTACK",
			updatedPhase:  snapshot.DeployPhaseState{DeployableTroops: 0},
			expectedPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name:          "deployable troops < 0 returns ATTACK",
			updatedPhase:  snapshot.DeployPhaseState{DeployableTroops: -1},
			expectedPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name:         "wrong phase state type returns error",
			updatedPhase: snapshot.EmptyPhaseState{},
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Construct a real service (deps unused by Walk).
			svc, _ := deploy.NewService(
				db.NewQuerier(t),
				phase.NewService(t),
				region.NewService(t),
			)

			wctx := moveservice.WalkContext{
				Effect: moveservice.MoveEffect{
					UpdatedPhase: tt.updatedPhase,
				},
			}

			got, err := svc.Walk(wctx)

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedPhase, got)
		})
	}
}
