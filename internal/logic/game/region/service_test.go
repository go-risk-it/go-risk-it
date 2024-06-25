package region_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	assignment2 "github.com/go-risk-it/go-risk-it/internal/logic/game/region/assignment"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region/assignment"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceImpl_CreateRegions(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewNop().Sugar()
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the gamestate under test
	service := region.NewService(logger, querier, assignmentService)

	// Set up test data
	ctx := context.Background()
	players := []sqlc.Player{
		{ID: 1, GameID: 1, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: 1, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: 1, UserID: "giovanni", TurnIndex: 2},
	}
	regions := []board.Region{
		{ExternalReference: "alaska", Name: "Alaska"},
		{ExternalReference: "northwest_territory", Name: "Northwest Territory"},
		{ExternalReference: "greenland", Name: "Greenland"},
		{ExternalReference: "alberta", Name: "Alberta"},
		{ExternalReference: "ontario", Name: "Ontario"},
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
	}).Return(int64(5), nil)

	// Call the method under test
	err := service.CreateRegions(ctx, querier, players, regions)

	// Assert the result
	require.NoError(t, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreateRegions_NoPlayers(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewNop().Sugar()
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the gamestate under test
	service := region.NewService(logger, querier, assignmentService)

	// Set up test data
	ctx := context.Background()

	var (
		players []sqlc.Player
		regions []board.Region
	)

	// Call the method under test
	err := service.CreateRegions(ctx, querier, players, regions)

	// Assert the result
	require.Error(t, err)
	require.Equal(t, region.ErrNoPlayers, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreateRegions_PlayersNotInSameGame(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewNop().Sugar()
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the gamestate under test
	service := region.NewService(logger, querier, assignmentService)

	// Set up test data
	ctx := context.Background()
	players := []sqlc.Player{
		{ID: 1, GameID: 1, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: 2, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: 1, UserID: "giovanni", TurnIndex: 2},
	}

	var regions []board.Region

	// Call the method under test
	err := service.CreateRegions(ctx, querier, players, regions)

	// Assert the result
	require.Error(t, err)
	require.Equal(t, region.ErrPlayersFromDifferentGames, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreateRegions_InsertRegionsError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewNop().Sugar()
	querier := db.NewQuerier(t)
	assignmentService := assignment.NewService(t)

	// Initialize the gamestate under test
	service := region.NewService(logger, querier, assignmentService)

	// Set up test data
	ctx := context.Background()
	players := []sqlc.Player{
		{ID: 1, GameID: 1, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: 1, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: 1, UserID: "giovanni", TurnIndex: 2},
	}
	regions := []board.Region{
		{ExternalReference: "alaska", Name: "Alaska"},
		{ExternalReference: "northwest_territory", Name: "Northwest Territory"},
		{ExternalReference: "greenland", Name: "Greenland"},
		{ExternalReference: "alberta", Name: "Alberta"},
		{ExternalReference: "ontario", Name: "Ontario"},
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
	}).Return(int64(0), fmt.Errorf("insert regions error"))

	// Call the method under test
	err := service.CreateRegions(ctx, querier, players, regions)

	// Assert the result
	require.Error(t, err)
	assignmentService.AssertExpectations(t)
	querier.AssertExpectations(t)
}
