package game

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/player"
	"github.com/tomfran/go-risk-it/internal/game/region"
	"go.uber.org/zap"
	"testing"
)

// creates a game with a valid board and list of users
func TestCreateGameWithValidBoardAndUsers(t *testing.T) {
	var gameId int64 = 1
	users := []string{"Giovanni", "Gabriele"}
	ctx := context.Background()
	mockQuerier := db.NewMockQuerier(t)

	players := []db.Player{
		{420, gameId, "Giovanni"},
		{69, gameId, "Gabriele"},
	}

	regions := []board.Region{
		{1, "Netherlands", 1},
		{2, "Italy", 1},
		{3, "Tasin", 2},
		{4, "Samon", 3},
	}

	gameBoard := &board.Board{
		Regions:    regions,
		Continents: []board.Continent{},
		Borders:    []board.Border{},
	}

	// setup mocks
	mockQuerier.EXPECT().InsertGame(ctx).Return(1, nil)

	playerServiceMock := player.NewMockService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(ctx, mockQuerier, gameId, users).
		Return(players, nil)

	regionServiceMock := region.NewMockService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(ctx, mockQuerier, players, regions).
		Return(nil)

	// Initialize the service
	service := &ServiceImpl{
		log:           zap.NewExample().Sugar(),
		playerService: playerServiceMock,
		regionService: regionServiceMock,
	}

	result := service.CreateGame(ctx, mockQuerier, gameBoard, users)

	assert.NoError(t, result)
}
