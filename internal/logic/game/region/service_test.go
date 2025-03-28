package region_test

import (
	"context"
	"errors"
	"testing"

	ctx2 "github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	assignment2 "github.com/go-risk-it/go-risk-it/internal/logic/game/region/assignment"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region/assignment"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func TestServiceImpl_CreateRegions(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the state under test
	service := region.NewService(querier, assignmentService)

	// Set up test data
	ctx := ctx2.WithUserID(
		ctx2.WithSpan(ctx2.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"francesco",
	)

	players := []sqlc.GamePlayer{
		{ID: 1, GameID: 1, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: 1, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: 1, UserID: "giovanni", TurnIndex: 2},
	}
	regions := []string{
		"alaska",
		"northwest_territory",
		"greenland",
		"alberta",
		"ontario",
	}

	// Set up expectations for AssignRegionsToPlayers method
	assignmentService.
		EXPECT().
		AssignRegionsToPlayers(players, regions).
		Return(assignment2.RegionAssignment{
			regions[0]: players[0],
			regions[1]: players[1],
			regions[2]: players[2],
			regions[3]: players[0],
			regions[4]: players[1],
		})

	// Set up expectations for InsertRegions method
	querier.On("InsertRegions", ctx, []sqlc.InsertRegionsParams{
		{ExternalReference: "alaska", PlayerID: 1, Troops: 3},
		{ExternalReference: "northwest_territory", PlayerID: 2, Troops: 3},
		{ExternalReference: "greenland", PlayerID: 3, Troops: 3},
		{ExternalReference: "alberta", PlayerID: 1, Troops: 3},
		{ExternalReference: "ontario", PlayerID: 2, Troops: 3},
	}).Return(int64(5), nil)

	// Call the method under test
	err := service.CreateRegionsQ(ctx, querier, players, regions)

	// Assert the result
	require.NoError(t, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreateRegions_NoPlayers(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the state under test
	service := region.NewService(querier, assignmentService)

	// Set up test data
	ctx := ctx2.WithUserID(
		ctx2.WithSpan(ctx2.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"francesco",
	)

	var (
		players []sqlc.GamePlayer
		regions []string
	)

	// Call the method under test
	err := service.CreateRegionsQ(ctx, querier, players, regions)

	// Assert the result
	require.Error(t, err)
	require.Equal(t, region.ErrNoPlayers, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreateRegions_PlayersNotInSameGame(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the state under test
	service := region.NewService(querier, assignmentService)

	// Set up test data
	ctx := ctx2.WithUserID(
		ctx2.WithSpan(ctx2.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"francesco",
	)
	players := []sqlc.GamePlayer{
		{ID: 1, GameID: 1, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: 2, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: 1, UserID: "giovanni", TurnIndex: 2},
	}

	var regions []string

	// Call the method under test
	err := service.CreateRegionsQ(ctx, querier, players, regions)

	// Assert the result
	require.Error(t, err)
	require.Equal(t, region.ErrPlayersFromDifferentGames, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreateRegions_InsertRegionsError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the state under test
	service := region.NewService(querier, assignmentService)

	// Set up test data
	ctx := ctx2.WithUserID(
		ctx2.WithSpan(ctx2.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"francesco",
	)
	players := []sqlc.GamePlayer{
		{ID: 1, GameID: 1, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: 1, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: 1, UserID: "giovanni", TurnIndex: 2},
	}
	regions := []string{
		"alaska",
		"northwest_territory",
		"greenland",
		"alberta",
		"ontario",
	}

	// Set up expectations for AssignRegionsToPlayers method
	assignmentService.On("AssignRegionsToPlayers", players, regions).
		Return(assignment2.RegionAssignment{
			regions[0]: players[0],
			regions[1]: players[1],
			regions[2]: players[2],
			regions[3]: players[0],
			regions[4]: players[1],
		})

	// Set up expectations for InsertRegions method
	querier.On("InsertRegions", ctx, []sqlc.InsertRegionsParams{
		{ExternalReference: "alaska", PlayerID: 1, Troops: 3},
		{ExternalReference: "northwest_territory", PlayerID: 2, Troops: 3},
		{ExternalReference: "greenland", PlayerID: 3, Troops: 3},
		{ExternalReference: "alberta", PlayerID: 1, Troops: 3},
		{ExternalReference: "ontario", PlayerID: 2, Troops: 3},
	}).Return(int64(0), errors.New("insert regions error"))

	// Call the method under test
	err := service.CreateRegionsQ(ctx, querier, players, regions)

	// Assert the result
	require.Error(t, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}
