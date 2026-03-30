package routes_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/routes"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	mockplayer "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/player"
	mockstate "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidateGameWSConnection_Success(t *testing.T) {
	t.Parallel()

	stateSvc := mockstate.NewService(t)
	playerSvc := mockplayer.NewService(t)

	stateSvc.EXPECT().
		GetGameState(mock.Anything).
		Return(&state.Game{}, nil)

	playerSvc.EXPECT().
		GetPlayersState(mock.Anything).
		Return([]sqlc.GetPlayersStateRow{{UserID: "user-123"}}, nil)

	gameCtx := testGameContext()

	err := routes.ValidateGameWSConnection(gameCtx, stateSvc, playerSvc)
	assert.NoError(t, err)
}

func TestValidateGameWSConnection_UserNotParticipating(t *testing.T) {
	t.Parallel()

	stateSvc := mockstate.NewService(t)
	playerSvc := mockplayer.NewService(t)

	stateSvc.EXPECT().
		GetGameState(mock.Anything).
		Return(&state.Game{}, nil)

	playerSvc.EXPECT().
		GetPlayersState(mock.Anything).
		Return([]sqlc.GetPlayersStateRow{{UserID: "other-user"}}, nil)

	gameCtx := testGameContext()

	err := routes.ValidateGameWSConnection(gameCtx, stateSvc, playerSvc)
	require.Error(t, err)

	var domainErr *domainerrors.DomainError

	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domainerrors.CategoryForbidden, domainErr.Category())
}

func TestValidateGameWSConnection_GetGameStateError(t *testing.T) {
	t.Parallel()

	stateSvc := mockstate.NewService(t)
	playerSvc := mockplayer.NewService(t)

	stateSvc.EXPECT().
		GetGameState(mock.Anything).
		Return(nil, assert.AnError)

	gameCtx := testGameContext()

	err := routes.ValidateGameWSConnection(gameCtx, stateSvc, playerSvc)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestValidateGameWSConnection_GetPlayersStateError(t *testing.T) {
	t.Parallel()

	stateSvc := mockstate.NewService(t)
	playerSvc := mockplayer.NewService(t)

	stateSvc.EXPECT().
		GetGameState(mock.Anything).
		Return(&state.Game{}, nil)

	playerSvc.EXPECT().
		GetPlayersState(mock.Anything).
		Return(nil, assert.AnError)

	gameCtx := testGameContext()

	err := routes.ValidateGameWSConnection(gameCtx, stateSvc, playerSvc)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
