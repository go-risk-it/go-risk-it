package board_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func gameContext(userID string) gamectx.GameContext {
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return gamectx.WithGameID(userContext, 1)
}

// buildGraph creates a simple graph: A -- B -- C (linear chain).
func buildGraph() board.Graph {
	dto := &board.BoardDto{
		Regions: []board.RegionDto{
			{ExternalReference: "A"},
			{ExternalReference: "B"},
			{ExternalReference: "C"},
		},
		Borders: []board.BorderDto{
			{Source: "A", Target: "B"},
			{Source: "B", Target: "C"},
		},
	}

	g, err := board.NewGraph(dto)
	if err != nil {
		panic(err)
	}

	return g
}

func TestCanPlayerReachWithRegions_Reachable(t *testing.T) {
	t.Parallel()

	ctx := gameContext("player1")
	svc := board.NewServiceWithGraph(nil, buildGraph())

	// Player1 owns A, B, C — A can reach C through B.
	regions := []snapshot.RegionState{
		{ID: "A", OwnerID: "player1", Troops: 3},
		{ID: "B", OwnerID: "player1", Troops: 1},
		{ID: "C", OwnerID: "player1", Troops: 1},
	}

	reachable, err := svc.CanPlayerReachWithRegions(ctx, "A", "C", regions)
	require.NoError(t, err)
	require.True(t, reachable)
}

func TestCanPlayerReachWithRegions_NotReachable(t *testing.T) {
	t.Parallel()

	ctx := gameContext("player1")
	svc := board.NewServiceWithGraph(nil, buildGraph())

	// Player1 owns A and C, but B is owned by player2 — A cannot reach C.
	regions := []snapshot.RegionState{
		{ID: "A", OwnerID: "player1", Troops: 3},
		{ID: "B", OwnerID: "player2", Troops: 1},
		{ID: "C", OwnerID: "player1", Troops: 1},
	}

	reachable, err := svc.CanPlayerReachWithRegions(ctx, "A", "C", regions)
	require.NoError(t, err)
	require.False(t, reachable)
}

func TestCanPlayerReachWithRegions_SourceNotFound(t *testing.T) {
	t.Parallel()

	ctx := gameContext("player1")
	svc := board.NewServiceWithGraph(nil, buildGraph())

	// Source "X" does not exist in the regions list.
	regions := []snapshot.RegionState{
		{ID: "A", OwnerID: "player1", Troops: 3},
		{ID: "B", OwnerID: "player1", Troops: 1},
	}

	reachable, err := svc.CanPlayerReachWithRegions(ctx, "X", "A", regions)
	require.Error(t, err)
	require.False(t, reachable)

	var domainErr *domainerrors.DomainError

	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, domainerrors.CategoryNotFound, domainErr.Category())
}
